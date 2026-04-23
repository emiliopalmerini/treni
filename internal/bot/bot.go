package bot

import (
	"context"

	"github.com/go-telegram/bot/models"
)

type Sender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
	SendMessageWithButtons(ctx context.Context, chatID int64, text string, buttons []Button) error
	EditMessageText(ctx context.Context, chatID int64, messageID int, text string) error
	EditMessageWithButtons(ctx context.Context, chatID int64, messageID int, text string, buttons []Button) error
	AnswerCallback(ctx context.Context, callbackID string) error
}

type Button struct {
	Text string
	Data string
}

type Handler func(ctx context.Context, s Sender, msg *models.Message) error

type CallbackHandler func(ctx context.Context, s Sender, cq *models.CallbackQuery) error
