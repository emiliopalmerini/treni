package journey

import (
	"time"

	"github.com/google/uuid"
)

type Journey struct {
	ID                 uuid.UUID `json:"id"`
	TrainNumber        int       `json:"trainNumber"`
	OriginID           string    `json:"originId"`
	OriginName         string    `json:"originName"`
	DestinationID      string    `json:"destinationId"`
	DestinationName    string    `json:"destinationName"`
	ScheduledDeparture time.Time `json:"scheduledDeparture"`
	ActualDeparture    time.Time `json:"actualDeparture,omitempty"`
	Delay              int       `json:"delay"`
	RecordedAt         time.Time `json:"recordedAt"`
}

type JourneyStop struct {
	ID                 uuid.UUID `json:"id"`
	JourneyID          uuid.UUID `json:"journeyId"`
	StationID          string    `json:"stationId"`
	StationName        string    `json:"stationName"`
	ScheduledArrival   time.Time `json:"scheduledArrival,omitempty"`
	ScheduledDeparture time.Time `json:"scheduledDeparture,omitempty"`
	ActualArrival      time.Time `json:"actualArrival,omitempty"`
	ActualDeparture    time.Time `json:"actualDeparture,omitempty"`
	ArrivalDelay       int       `json:"arrivalDelay"`
	DepartureDelay     int       `json:"departureDelay"`
	Platform           string    `json:"platform,omitempty"`
}
