package bot

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/go-telegram/bot/models"
)

type QueryService interface {
	SearchStations(ctx context.Context, query string) ([]domain.Station, error)
	QueryDepartures(ctx context.Context, line, stationCode string, window time.Duration) ([]domain.Departure, error)
}

const (
	formatHint      = "Query format: <LINE> <STATION>. Example: S9 Desio. /help for more."
	upstreamDownMsg = "Couldn't reach ViaggiaTreno. Try again in a sec."

	maxStationChoices = 5
	callbackPrefix    = "q:"
	buttonLabelMax    = 40
)

func NewQueryHandler(svc QueryService, window time.Duration) Handler {
	return func(ctx context.Context, s Sender, msg *models.Message) error {
		line, stationQuery, ok := parseQuery(msg.Text)
		if !ok {
			return s.SendMessage(ctx, msg.Chat.ID, formatHint)
		}

		stations, err := svc.SearchStations(ctx, stationQuery)
		if err != nil {
			log.Printf("SearchStations %q: %v", stationQuery, err)
			return s.SendMessage(ctx, msg.Chat.ID, upstreamDownMsg)
		}
		if len(stations) == 0 {
			return s.SendMessage(ctx, msg.Chat.ID, "No station found for '"+stationQuery+"'.")
		}
		if len(stations) > 1 {
			return sendPicker(ctx, s, msg.Chat.ID, line, stations)
		}
		return replyWithDepartures(ctx, s, svc, msg.Chat.ID, line, stations[0], window, sendReply)
	}
}

func sendPicker(ctx context.Context, s Sender, chatID int64, line string, stations []domain.Station) error {
	if len(stations) > maxStationChoices {
		stations = stations[:maxStationChoices]
	}
	buttons := make([]Button, len(stations))
	for i, st := range stations {
		buttons[i] = Button{
			Text: truncate(st.Name, buttonLabelMax),
			Data: callbackPrefix + line + ":" + st.Code,
		}
	}
	return s.SendMessageWithButtons(ctx, chatID,
		"Multiple matches for "+line+". Pick one:", buttons)
}

func replyWithDepartures(
	ctx context.Context, s Sender, svc QueryService,
	chatID int64, line string, station domain.Station, window time.Duration,
	reply func(ctx context.Context, s Sender, chatID int64, text string) error,
) error {
	deps, err := svc.QueryDepartures(ctx, line, station.Code, window)
	if err != nil {
		log.Printf("QueryDepartures %s @ %s: %v", line, station.Code, err)
		return reply(ctx, s, chatID, upstreamDownMsg)
	}
	if len(deps) == 0 {
		return reply(ctx, s, chatID,
			"No "+line+" departures from "+station.Name+" in the next "+fmtMin(window)+".")
	}
	return reply(ctx, s, chatID, formatDepartures(line, station.Name, window, deps))
}

func sendReply(ctx context.Context, s Sender, chatID int64, text string) error {
	return s.SendMessage(ctx, chatID, text)
}

func fmtMin(d time.Duration) string {
	return strconv.Itoa(int(d.Minutes())) + " min"
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 1 {
		return ""
	}
	return s[:max-1] + "…"
}
