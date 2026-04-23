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
// Data format: `q:<STATION_CODE>:<TO>`.
func NewCallbackHandler(svc QueryService, window time.Duration) CallbackHandler {
	return func(ctx context.Context, s Sender, cq *models.CallbackQuery) error {
		stationCode, to, ok := parseStationCallback(cq.Data)
		if !ok {
			log.Printf("malformed station callback: %q", cq.Data)
			return s.AnswerCallback(ctx, cq.ID)
		}

		target := messageTarget{
			chatID:    callbackChatID(cq),
			messageID: callbackMessageID(cq),
		}
		// The callback data doesn't carry the friendly station name; use the
		// code as a label. Good enough for the edit header.
		_ = renderDepartures(ctx, s, svc, target,
			domain.Station{Code: stationCode, Name: stationCode}, to, window)
		return s.AnswerCallback(ctx, cq.ID)
	}
}

func parseStationCallback(data string) (stationCode, to string, ok bool) {
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

func callbackMessageID(cq *models.CallbackQuery) int {
	if cq == nil || cq.Message.Message == nil {
		return 0
	}
	return cq.Message.Message.ID
}
