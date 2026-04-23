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
	stationCallback   = "q:"
	directionCallback = "d:"
	buttonLabelMax    = 40
	terminusMax       = 40
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

		target := messageTarget{chatID: msg.Chat.ID}
		if len(stations) > 1 {
			return sendStationPicker(ctx, s, target, line, stations)
		}
		return routeAfterStation(ctx, s, svc, target, line, stations[0], window)
	}
}

func sendStationPicker(ctx context.Context, s Sender, target messageTarget, line string, stations []domain.Station) error {
	if len(stations) > maxStationChoices {
		stations = stations[:maxStationChoices]
	}
	buttons := make([]Button, len(stations))
	for i, st := range stations {
		buttons[i] = Button{
			Text: truncate(st.Name, buttonLabelMax),
			Data: stationCallback + line + ":" + st.Code,
		}
	}
	return target.renderWithButtons(ctx, s,
		"Multiple matches for "+line+". Pick one:", buttons)
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
