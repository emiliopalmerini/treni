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

// routeAfterStation runs the shared path for "we know line + station; now
// fetch and decide whether to list or ask for direction."
func routeAfterStation(
	ctx context.Context, s Sender, svc QueryService,
	target messageTarget, line string, station domain.Station, window time.Duration,
) error {
	deps, err := svc.QueryDepartures(ctx, line, station.Code, window)
	if err != nil {
		log.Printf("QueryDepartures %s @ %s: %v", line, station.Code, err)
		return target.renderText(ctx, s, upstreamDownMsg)
	}
	now := svc.Now()
	if len(deps) == 0 {
		return target.renderText(ctx, s, formatEmpty(line, station.Name, now, window))
	}

	termini := distinctTermini(deps)
	if len(termini) == 1 {
		return target.renderText(ctx, s, formatDepartures(line, station.Name, now, window, deps))
	}

	return target.renderWithButtons(ctx, s,
		"Which direction from "+station.Name+"?",
		directionButtons(line, station.Code, termini))
}

func distinctTermini(deps []domain.Departure) []string {
	seen := make(map[string]struct{}, 4)
	out := make([]string, 0, 4)
	for _, d := range deps {
		if _, ok := seen[d.Destination]; ok {
			continue
		}
		seen[d.Destination] = struct{}{}
		out = append(out, d.Destination)
	}
	return out
}

func directionButtons(line, stationCode string, termini []string) []Button {
	btns := make([]Button, 0, len(termini)+1)
	for _, t := range termini {
		btns = append(btns, Button{
			Text: "→ " + t,
			Data: directionCallback + line + ":" + stationCode + ":" + truncate(t, terminusMax),
		})
	}
	btns = append(btns, Button{
		Text: "Tutti",
		Data: directionCallback + line + ":" + stationCode + ":*",
	})
	return btns
}
