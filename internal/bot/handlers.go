package bot

import (
	"context"

	"github.com/go-telegram/bot/models"
)

const startText = "Ciao. Send me <FROM> > <TO> (e.g. Desio > Milano) to see departures.\n/help for commands and favorites."

const helpText = `Query:
  <FROM> > <TO>        e.g. Desio > Milano

Favorites (per chat, max 10):
  /save <name> <FROM> > <TO>   save a route
  /unsave <name>               delete a saved route
  /favorites                   list saved routes
  <name>                       run a saved route

Times are the next 60 min from now.`

const unknownText = "Unknown command. /help for usage."

func startHandler(ctx context.Context, s Sender, msg *models.Message) error {
	return s.SendMessage(ctx, msg.Chat.ID, startText)
}

func helpHandler(ctx context.Context, s Sender, msg *models.Message) error {
	return s.SendMessage(ctx, msg.Chat.ID, helpText)
}

func fallbackHandler(ctx context.Context, s Sender, msg *models.Message) error {
	return s.SendMessage(ctx, msg.Chat.ID, unknownText)
}
