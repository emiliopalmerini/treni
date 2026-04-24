// Package state provides persistent storage for user preferences
// (currently: per-chat favorite routes) backed by a JSON file.
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/emiliopalmerini/treni/internal/domain"
)

// ErrFavoriteLimit is returned by Save when adding a new favorite
// would exceed the per-chat cap.
var ErrFavoriteLimit = errors.New("favorite limit reached")

// MaxFavoritesPerChat is the hard cap enforced by Save. ADR-013 §Limits.
const MaxFavoritesPerChat = 10

// JSONStore is a mutex-guarded, file-backed FavoritesStore.
// Writes are atomic (write-tmp + rename) so a crash mid-write cannot
// leave the file half-written.
type JSONStore struct {
	path string

	mu    sync.Mutex
	state fileState
}

type fileState struct {
	Favorites map[string]map[string]favoriteRecord `json:"favorites"`
}

type favoriteRecord struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// NewJSONStore loads state from `path`. A missing file yields an
// empty store; a corrupt file returns an error (better to fail fast
// than silently drop user data).
func NewJSONStore(path string) (*JSONStore, error) {
	s := &JSONStore{path: path, state: fileState{Favorites: map[string]map[string]favoriteRecord{}}}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		return s, nil
	}
	var fs fileState
	if err := json.Unmarshal(data, &fs); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if fs.Favorites == nil {
		fs.Favorites = map[string]map[string]favoriteRecord{}
	}
	s.state = fs
	return s, nil
}

func (s *JSONStore) List(chatID int64) []domain.Favorite {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.state.Favorites[chatIDKey(chatID)]
	out := make([]domain.Favorite, 0, len(m))
	for name, rec := range m {
		out = append(out, domain.Favorite{Name: name, From: rec.From, To: rec.To})
	}
	return out
}

func (s *JSONStore) Get(chatID int64, name string) (domain.Favorite, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.state.Favorites[chatIDKey(chatID)]
	rec, ok := m[strings.ToLower(name)]
	if !ok {
		return domain.Favorite{}, false
	}
	return domain.Favorite{Name: strings.ToLower(name), From: rec.From, To: rec.To}, true
}

func (s *JSONStore) Save(chatID int64, fav domain.Favorite) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := chatIDKey(chatID)
	m, ok := s.state.Favorites[key]
	if !ok {
		m = make(map[string]favoriteRecord)
		s.state.Favorites[key] = m
	}
	name := strings.ToLower(fav.Name)
	if _, exists := m[name]; !exists && len(m) >= MaxFavoritesPerChat {
		return ErrFavoriteLimit
	}
	m[name] = favoriteRecord{From: fav.From, To: fav.To}
	return s.flushLocked()
}

func (s *JSONStore) Delete(chatID int64, name string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m := s.state.Favorites[chatIDKey(chatID)]
	key := strings.ToLower(name)
	if _, ok := m[key]; !ok {
		return false, nil
	}
	delete(m, key)
	if err := s.flushLocked(); err != nil {
		return false, err
	}
	return true, nil
}

// flushLocked writes the current state atomically. Caller holds s.mu.
func (s *JSONStore) flushLocked() error {
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	dir := filepath.Dir(s.path)
	if dir == "" {
		dir = "."
	}
	tmp, err := os.CreateTemp(dir, ".state-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

func chatIDKey(id int64) string { return fmt.Sprintf("%d", id) }
