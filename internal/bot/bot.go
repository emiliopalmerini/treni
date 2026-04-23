package bot

import (
	"context"

	"github.com/go-telegram/bot/models"
)

type Sender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type Handler func(ctx context.Context, s Sender, msg *models.Message) error
