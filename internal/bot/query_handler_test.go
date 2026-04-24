package bot_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/emiliopalmerini/treni/internal/bot"
	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/go-telegram/bot/models"
)

type fakeQuerySvc struct {
	stations      []domain.Station
	searchErr     error
	departures    []domain.Departure
	departuresErr error

	gotStation string
	gotTo      string
	gotWindow  time.Duration
}

func (f *fakeQuerySvc) SearchStations(ctx context.Context, q string) ([]domain.Station, error) {
	return f.stations, f.searchErr
}

func (f *fakeQuerySvc) DeparturesVia(ctx context.Context, stationCode, viaMatch string, window time.Duration) ([]domain.Departure, error) {
	f.gotStation = stationCode
	f.gotTo = viaMatch
	f.gotWindow = window
	return f.departures, f.departuresErr
}

func (f *fakeQuerySvc) Now() time.Time {
	return time.Date(2026, 4, 25, 8, 0, 0, 0, time.UTC)
}

func newTextUpdate(chatID int64, text string) *models.Update {
	return &models.Update{
		ID:      1,
		Message: &models.Message{ID: 1, Chat: models.Chat{ID: chatID}, Text: text},
	}
}

func TestQueryHandler_successPath(t *testing.T) {
	fc := &fakeQuerySvc{
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
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	if err := d.Handle(context.Background(), sender, newTextUpdate(42, "Desio: Milano")); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if fc.gotStation != "S01234" {
		t.Errorf("got station %q, want S01234", fc.gotStation)
	}
	if fc.gotTo != "Milano" {
		t.Errorf("got to %q, want Milano", fc.gotTo)
	}
	if fc.gotWindow != 60*time.Minute {
		t.Errorf("got window %v, want 60m", fc.gotWindow)
	}

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	text := msgs[0].Text
	for _, s := range []string{"Desio", "Milano", "Milano Porta Garibaldi"} {
		if !strings.Contains(text, s) {
			t.Errorf("reply missing %q: %q", s, text)
		}
	}
}

func TestQueryHandler_noStation(t *testing.T) {
	fc := &fakeQuerySvc{stations: nil}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "NotARealPlace: Milano"))

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d, want 1", len(msgs))
	}
	if !strings.Contains(strings.ToLower(msgs[0].Text), "no station") {
		t.Errorf("expected 'no station' reply, got %q", msgs[0].Text)
	}
}

func TestQueryHandler_noDepartures(t *testing.T) {
	fc := &fakeQuerySvc{
		stations:   []domain.Station{{Code: "S01234", Name: "Desio"}},
		departures: nil,
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "Desio: Milano"))

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d, want 1", len(msgs))
	}
	lower := strings.ToLower(msgs[0].Text)
	if !strings.Contains(lower, "no trains") {
		t.Errorf("expected 'no trains' reply, got %q", msgs[0].Text)
	}
	if !strings.Contains(lower, "now ") {
		t.Errorf("empty-state should include current time: %q", msgs[0].Text)
	}
}

func TestQueryHandler_searchError(t *testing.T) {
	fc := &fakeQuerySvc{searchErr: errors.New("upstream down")}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "Desio: Milano"))

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d, want 1", len(msgs))
	}
	if strings.Contains(msgs[0].Text, "upstream down") {
		t.Errorf("reply leaks internal error: %q", msgs[0].Text)
	}
}

func TestQueryHandler_badFormat(t *testing.T) {
	fc := &fakeQuerySvc{}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "gibberish"))

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d, want 1", len(msgs))
	}
	if !strings.Contains(strings.ToLower(msgs[0].Text), "format") {
		t.Errorf("reply missing format hint: %q", msgs[0].Text)
	}
}

func TestQueryHandler_slashStillGoesToCommand(t *testing.T) {
	fc := &fakeQuerySvc{}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "/start"))

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d, want 1", len(msgs))
	}
	if fc.gotStation != "" {
		t.Error("slash command must not reach query service")
	}
	if !strings.Contains(strings.ToLower(msgs[0].Text), "ciao") {
		t.Errorf("expected welcome, got %q", msgs[0].Text)
	}
}
