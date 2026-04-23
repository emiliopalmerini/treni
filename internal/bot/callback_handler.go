package bot

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/go-telegram/bot/models"
)

// NewCallbackHandler returns a handler for callback queries whose data
// matches `q:<LINE>:<STATION_CODE>`. The handler fetches departures,
// edits the originating message, and answers the callback query.
func NewCallbackHandler(svc QueryService, window time.Duration) CallbackHandler {
	return func(ctx context.Context, s Sender, cq *models.CallbackQuery) error {
		line, stationCode, ok := parseCallbackData(cq.Data)
		if !ok {
			log.Printf("malformed callback data: %q", cq.Data)
			return s.AnswerCallback(ctx, cq.ID)
		}

		chatID := callbackChatID(cq)
		messageID := callbackMessageID(cq)

		edit := func(ctx context.Context, s Sender, _ int64, text string) error {
			return s.EditMessageText(ctx, chatID, messageID, text)
		}

		_ = replyWithDepartures(
			ctx, s, svc, chatID, line,
			domain.Station{Code: stationCode, Name: stationCode},
			window, edit,
		)
		return s.AnswerCallback(ctx, cq.ID)
	}
}

func parseCallbackData(data string) (line, stationCode string, ok bool) {
	if !strings.HasPrefix(data, callbackPrefix) {
		return "", "", false
	}
	rest := strings.TrimPrefix(data, callbackPrefix)
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
