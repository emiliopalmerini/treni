package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/emiliopalmerini/treni/internal/bot"
	"github.com/emiliopalmerini/treni/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dispatcher := bot.NewDispatcher(cfg.AllowedChats)

	var sender bot.Sender
	defaultHandler := func(hctx context.Context, b *tgbot.Bot, update *models.Update) {
		if err := dispatcher.Handle(hctx, sender, update); err != nil {
			log.Printf("dispatch: %v", err)
		}
	}

	b, err := tgbot.New(cfg.BotToken, tgbot.WithDefaultHandler(defaultHandler))
	if err != nil {
		log.Fatalf("bot init: %v", err)
	}
	sender = bot.NewSender(b)

	log.Printf("trenibot starting (allowed chats: %d)", len(cfg.AllowedChats))
	b.Start(ctx)
}
