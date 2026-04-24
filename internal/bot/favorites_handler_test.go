package bot_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/emiliopalmerini/treni/internal/bot"
	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/emiliopalmerini/treni/internal/state"
)

// fakeFavStore is an in-memory FavoritesStore for handler tests.
type fakeFavStore struct {
	mu   sync.Mutex
	byID map[int64]map[string]domain.Favorite
	cap  int
}

func newFakeFavStore() *fakeFavStore {
	return &fakeFavStore{byID: make(map[int64]map[string]domain.Favorite), cap: 10}
}

func (f *fakeFavStore) List(chatID int64) []domain.Favorite {
	f.mu.Lock()
	defer f.mu.Unlock()
	m := f.byID[chatID]
	out := make([]domain.Favorite, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func (f *fakeFavStore) Get(chatID int64, name string) (domain.Favorite, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fav, ok := f.byID[chatID][strings.ToLower(name)]
	return fav, ok
}

func (f *fakeFavStore) Save(chatID int64, fav domain.Favorite) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.byID[chatID]
	if !ok {
		m = make(map[string]domain.Favorite)
		f.byID[chatID] = m
	}
	key := strings.ToLower(fav.Name)
	if _, exists := m[key]; !exists && len(m) >= f.cap {
		return state.ErrFavoriteLimit
	}
	fav.Name = key
	m[key] = fav
	return nil
}

func (f *fakeFavStore) Delete(chatID int64, name string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m := f.byID[chatID]
	key := strings.ToLower(name)
	if _, ok := m[key]; !ok {
		return false, nil
	}
	delete(m, key)
	return true, nil
}

func TestSaveHandler_savesAndConfirms(t *testing.T) {
	favs := newFakeFavStore()
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.Register("/save", bot.NewSaveHandler(favs))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/save home Desio > Milano"))

	got, ok := favs.Get(42, "home")
	if !ok {
		t.Fatal("favorite not saved")
	}
	if got.From != "Desio" || got.To != "Milano" {
		t.Errorf("saved = %+v", got)
	}

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	lower := strings.ToLower(msgs[0].Text)
	if !strings.Contains(lower, "home") {
		t.Errorf("confirmation missing name: %q", msgs[0].Text)
	}
}

func TestSaveHandler_overwriteReplies_updated(t *testing.T) {
	favs := newFakeFavStore()
	_ = favs.Save(42, domain.Favorite{Name: "home", From: "Old", To: "Old"})
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.Register("/save", bot.NewSaveHandler(favs))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/save home Desio > Milano"))

	got, _ := favs.Get(42, "home")
	if got.From != "Desio" {
		t.Errorf("overwrite failed: %+v", got)
	}
	lower := strings.ToLower(sender.messages()[0].Text)
	if !strings.Contains(lower, "updated") {
		t.Errorf("overwrite should say 'updated', got %q", sender.messages()[0].Text)
	}
}

func TestSaveHandler_usageError(t *testing.T) {
	favs := newFakeFavStore()
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.Register("/save", bot.NewSaveHandler(favs))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/save"))

	if len(favs.List(42)) != 0 {
		t.Error("nothing should be saved on usage error")
	}
	lower := strings.ToLower(sender.messages()[0].Text)
	if !strings.Contains(lower, "usage") {
		t.Errorf("usage-error reply: %q", sender.messages()[0].Text)
	}
}

func TestSaveHandler_invalidName(t *testing.T) {
	favs := newFakeFavStore()
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.Register("/save", bot.NewSaveHandler(favs))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/save /home Desio > Milano"))

	if len(favs.List(42)) != 0 {
		t.Error("invalid name should not save")
	}
	lower := strings.ToLower(sender.messages()[0].Text)
	if !strings.Contains(lower, "invalid") {
		t.Errorf("invalid-name reply: %q", sender.messages()[0].Text)
	}
}

func TestSaveHandler_atCap(t *testing.T) {
	favs := newFakeFavStore()
	for i := 0; i < 10; i++ {
		name := string(rune('a' + i))
		_ = favs.Save(42, domain.Favorite{Name: name, From: "X", To: "Y"})
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.Register("/save", bot.NewSaveHandler(favs))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/save eleventh Desio > Milano"))

	if _, ok := favs.Get(42, "eleventh"); ok {
		t.Error("11th favorite should not be saved")
	}
	lower := strings.ToLower(sender.messages()[0].Text)
	if !strings.Contains(lower, "limit") {
		t.Errorf("cap reply should mention limit: %q", sender.messages()[0].Text)
	}
}

func TestUnsaveHandler_deletesAndConfirms(t *testing.T) {
	favs := newFakeFavStore()
	_ = favs.Save(42, domain.Favorite{Name: "home", From: "Desio", To: "Milano"})
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.Register("/unsave", bot.NewUnsaveHandler(favs))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/unsave home"))

	if _, ok := favs.Get(42, "home"); ok {
		t.Error("favorite still present after /unsave")
	}
	if len(sender.messages()) != 1 {
		t.Fatal("no confirmation")
	}
}

func TestUnsaveHandler_missingName(t *testing.T) {
	favs := newFakeFavStore()
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.Register("/unsave", bot.NewUnsaveHandler(favs))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/unsave nope"))

	lower := strings.ToLower(sender.messages()[0].Text)
	if !strings.Contains(lower, "no favorite") {
		t.Errorf("reply: %q", sender.messages()[0].Text)
	}
}

func TestUnsaveHandler_usage(t *testing.T) {
	favs := newFakeFavStore()
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.Register("/unsave", bot.NewUnsaveHandler(favs))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/unsave"))

	lower := strings.ToLower(sender.messages()[0].Text)
	if !strings.Contains(lower, "usage") {
		t.Errorf("reply: %q", sender.messages()[0].Text)
	}
}

func TestFavoritesHandler_emptyState(t *testing.T) {
	favs := newFakeFavStore()
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.Register("/favorites", bot.NewFavoritesHandler(favs))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/favorites"))

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if len(sender.sentKeyboards()) != 0 {
		t.Error("empty state should not send a keyboard")
	}
	lower := strings.ToLower(msgs[0].Text)
	if !strings.Contains(lower, "no favorites") {
		t.Errorf("empty reply: %q", msgs[0].Text)
	}
}

func TestFavoritesHandler_listsWithButtons(t *testing.T) {
	favs := newFakeFavStore()
	_ = favs.Save(42, domain.Favorite{Name: "home", From: "Desio", To: "Milano"})
	_ = favs.Save(42, domain.Favorite{Name: "office", From: "Milano Centrale", To: "Brescia"})
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.Register("/favorites", bot.NewFavoritesHandler(favs))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/favorites"))

	kbs := sender.sentKeyboards()
	if len(kbs) != 1 {
		t.Fatalf("got %d keyboards, want 1", len(kbs))
	}
	kb := kbs[0]
	// 2 favorites x 2 buttons each
	if len(kb.Buttons) != 4 {
		t.Fatalf("got %d buttons, want 4", len(kb.Buttons))
	}

	// Each favorite should appear in the body and have fr: + fd: callbacks.
	if !strings.Contains(kb.Text, "home") || !strings.Contains(kb.Text, "office") {
		t.Errorf("list missing favorite names: %q", kb.Text)
	}
	if !strings.Contains(kb.Text, "Desio") || !strings.Contains(kb.Text, "Milano") {
		t.Errorf("list missing route: %q", kb.Text)
	}

	var sawRun, sawDel bool
	for _, b := range kb.Buttons {
		if strings.HasPrefix(b.Data, "fr:") {
			sawRun = true
		}
		if strings.HasPrefix(b.Data, "fd:") {
			sawDel = true
		}
	}
	if !sawRun || !sawDel {
		t.Errorf("missing fr:/fd: callbacks; buttons: %+v", kb.Buttons)
	}
}

func TestFavoritesRunCallback_runsSavedRoute(t *testing.T) {
	favs := newFakeFavStore()
	_ = favs.Save(42, domain.Favorite{Name: "home", From: "Desio", To: "Milano"})

	query := &fakeQuerySvc{
		stations: []domain.Station{{Code: "S01234", Name: "Desio"}},
		departures: []domain.Departure{
			{
				TrainCategory: "S",
				ScheduledTime: time.Date(2026, 4, 25, 8, 32, 0, 0, time.UTC),
				Destination:   "Milano Porta Garibaldi",
				Platform:      "2",
			},
		},
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnCallback("fr:", bot.NewFavoritesRunCallback(favs, query, 60*time.Minute))

	_ = d.Handle(context.Background(), sender,
		newCallbackUpdate(42, 7, "cb-run", "fr:home"))

	if query.gotStation != "S01234" || query.gotTo != "Milano" {
		t.Errorf("query did not see resolved route: station=%q to=%q", query.gotStation, query.gotTo)
	}
	edits := sender.editedMessages()
	if len(edits) != 1 || !strings.Contains(edits[0].Text, "Milano Porta Garibaldi") {
		t.Errorf("expected edit with departures, got: %+v", edits)
	}
	if len(sender.answered()) != 1 {
		t.Error("callback spinner not dismissed")
	}
}

func TestFavoritesRunCallback_missingFavorite(t *testing.T) {
	favs := newFakeFavStore()
	query := &fakeQuerySvc{}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnCallback("fr:", bot.NewFavoritesRunCallback(favs, query, 60*time.Minute))

	_ = d.Handle(context.Background(), sender,
		newCallbackUpdate(42, 7, "cb-run", "fr:ghost"))

	if query.gotStation != "" {
		t.Error("missing favorite should not trigger query")
	}
	edits := sender.editedMessages()
	if len(edits) != 1 {
		t.Fatalf("expected one edit explaining gone, got %d", len(edits))
	}
	lower := strings.ToLower(edits[0].Text)
	if !strings.Contains(lower, "gone") && !strings.Contains(lower, "no favorite") {
		t.Errorf("missing-favorite text: %q", edits[0].Text)
	}
}

func TestFavoritesDeleteCallback_editsMessage(t *testing.T) {
	favs := newFakeFavStore()
	_ = favs.Save(42, domain.Favorite{Name: "home", From: "Desio", To: "Milano"})
	_ = favs.Save(42, domain.Favorite{Name: "office", From: "A", To: "B"})
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnCallback("fd:", bot.NewFavoritesDeleteCallback(favs))

	_ = d.Handle(context.Background(), sender,
		newCallbackUpdate(42, 7, "cb-del", "fd:home"))

	if _, ok := favs.Get(42, "home"); ok {
		t.Error("home still present after delete callback")
	}
	edits := sender.editedMessages()
	if len(edits) != 1 {
		t.Fatalf("got %d edits, want 1", len(edits))
	}
	if strings.Contains(edits[0].Text, "home") {
		t.Errorf("deleted row still in body: %q", edits[0].Text)
	}
	if !strings.Contains(edits[0].Text, "office") {
		t.Errorf("remaining row missing: %q", edits[0].Text)
	}
	if len(sender.answered()) != 1 {
		t.Error("callback spinner not dismissed")
	}
}

func TestFavoritesDeleteCallback_lastOneShowsEmptyState(t *testing.T) {
	favs := newFakeFavStore()
	_ = favs.Save(42, domain.Favorite{Name: "home", From: "Desio", To: "Milano"})
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnCallback("fd:", bot.NewFavoritesDeleteCallback(favs))

	_ = d.Handle(context.Background(), sender,
		newCallbackUpdate(42, 7, "cb-del", "fd:home"))

	edits := sender.editedMessages()
	if len(edits) != 1 {
		t.Fatalf("got %d edits, want 1", len(edits))
	}
	lower := strings.ToLower(edits[0].Text)
	if !strings.Contains(lower, "no favorites") {
		t.Errorf("last-deleted text: %q", edits[0].Text)
	}
	if len(edits[0].Buttons) != 0 {
		t.Errorf("empty state should drop keyboard; got %d buttons", len(edits[0].Buttons))
	}
}

func TestAliasHandler_hitsNicknameBeforeQuery(t *testing.T) {
	favs := newFakeFavStore()
	_ = favs.Save(42, domain.Favorite{Name: "home", From: "Desio", To: "Milano"})

	query := &fakeQuerySvc{
		stations: []domain.Station{{Code: "S01234", Name: "Desio"}},
		departures: []domain.Departure{
			{
				TrainCategory: "S",
				ScheduledTime: time.Date(2026, 4, 25, 8, 32, 0, 0, time.UTC),
				Destination:   "Milano Porta Garibaldi",
			},
		},
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewAliasHandler(favs, bot.NewQueryHandler(query, 60*time.Minute)))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "home"))

	if query.gotStation != "S01234" || query.gotTo != "Milano" {
		t.Errorf("alias did not expand to saved route: station=%q to=%q", query.gotStation, query.gotTo)
	}
	msgs := sender.messages()
	if len(msgs) != 1 || !strings.Contains(msgs[0].Text, "Milano Porta Garibaldi") {
		t.Errorf("unexpected reply: %+v", msgs)
	}
}

func TestAliasHandler_caseInsensitive(t *testing.T) {
	favs := newFakeFavStore()
	_ = favs.Save(42, domain.Favorite{Name: "home", From: "Desio", To: "Milano"})

	query := &fakeQuerySvc{
		stations: []domain.Station{{Code: "S01234", Name: "Desio"}},
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewAliasHandler(favs, bot.NewQueryHandler(query, 60*time.Minute)))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "HOME"))

	if query.gotStation != "S01234" {
		t.Errorf("case-insensitive alias failed: station=%q", query.gotStation)
	}
}

