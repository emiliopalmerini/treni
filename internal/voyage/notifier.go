package voyage

import (
	"context"
	"log"

	"github.com/emiliopalmerini/treni/internal/observation"
)

// ObservationNotifier handles observation events by updating voyage data.
type ObservationNotifier struct {
	service *Service
}

// NewObservationNotifier creates a new voyage observation notifier.
func NewObservationNotifier(service *Service) *ObservationNotifier {
	return &ObservationNotifier{service: service}
}

// OnObservation processes an observation event.
func (n *ObservationNotifier) OnObservation(ctx context.Context, event observation.ObservationEvent) {
	if n.service == nil {
		return
	}

	voyageID, err := n.service.EnsureVoyageForTrain(ctx, event.TrainNumber, event.OriginID, event.DepartureTime)
	if err != nil {
		log.Printf("failed to ensure voyage for train %d: %v", event.TrainNumber, err)
		return
	}

	if err := n.service.UpdateVoyageStop(ctx, voyageID, event.StationID, event.ArrivalDelay, event.DepartureDelay, event.Platform); err != nil {
		log.Printf("failed to update voyage stop for train %d at station %s: %v", event.TrainNumber, event.StationID, err)
	}
}
