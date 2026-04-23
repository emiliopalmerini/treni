package bot

import (
	"context"

	"github.com/go-telegram/bot/models"
)

const welcomeText = "Ciao. Send me `<LINE> <STATION>` (e.g. `S9 Desio`) or a train number. /help for more."

const unknownText = "Unknown command. /help for usage."

func startHandler(ctx context.Context, s Sender, msg *models.Message) error {
	return s.SendMessage(ctx, msg.Chat.ID, welcomeText)
}

func helpHandler(ctx context.Context, s Sender, msg *models.Message) error {
	return s.SendMessage(ctx, msg.Chat.ID, welcomeText)
}

func fallbackHandler(ctx context.Context, s Sender, msg *models.Message) error {
	return s.SendMessage(ctx, msg.Chat.ID, unknownText)
}