func TestAliasHandler_passesThroughWhenMessageHasArrow(t *testing.T) {
	favs := newFakeFavStore()
	// Even if a favorite named "milano" exists, `milano > bergamo` must
	// go to the query handler because it contains `>`.
	_ = favs.Save(42, domain.Favorite{Name: "milano", From: "Wrong", To: "Wrong"})

	query := &fakeQuerySvc{
		stations: []domain.Station{{Code: "MILANO", Name: "Milano"}},
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewAliasHandler(favs, bot.NewQueryHandler(query, 60*time.Minute)))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "Milano > Bergamo"))

	if query.gotStation != "MILANO" {
		t.Errorf("message with > should go to query handler; got station=%q", query.gotStation)
	}
	if query.gotTo != "Bergamo" {
		t.Errorf("TO should be 'Bergamo', got %q", query.gotTo)
	}
}

func TestAliasHandler_missFallsThroughToQuery(t *testing.T) {
	favs := newFakeFavStore()
	query := &fakeQuerySvc{}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewAliasHandler(favs, bot.NewQueryHandler(query, 60*time.Minute)))

	// No favorite named "gibberish" and no `>` — must fall through to
	// the existing format-hint reply.
	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "gibberish"))

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if !strings.Contains(strings.ToLower(msgs[0].Text), "format") {
		t.Errorf("expected format hint, got %q", msgs[0].Text)
	}
}

