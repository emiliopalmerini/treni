package bot

import (
	"context"
	"log"
	"time"

	"github.com/emiliopalmerini/treni/internal/domain"
)

// messageTarget abstracts "respond to this chat" vs "edit this existing message".
// messageID == 0 means a fresh send; non-zero means an edit.
type messageTarget struct {
	chatID    int64
	messageID int
}

func (t messageTarget) renderText(ctx context.Context, s Sender, text string) error {
	if t.messageID == 0 {
		return s.SendMessage(ctx, t.chatID, text)
	}
	return s.EditMessageText(ctx, t.chatID, t.messageID, text)
}

func (t messageTarget) renderWithButtons(ctx context.Context, s Sender, text string, buttons []Button) error {
	if t.messageID == 0 {
		return s.SendMessageWithButtons(ctx, t.chatID, text, buttons)
	}
	return s.EditMessageWithButtons(ctx, t.chatID, t.messageID, text, buttons)
}

// renderDepartures fetches and renders departures from `station` toward `to`.
func renderDepartures(
	ctx context.Context, s Sender, svc QueryService,
	target messageTarget, station domain.Station, to string, window time.Duration,
) error {
	deps, err := svc.DeparturesVia(ctx, station.Code, to, window)
	if err != nil {
		log.Printf("DeparturesVia %s via %q: %v", station.Code, to, err)
		return target.renderText(ctx, s, upstreamDownMsg)
	}
	now := svc.Now()
	if len(deps) == 0 {
		return target.renderText(ctx, s, formatEmpty(station.Name, to, now, window))
	}
	return target.renderText(ctx, s, formatDepartures(station.Name, to, now, window, deps))
}
