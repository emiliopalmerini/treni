package bot

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

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

func (s *tgSender) SendMessageWithButtons(ctx context.Context, chatID int64, text string, buttons []Button) error {
	_, err := s.b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: toInlineKeyboard(buttons),
	})
	return err
}

func (s *tgSender) EditMessageText(ctx context.Context, chatID int64, messageID int, text string) error {
	_, err := s.b.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      text,
	})
	return err
}

func (s *tgSender) AnswerCallback(ctx context.Context, callbackID string) error {
	_, err := s.b.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
		CallbackQueryID: callbackID,
	})
	return err
}

func toInlineKeyboard(buttons []Button) models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, 0, len(buttons))
	for _, btn := range buttons {
		rows = append(rows, []models.InlineKeyboardButton{{
			Text:         btn.Text,
			CallbackData: btn.Data,
		}})
	}
	return models.InlineKeyboardMarkup{InlineKeyboard: rows}
}
