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
		station := stations[0]

		deps, err := svc.QueryDepartures(ctx, line, station.Code, window)
		if err != nil {
			log.Printf("QueryDepartures %s @ %s: %v", line, station.Code, err)
			return s.SendMessage(ctx, msg.Chat.ID, upstreamDownMsg)
		}
		if len(deps) == 0 {
			return s.SendMessage(ctx, msg.Chat.ID,
				"No "+line+" departures from "+station.Name+" in the next "+fmtMin(window)+".")
		}

		return s.SendMessage(ctx, msg.Chat.ID, formatDepartures(line, station.Name, window, deps))
	}
}

func fmtMin(d time.Duration) string {
	return strconv.Itoa(int(d.Minutes())) + " min"
}
