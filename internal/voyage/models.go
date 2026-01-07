package voyage

import (
	"time"

	"github.com/google/uuid"
)

type Voyage struct {
	ID                 uuid.UUID `json:"id"`
	TrainNumber        int       `json:"trainNumber"`
	TrainCategory      string    `json:"trainCategory"`
	OriginID           string    `json:"originId"`
	OriginName         string    `json:"originName"`
	DestinationID      string    `json:"destinationId"`
	DestinationName    string    `json:"destinationName"`
	ScheduledDate      string    `json:"scheduledDate"` // "2026-01-07" format for deduplication
	ScheduledDeparture time.Time `json:"scheduledDeparture"`
	CirculationState   int       `json:"circulationState"` // 0=normal, 1=cancelled, 2=partially cancelled
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type VoyageStop struct {
	ID                 uuid.UUID  `json:"id"`
	VoyageID           uuid.UUID  `json:"voyageId"`
	StationID          string     `json:"stationId"`
	StationName        string     `json:"stationName"`
	StopSequence       int        `json:"stopSequence"` // Order in route
	StopType           string     `json:"stopType"`     // 'P' (origin), 'A' (destination), 'F' (intermediate)
	ScheduledArrival   *time.Time `json:"scheduledArrival,omitempty"`
	ScheduledDeparture *time.Time `json:"scheduledDeparture,omitempty"`
	ActualArrival      *time.Time `json:"actualArrival,omitempty"`
	ActualDeparture    *time.Time `json:"actualDeparture,omitempty"`
	ArrivalDelay       int        `json:"arrivalDelay"`
	DepartureDelay     int        `json:"departureDelay"`
	Platform           string     `json:"platform,omitempty"`
	IsSuppressed       bool       `json:"isSuppressed"` // Stop was cancelled
	LastObservationAt  *time.Time `json:"lastObservationAt,omitempty"`
}

type VoyageWithStops struct {
	Voyage
	Stops []VoyageStop `json:"stops"`
}
