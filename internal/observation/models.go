package observation

import (
	"time"

	"github.com/google/uuid"
)

type ObservationType string

const (
	ObservationTypeDeparture ObservationType = "departure"
	ObservationTypeArrival   ObservationType = "arrival"
)

type TrainObservation struct {
	ID               uuid.UUID       `json:"id"`
	ObservedAt       time.Time       `json:"observedAt"`
	StationID        string          `json:"stationId"`
	StationName      string          `json:"stationName"`
	ObservationType  ObservationType `json:"observationType"`
	TrainNumber      int             `json:"trainNumber"`
	TrainCategory    string          `json:"trainCategory"`
	OriginID         string          `json:"originId"`
	OriginName       string          `json:"originName"`
	DestinationID    string          `json:"destinationId"`
	DestinationName  string          `json:"destinationName"`
	ScheduledTime    time.Time       `json:"scheduledTime"`
	Delay            int             `json:"delay"`
	Platform         string          `json:"platform"`
	CirculationState int             `json:"circulationState"`
}

type GlobalStats struct {
	TotalObservations int     `json:"totalObservations"`
	AverageDelay      float64 `json:"averageDelay"`
	OnTimeCount       int     `json:"onTimeCount"`
	OnTimePercentage  float64 `json:"onTimePercentage"`
	CancelledCount    int     `json:"cancelledCount"`
}

type CategoryStats struct {
	Category         string  `json:"category"`
	ObservationCount int     `json:"observationCount"`
	AverageDelay     float64 `json:"averageDelay"`
	OnTimePercentage float64 `json:"onTimePercentage"`
}

type TrainStats struct {
	TrainNumber      int     `json:"trainNumber"`
	Category         string  `json:"category"`
	OriginID         string  `json:"originId"`
	OriginName       string  `json:"originName"`
	DestinationID    string  `json:"destinationId"`
	DestinationName  string  `json:"destinationName"`
	ObservationCount int     `json:"observationCount"`
	AverageDelay     float64 `json:"averageDelay"`
	MaxDelay         int     `json:"maxDelay"`
	OnTimePercentage float64 `json:"onTimePercentage"`
}

type StationStats struct {
	StationID        string  `json:"stationId"`
	StationName      string  `json:"stationName"`
	ObservationCount int     `json:"observationCount"`
	AverageDelay     float64 `json:"averageDelay"`
	OnTimePercentage float64 `json:"onTimePercentage"`
}
