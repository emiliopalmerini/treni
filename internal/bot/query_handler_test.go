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

	gotLine    string
	gotStation string
	gotWindow  time.Duration
}

func (f *fakeQuerySvc) SearchStations(ctx context.Context, q string) ([]domain.Station, error) {
	return f.stations, f.searchErr
}

func (f *fakeQuerySvc) QueryDepartures(ctx context.Context, line, stationCode string, window time.Duration) ([]domain.Departure, error) {
	f.gotLine = line
	f.gotStation = stationCode
	f.gotWindow = window
	return f.departures, f.departuresErr
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
				TrainCategory: "S9",
				ScheduledTime: time.Date(2026, 4, 24, 14, 32, 0, 0, time.UTC),
				Destination:   "Saronno",
				Delay:         3,
				Platform:      "2",
			},
		},
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	if err := d.Handle(context.Background(), sender, newTextUpdate(42, "S9 Desio")); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if fc.gotLine != "S9" {
		t.Errorf("service got line %q, want S9", fc.gotLine)
	}
	if fc.gotStation != "S01234" {
		t.Errorf("service got station %q, want S01234", fc.gotStation)
	}
	if fc.gotWindow != 60*time.Minute {
		t.Errorf("service got window %v, want 60m", fc.gotWindow)
	}

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	text := msgs[0].Text
	if !strings.Contains(text, "S9") {
		t.Errorf("reply missing line: %q", text)
	}
	if !strings.Contains(text, "Desio") {
		t.Errorf("reply missing station: %q", text)
	}
	if !strings.Contains(text, "Saronno") {
		t.Errorf("reply missing destination: %q", text)
	}
}

func TestQueryHandler_noStation(t *testing.T) {
	fc := &fakeQuerySvc{stations: nil}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "S9 NotARealPlace"))

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
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

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "S9 Desio"))

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	lower := strings.ToLower(msgs[0].Text)
	if !strings.Contains(lower, "no") || !strings.Contains(lower, "s9") {
		t.Errorf("expected 'no S9' reply, got %q", msgs[0].Text)
	}
}

func TestQueryHandler_searchError(t *testing.T) {
	fc := &fakeQuerySvc{searchErr: errors.New("upstream down")}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "S9 Desio"))

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	// User-facing text must not leak "upstream down"
	if strings.Contains(msgs[0].Text, "upstream down") {
		t.Errorf("reply leaks internal error: %q", msgs[0].Text)
	}
	if !strings.Contains(strings.ToLower(msgs[0].Text), "viaggiatreno") &&
		!strings.Contains(strings.ToLower(msgs[0].Text), "try again") {
		t.Errorf("reply missing friendly error: %q", msgs[0].Text)
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
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if !strings.Contains(strings.ToLower(msgs[0].Text), "format") &&
		!strings.Contains(strings.ToLower(msgs[0].Text), "example") {
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
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if fc.gotLine != "" {
		t.Error("slash command must not reach query service")
	}
	if !strings.Contains(strings.ToLower(msgs[0].Text), "ciao") {
		t.Errorf("expected welcome, got %q", msgs[0].Text)
	}
}