func TestHelp_mentionsFavoriteCommands(t *testing.T) {
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/help"))

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	lower := strings.ToLower(msgs[0].Text)
	for _, want := range []string{"/save", "/unsave", "/favorites"} {
		if !strings.Contains(lower, want) {
			t.Errorf("/help missing %q: %q", want, msgs[0].Text)
		}
	}
}

func TestStart_usesFromToGrammar(t *testing.T) {
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/start"))

	msg := sender.messages()[0].Text
	// The stale `S9 Desio` example from before ADR-010 must be gone.
	if strings.Contains(msg, "S9 Desio") {
		t.Errorf("/start still references pre-ADR-010 grammar: %q", msg)
	}
	if !strings.Contains(msg, ">") {
		t.Errorf("/start should show the FROM > TO grammar: %q", msg)
	}
}

// Acceptance: full round trip through the bot surface.
func TestFavorites_acceptanceRoundTrip(t *testing.T) {
	favs := newFakeFavStore()
	query := &fakeQuerySvc{
		stations: []domain.Station{{Code: "S01234", Name: "Desio"}},
		departures: []domain.Departure{
			{
				TrainCategory: "S",
				ScheduledTime: time.Date(2026, 4, 25, 8, 32, 0, 0, time.UTC),
				Destination:   "Milano Porta Garibaldi",
			},
		},
	}
	sender := &fakeSender{}

	d := bot.NewDispatcher([]int64{42})
	d.Register("/save", bot.NewSaveHandler(favs))
	d.Register("/unsave", bot.NewUnsaveHandler(favs))
	d.Register("/favorites", bot.NewFavoritesHandler(favs))
	d.OnText(bot.NewAliasHandler(favs, bot.NewQueryHandler(query, 60*time.Minute)))
	d.OnCallback("fr:", bot.NewFavoritesRunCallback(favs, query, 60*time.Minute))
	d.OnCallback("fd:", bot.NewFavoritesDeleteCallback(favs))
	d.OnCallback("q:", bot.NewCallbackHandler(query, 60*time.Minute))

	ctx := context.Background()

	// 1. Save
	_ = d.Handle(ctx, sender, newTextUpdate(42, "/save home Desio > Milano"))
	if _, ok := favs.Get(42, "home"); !ok {
		t.Fatal("step 1: save didn't persist")
	}

	// 2. /favorites lists it
	_ = d.Handle(ctx, sender, newTextUpdate(42, "/favorites"))
	kbs := sender.sentKeyboards()
	if len(kbs) == 0 {
		t.Fatal("step 2: /favorites produced no keyboard")
	}
	lastKB := kbs[len(kbs)-1]
	if !strings.Contains(lastKB.Text, "home") {
		t.Errorf("step 2: list missing favorite: %q", lastKB.Text)
	}

	// 3. Typing the nickname runs the route
	_ = d.Handle(ctx, sender, newTextUpdate(42, "home"))
	if query.gotStation != "S01234" || query.gotTo != "Milano" {
		t.Errorf("step 3: nickname did not run route: station=%q to=%q", query.gotStation, query.gotTo)
	}

	// 4. Tapping the run button also runs the route
	query.gotStation, query.gotTo = "", ""
	_ = d.Handle(ctx, sender,
		newCallbackUpdate(42, 100, "cb-1", "fr:home"))
	if query.gotStation != "S01234" || query.gotTo != "Milano" {
		t.Errorf("step 4: run button did not execute: station=%q to=%q", query.gotStation, query.gotTo)
	}

	// 5. Delete via /unsave
	_ = d.Handle(ctx, sender, newTextUpdate(42, "/unsave home"))
	if _, ok := favs.Get(42, "home"); ok {
		t.Fatal("step 5: /unsave didn't delete")
	}

	// 6. Nickname no longer works; alias falls through to format hint
	query.gotStation = ""
	_ = d.Handle(ctx, sender, newTextUpdate(42, "home"))
	if query.gotStation != "" {
		t.Error("step 6: deleted favorite should not resolve")
	}
	final := sender.messages()
	if !strings.Contains(strings.ToLower(final[len(final)-1].Text), "format") {
		t.Errorf("step 6: expected format hint after delete, got %q", final[len(final)-1].Text)
	}
}
