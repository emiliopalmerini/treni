package bot

import (
	"context"

	tgbot "github.com/go-telegram/bot"
)

// tgSender adapts *tgbot.Bot to the Sender interface used by handlers.
type tgSender struct {
	b *tgbot.Bot
}

func NewSender(b *tgbot.Bot) Sender {
	return &tgSender{b: b}
}

func (s *tgSender) SendMessage(ctx context.Context, chatID int64, text string) error {
	_, err := s.b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	return err
}
