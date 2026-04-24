package state_test

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"

	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/emiliopalmerini/treni/internal/state"
)

func tempStorePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "state.json")
}

func TestJSONStore_missingFileLoadsEmpty(t *testing.T) {
	path := tempStorePath(t)
	s, err := state.NewJSONStore(path)
	if err != nil {
		t.Fatalf("NewJSONStore(missing): %v", err)
	}
	if got := s.List(42); len(got) != 0 {
		t.Errorf("List on empty store = %v, want empty", got)
	}
}

func TestJSONStore_saveListGetDeleteRoundTrip(t *testing.T) {
	path := tempStorePath(t)
	s, err := state.NewJSONStore(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Save(42, domain.Favorite{Name: "home", From: "Desio", To: "Milano"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := s.Save(42, domain.Favorite{Name: "office", From: "Milano Centrale", To: "Brescia"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, ok := s.Get(42, "home")
	if !ok {
		t.Fatal("Get(home): not found")
	}
	if got.From != "Desio" || got.To != "Milano" {
		t.Errorf("Get home = %+v", got)
	}

	list := s.List(42)
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	if len(list) != 2 || list[0].Name != "home" || list[1].Name != "office" {
		t.Fatalf("List = %+v", list)
	}

	deleted, err := s.Delete(42, "home")
	if err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Error("Delete(home) = false, want true")
	}
	if _, ok := s.Get(42, "home"); ok {
		t.Error("Get(home) still present after Delete")
	}
}

func TestJSONStore_saveIsCaseInsensitive(t *testing.T) {
	path := tempStorePath(t)
	s, _ := state.NewJSONStore(path)

	if err := s.Save(42, domain.Favorite{Name: "Home", From: "Desio", To: "Milano"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Get(42, "home"); !ok {
		t.Error("lookup should be case-insensitive: Home saved, home missing")
	}
	if _, ok := s.Get(42, "HOME"); !ok {
		t.Error("lookup should be case-insensitive: Home saved, HOME missing")
	}
}

func TestJSONStore_saveOverwriteSameName(t *testing.T) {
	path := tempStorePath(t)
	s, _ := state.NewJSONStore(path)

	_ = s.Save(42, domain.Favorite{Name: "home", From: "Desio", To: "Milano"})
	if err := s.Save(42, domain.Favorite{Name: "home", From: "Monza", To: "Lecco"}); err != nil {
		t.Fatalf("overwrite: %v", err)
	}

	got, _ := s.Get(42, "home")
	if got.From != "Monza" || got.To != "Lecco" {
		t.Errorf("overwrite failed: %+v", got)
	}
	if list := s.List(42); len(list) != 1 {
		t.Errorf("expected 1 favorite, got %d", len(list))
	}
}

func TestJSONStore_perChatIsolation(t *testing.T) {
	path := tempStorePath(t)
	s, _ := state.NewJSONStore(path)

	_ = s.Save(42, domain.Favorite{Name: "home", From: "Desio", To: "Milano"})
	_ = s.Save(99, domain.Favorite{Name: "home", From: "Roma", To: "Napoli"})

	got42, _ := s.Get(42, "home")
	got99, _ := s.Get(99, "home")
	if got42.From != "Desio" {
		t.Errorf("chat 42 home = %+v", got42)
	}
	if got99.From != "Roma" {
		t.Errorf("chat 99 home = %+v", got99)
	}
}

func TestJSONStore_capEnforced(t *testing.T) {
	path := tempStorePath(t)
	s, _ := state.NewJSONStore(path)

	for i := 0; i < 10; i++ {
		name := string(rune('a' + i))
		if err := s.Save(42, domain.Favorite{Name: name, From: "X", To: "Y"}); err != nil {
			t.Fatalf("Save %d: %v", i, err)
		}
	}
	err := s.Save(42, domain.Favorite{Name: "eleventh", From: "X", To: "Y"})
	if !errors.Is(err, state.ErrFavoriteLimit) {
		t.Errorf("11th Save err = %v, want ErrFavoriteLimit", err)
	}
}

func TestJSONStore_capDoesNotBlockOverwrite(t *testing.T) {
	path := tempStorePath(t)
	s, _ := state.NewJSONStore(path)

	for i := 0; i < 10; i++ {
		name := string(rune('a' + i))
		if err := s.Save(42, domain.Favorite{Name: name, From: "X", To: "Y"}); err != nil {
			t.Fatalf("Save %d: %v", i, err)
		}
	}
	// Overwriting an existing name must succeed even at cap.
	if err := s.Save(42, domain.Favorite{Name: "a", From: "NEW", To: "NEW"}); err != nil {
		t.Errorf("overwrite at cap should succeed, got %v", err)
	}
}

func TestJSONStore_deleteMissingReturnsFalse(t *testing.T) {
	path := tempStorePath(t)
	s, _ := state.NewJSONStore(path)

	deleted, err := s.Delete(42, "nope")
	if err != nil {
		t.Fatal(err)
	}
	if deleted {
		t.Error("Delete of missing = true, want false")
	}
}

func TestJSONStore_persistsAcrossInstances(t *testing.T) {
	path := tempStorePath(t)
	s1, _ := state.NewJSONStore(path)
	if err := s1.Save(42, domain.Favorite{Name: "home", From: "Desio", To: "Milano"}); err != nil {
		t.Fatal(err)
	}

	s2, err := state.NewJSONStore(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	got, ok := s2.Get(42, "home")
	if !ok {
		t.Fatal("persisted favorite not found after reload")
	}
	if got.From != "Desio" || got.To != "Milano" {
		t.Errorf("persisted = %+v", got)
	}
}

func TestJSONStore_corruptFileReturnsError(t *testing.T) {
	path := tempStorePath(t)
	if err := os.WriteFile(path, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := state.NewJSONStore(path); err == nil {
		t.Error("corrupt file should yield error, got nil")
	}
}

func TestJSONStore_writesAtomically(t *testing.T) {
	// After a successful Save, no stray `.tmp` file should linger.
	path := tempStorePath(t)
	s, _ := state.NewJSONStore(path)
	if err := s.Save(42, domain.Favorite{Name: "home", From: "Desio", To: "Milano"}); err != nil {
		t.Fatal(err)
	}

	dir := filepath.Dir(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("stray temp file: %s", e.Name())
		}
	}
}

func TestJSONStore_concurrentSavesSerialize(t *testing.T) {
	// Mutex + file rename should keep state consistent under goroutines.
	path := tempStorePath(t)
	s, _ := state.NewJSONStore(path)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := string(rune('a' + i))
			_ = s.Save(42, domain.Favorite{Name: name, From: "X", To: "Y"})
		}(i)
	}
	wg.Wait()

	if got := len(s.List(42)); got != 5 {
		t.Errorf("after concurrent saves: %d entries, want 5", got)
	}

	// Reload from disk and verify integrity.
	s2, err := state.NewJSONStore(path)
	if err != nil {
		t.Fatalf("reload after concurrent writes: %v", err)
	}
	if got := len(s2.List(42)); got != 5 {
		t.Errorf("after reload: %d entries, want 5", got)
	}
}
