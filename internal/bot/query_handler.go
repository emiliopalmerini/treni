package bot

import (
	"context"
	"log"
	"time"

	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/go-telegram/bot/models"
)

type QueryService interface {
	SearchStations(ctx context.Context, query string) ([]domain.Station, error)
	DeparturesVia(ctx context.Context, stationCode, viaMatch string, window time.Duration) ([]domain.Departure, error)
	Now() time.Time
}

const (
	formatHint      = "Query format: <FROM> > <TO>. Example: Desio > Milano. /help for more."
	upstreamDownMsg = "Couldn't reach ViaggiaTreno. Try again in a sec."

	maxStationChoices = 5
	stationCallback   = "q:"
	buttonLabelMax    = 40
	toMax             = 40
)

func NewQueryHandler(svc QueryService, window time.Duration) Handler {
	return func(ctx context.Context, s Sender, msg *models.Message) error {
		from, to, ok := parseFromTo(msg.Text)
		if !ok {
			return s.SendMessage(ctx, msg.Chat.ID, formatHint)
		}

		stations, err := svc.SearchStations(ctx, from)
		if err != nil {
			log.Printf("SearchStations %q: %v", from, err)
			return s.SendMessage(ctx, msg.Chat.ID, upstreamDownMsg)
		}
		if len(stations) == 0 {
			return s.SendMessage(ctx, msg.Chat.ID, "No station found for '"+from+"'.")
		}

		target := messageTarget{chatID: msg.Chat.ID}
		if len(stations) > 1 {
			return sendStationPicker(ctx, s, target, stations, to)
		}
		return renderDepartures(ctx, s, svc, target, stations[0], to, window)
	}
}

func sendStationPicker(ctx context.Context, s Sender, target messageTarget, stations []domain.Station, to string) error {
	if len(stations) > maxStationChoices {
		stations = stations[:maxStationChoices]
	}
	truncatedTo := truncate(to, toMax)
	buttons := make([]Button, len(stations))
	for i, st := range stations {
		buttons[i] = Button{
			Text: truncate(st.Name, buttonLabelMax),
			Data: stationCallback + st.Code + ":" + truncatedTo,
		}
	}
	return target.renderWithButtons(ctx, s,
		"Multiple matches for origin. Pick one:", buttons)
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
