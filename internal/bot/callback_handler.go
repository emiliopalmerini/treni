package bot

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/go-telegram/bot/models"
)

// NewCallbackHandler handles the station-pick callback (`q:` prefix).
// It edits the picker message with either the direction picker (if the
// selected station has multiple termini in window) or the final list.
func NewCallbackHandler(svc QueryService, window time.Duration) CallbackHandler {
	return func(ctx context.Context, s Sender, cq *models.CallbackQuery) error {
		line, stationCode, ok := parseStationCallback(cq.Data)
		if !ok {
			log.Printf("malformed station callback: %q", cq.Data)
			return s.AnswerCallback(ctx, cq.ID)
		}

		target := messageTarget{
			chatID:    callbackChatID(cq),
			messageID: callbackMessageID(cq),
		}
		_ = routeAfterStation(ctx, s, svc, target, line,
			domain.Station{Code: stationCode, Name: stationCode}, window)
		return s.AnswerCallback(ctx, cq.ID)
	}
}

// NewDirectionHandler handles the direction-pick callback (`d:` prefix).
// It re-fetches departures, filters to the picked terminus (or all), and
// edits the message with the list.
func NewDirectionHandler(svc QueryService, window time.Duration) CallbackHandler {
	return func(ctx context.Context, s Sender, cq *models.CallbackQuery) error {
		line, stationCode, terminus, ok := parseDirectionCallback(cq.Data)
		if !ok {
			log.Printf("malformed direction callback: %q", cq.Data)
			return s.AnswerCallback(ctx, cq.ID)
		}

		chatID := callbackChatID(cq)
		messageID := callbackMessageID(cq)

		deps, err := svc.QueryDepartures(ctx, line, stationCode, window)
		if err != nil {
			log.Printf("QueryDepartures %s @ %s: %v", line, stationCode, err)
			_ = s.EditMessageText(ctx, chatID, messageID, upstreamDownMsg)
			return s.AnswerCallback(ctx, cq.ID)
		}

		if terminus != "" {
			deps = filterByTerminus(deps, terminus)
		}

		stationName := stationCode // callback data doesn't carry the friendly name
		var text string
		if len(deps) == 0 {
			text = "No " + line + " departures from " + stationName + " in the next " + fmtMin(window) + "."
		} else {
			text = formatDepartures(line, stationName, window, deps)
		}
		_ = s.EditMessageText(ctx, chatID, messageID, text)
		return s.AnswerCallback(ctx, cq.ID)
	}
}

func parseStationCallback(data string) (line, stationCode string, ok bool) {
	if !strings.HasPrefix(data, stationCallback) {
		return "", "", false
	}
	rest := strings.TrimPrefix(data, stationCallback)
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func parseDirectionCallback(data string) (line, stationCode, terminus string, ok bool) {
	if !strings.HasPrefix(data, directionCallback) {
		return "", "", "", false
	}
	rest := strings.TrimPrefix(data, directionCallback)
	parts := strings.SplitN(rest, ":", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", false
	}
	t := parts[2]
	if t == "*" {
		t = "" // empty signals "no filter"
	}
	return parts[0], parts[1], t, true
}

// filterByTerminus returns departures whose destination starts with terminus
// (case-insensitive). Prefix match covers the case where the callback data
// carried a truncated terminus.
func filterByTerminus(deps []domain.Departure, terminus string) []domain.Departure {
	out := deps[:0]
	tLower := strings.ToLower(terminus)
	for _, d := range deps {
		if strings.HasPrefix(strings.ToLower(d.Destination), tLower) {
			out = append(out, d)
		}
	}
	return out
}

func callbackMessageID(cq *models.CallbackQuery) int {
	if cq == nil || cq.Message.Message == nil {
		return 0
	}
	return cq.Message.Message.ID
}
