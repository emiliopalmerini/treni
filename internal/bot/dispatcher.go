package bot

import (
	"context"
	"strings"

	"github.com/go-telegram/bot/models"
)

type Dispatcher struct {
	allowed         map[int64]struct{}
	commands        map[string]Handler
	onText          Handler
	onUnknownCmd    Handler
	callbackRoutes  []callbackRoute
}

type callbackRoute struct {
	prefix  string
	handler CallbackHandler
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

func (d *Dispatcher) OnText(h Handler) {
	d.onText = h
}

func (d *Dispatcher) OnCallback(prefix string, h CallbackHandler) {
	d.callbackRoutes = append(d.callbackRoutes, callbackRoute{prefix: prefix, handler: h})
}

func (d *Dispatcher) Handle(ctx context.Context, s Sender, update *models.Update) error {
	if update == nil {
		return nil
	}
	if update.CallbackQuery != nil {
		return d.handleCallback(ctx, s, update.CallbackQuery)
	}
	if update.Message != nil {
		return d.handleMessage(ctx, s, update.Message)
	}
	return nil
}

func (d *Dispatcher) handleMessage(ctx context.Context, s Sender, msg *models.Message) error {
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

func (d *Dispatcher) handleCallback(ctx context.Context, s Sender, cq *models.CallbackQuery) error {
	chatID := callbackChatID(cq)
	if chatID == 0 {
		return nil
	}
	if _, ok := d.allowed[chatID]; !ok {
		return nil
	}
	for _, r := range d.callbackRoutes {
		if strings.HasPrefix(cq.Data, r.prefix) {
			return r.handler(ctx, s, cq)
		}
	}
	// Unknown callback: dismiss the spinner, do nothing else.
	return s.AnswerCallback(ctx, cq.ID)
}

func callbackChatID(cq *models.CallbackQuery) int64 {
	if cq == nil {
		return 0
	}
	if cq.Message.Message != nil {
		return cq.Message.Message.Chat.ID
	}
	return cq.From.ID
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
