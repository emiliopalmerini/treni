package bot_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/emiliopalmerini/treni/internal/bot"
	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/go-telegram/bot/models"
)

func newCallbackUpdate(chatID int64, messageID int, callbackID, data string) *models.Update {
	return &models.Update{
		ID: 1,
		CallbackQuery: &models.CallbackQuery{
			ID:      callbackID,
			From:    models.User{ID: chatID},
			Message: models.MaybeInaccessibleMessage{Message: &models.Message{ID: messageID, Chat: models.Chat{ID: chatID}}},
			Data:    data,
		},
	}
}

func TestQueryHandler_multipleStationsShowsPicker(t *testing.T) {
	fc := &fakeQuerySvc{
		stations: []domain.Station{
			{Code: "S01234", Name: "Milano Centrale"},
			{Code: "S01645", Name: "Milano Porta Garibaldi"},
			{Code: "S01701", Name: "Milano Lambrate"},
		},
	}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "RV Milano"))

	if len(sender.messages()) != 0 {
		t.Errorf("expected no plain messages, got %d", len(sender.messages()))
	}
	kbs := sender.sentKeyboards()
	if len(kbs) != 1 {
		t.Fatalf("got %d keyboards, want 1", len(kbs))
	}
	kb := kbs[0]
	if kb.ChatID != 42 {
		t.Errorf("chat = %d, want 42", kb.ChatID)
	}
	if len(kb.Buttons) != 3 {
		t.Fatalf("got %d buttons, want 3", len(kb.Buttons))
	}
	if !strings.Contains(kb.Buttons[0].Data, "S01234") {
		t.Errorf("button 0 data missing code: %q", kb.Buttons[0].Data)
	}
	if !strings.Contains(kb.Buttons[0].Data, "RV") {
		t.Errorf("button 0 data missing line: %q", kb.Buttons[0].Data)
	}
	if kb.Buttons[0].Text != "Milano Centrale" {
		t.Errorf("button 0 text = %q, want Milano Centrale", kb.Buttons[0].Text)
	}
}

func TestQueryHandler_cappedAtMaxChoices(t *testing.T) {
	stations := []domain.Station{
		{Code: "A", Name: "A"}, {Code: "B", Name: "B"},
		{Code: "C", Name: "C"}, {Code: "D", Name: "D"},
		{Code: "E", Name: "E"}, {Code: "F", Name: "F"},
		{Code: "G", Name: "G"},
	}
	fc := &fakeQuerySvc{stations: stations}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnText(bot.NewQueryHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender, newTextUpdate(42, "S9 San"))

	kbs := sender.sentKeyboards()
	if len(kbs) != 1 {
		t.Fatalf("want 1 keyboard, got %d", len(kbs))
	}
	if len(kbs[0].Buttons) != 5 {
		t.Errorf("got %d buttons, want 5 (capped)", len(kbs[0].Buttons))
	}
}

func TestCallbackHandler_decodesAndReplies(t *testing.T) {
	fc := &fakeQuerySvc{
		departures: []domain.Departure{
			{
				TrainCategory: "RV",
				ScheduledTime: time.Date(2026, 4, 24, 14, 32, 0, 0, time.UTC),
				Destination:   "Brescia",
				Platform:      "7",
			},
		},
	}
	// Pre-seed the station name in the service so the callback handler can
	// format a nice header. For this test, we accept that the handler may
	// fall back to the station code if it doesn't have a name.
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnCallback("q:", bot.NewCallbackHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender,
		newCallbackUpdate(42, 101, "cb-xyz", "q:RV:S01234"))

	if fc.gotLine != "RV" {
		t.Errorf("line = %q, want RV", fc.gotLine)
	}
	if fc.gotStation != "S01234" {
		t.Errorf("station = %q, want S01234", fc.gotStation)
	}

	edits := sender.editedMessages()
	if len(edits) != 1 {
		t.Fatalf("got %d edits, want 1", len(edits))
	}
	if edits[0].MessageID != 101 {
		t.Errorf("message id = %d, want 101", edits[0].MessageID)
	}
	if !strings.Contains(edits[0].Text, "Brescia") {
		t.Errorf("edit missing destination: %q", edits[0].Text)
	}

	ans := sender.answered()
	if len(ans) != 1 || ans[0] != "cb-xyz" {
		t.Errorf("callback not answered: %v", ans)
	}
}

func TestCallbackHandler_nonWhitelistedDropped(t *testing.T) {
	fc := &fakeQuerySvc{}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnCallback("q:", bot.NewCallbackHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender,
		newCallbackUpdate(999, 101, "cb", "q:RV:S01234"))

	if len(sender.editedMessages()) != 0 {
		t.Errorf("expected no edits for non-whitelisted callback")
	}
	if len(sender.answered()) != 0 {
		t.Errorf("expected no answers for non-whitelisted callback")
	}
	if fc.gotLine != "" {
		t.Errorf("service invoked for non-whitelisted callback")
	}
}

func TestCallbackHandler_malformedDataDropped(t *testing.T) {
	fc := &fakeQuerySvc{}
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})
	d.OnCallback("q:", bot.NewCallbackHandler(fc, 60*time.Minute))

	_ = d.Handle(context.Background(), sender,
		newCallbackUpdate(42, 101, "cb", "q:onlytwo"))

	// Malformed callback: answer it so Telegram dismisses the spinner,
	// but don't invoke the service or edit the message.
	if fc.gotLine != "" {
		t.Errorf("service invoked on malformed callback")
	}
	ans := sender.answered()
	if len(ans) != 1 {
		t.Errorf("expected 1 answer, got %d", len(ans))
	}
}
