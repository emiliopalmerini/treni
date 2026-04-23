package bot

import (
	"context"
	"strings"

	"github.com/go-telegram/bot/models"
)

type Dispatcher struct {
	allowed        map[int64]struct{}
	commands       map[string]Handler
	onText         Handler
	onUnknownCmd   Handler
}

func NewDispatcher(allowedChats []int64) *Dispatcher {
	allowed := make(map[int64]struct{}, len(allowedChats))
	for _, id := range allowedChats {
		allowed[id] = struct{}{}
	}
	d := &Dispatcher{
		allowed:      allowed,
		commands:     make(map[string]Handler),
		onText:       fallbackHandler,
		onUnknownCmd: fallbackHandler,
	}
	d.commands["/start"] = startHandler
	d.commands["/help"] = helpHandler
	return d
}

func (d *Dispatcher) Register(command string, h Handler) {
	d.commands[command] = h
}

// OnText sets the handler for non-slash messages.
func (d *Dispatcher) OnText(h Handler) {
	d.onText = h
}

func (d *Dispatcher) Handle(ctx context.Context, s Sender, update *models.Update) error {
	if update == nil || update.Message == nil {
		return nil
	}
	msg := update.Message
	if msg.Chat.ID == 0 {
		return nil
	}
	if _, ok := d.allowed[msg.Chat.ID]; !ok {
		return nil
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return nil
	}

	if strings.HasPrefix(text, "/") {
		cmd := firstToken(text)
		if h, ok := d.commands[cmd]; ok {
			return h(ctx, s, msg)
		}
		return d.onUnknownCmd(ctx, s, msg)
	}
	return d.onText(ctx, s, msg)
}

func firstToken(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if i := strings.IndexAny(text, " \t\n"); i >= 0 {
		return text[:i]
	}
	return text
}
