package observation

import (
	"context"
	"time"
)

// ObservationEvent represents an observation that was recorded.
type ObservationEvent struct {
	TrainNumber    int
	OriginID       string
	DepartureTime  time.Time
	StationID      string
	ArrivalDelay   int
	DepartureDelay int
	Platform       string
}

// ObservationNotifier is called when observations are recorded.
// Implementations can react to observations (e.g., update voyage data).
type ObservationNotifier interface {
	OnObservation(ctx context.Context, event ObservationEvent)
}

// NoopNotifier is a no-op implementation of ObservationNotifier.
type NoopNotifier struct{}

// OnObservation does nothing.
func (n *NoopNotifier) OnObservation(ctx context.Context, event ObservationEvent) {}
