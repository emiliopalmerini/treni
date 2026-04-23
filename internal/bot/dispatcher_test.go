package bot_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/emiliopalmerini/treni/internal/bot"
	"github.com/go-telegram/bot/models"
)

type fakeSender struct {
	mu   sync.Mutex
	sent []sentMessage
}

type sentMessage struct {
	ChatID int64
	Text   string
}

func (f *fakeSender) SendMessage(_ context.Context, chatID int64, text string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, sentMessage{ChatID: chatID, Text: text})
	return nil
}

func (f *fakeSender) messages() []sentMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]sentMessage, len(f.sent))
	copy(out, f.sent)
	return out
}

func newUpdate(chatID int64, text string) *models.Update {
	return &models.Update{
		ID: 1,
		Message: &models.Message{
			ID:   1,
			Chat: models.Chat{ID: chatID},
			Text: text,
		},
	}
}

func TestDispatch_startRepliesWelcome(t *testing.T) {
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})

	if err := d.Handle(context.Background(), sender, newUpdate(42, "/start")); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if msgs[0].ChatID != 42 {
		t.Errorf("ChatID = %d, want 42", msgs[0].ChatID)
	}
	if !strings.Contains(strings.ToLower(msgs[0].Text), "ciao") {
		t.Errorf("welcome text missing 'ciao': %q", msgs[0].Text)
	}
}

func TestDispatch_helpReplies(t *testing.T) {
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})

	if err := d.Handle(context.Background(), sender, newUpdate(42, "/help")); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
}

func TestDispatch_nonWhitelistedDropped(t *testing.T) {
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})

	if err := d.Handle(context.Background(), sender, newUpdate(999, "/start")); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	msgs := sender.messages()
	if len(msgs) != 0 {
		t.Fatalf("got %d messages, want 0 (should be dropped silently)", len(msgs))
	}
}

func TestDispatch_unknownCommandFallback(t *testing.T) {
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})

	if err := d.Handle(context.Background(), sender, newUpdate(42, "/bogus")); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if !strings.Contains(strings.ToLower(msgs[0].Text), "unknown") {
		t.Errorf("fallback text missing 'unknown': %q", msgs[0].Text)
	}
}

func TestDispatch_plainTextFallback(t *testing.T) {
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})

	if err := d.Handle(context.Background(), sender, newUpdate(42, "hello bot")); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	msgs := sender.messages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
}

func TestDispatch_updateWithoutMessageIgnored(t *testing.T) {
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})

	update := &models.Update{ID: 1}
	if err := d.Handle(context.Background(), sender, update); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(sender.messages()) != 0 {
		t.Fatal("expected no messages for update without Message")
	}
}

func TestDispatch_messageWithoutChatIgnored(t *testing.T) {
	sender := &fakeSender{}
	d := bot.NewDispatcher([]int64{42})

	update := &models.Update{
		ID:      1,
		Message: &models.Message{ID: 1, Text: "/start"},
	}
	if err := d.Handle(context.Background(), sender, update); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(sender.messages()) != 0 {
		t.Fatal("expected no messages for message without Chat")
	}
}
