package domain

import "time"

// StationRef is a lightweight reference to a saved station (code + display name).
type StationRef struct {
	Code string
	Name string
}

type Station struct {
	Code       string
	Name       string
	City       string
	Region     string
	Latitude   float64
	Longitude  float64
	Arrivals   []Arrival
	Departures []Departure
}

type Arrival struct {
	TrainNumber   string
	TrainCategory string
	Origin        string
	ScheduledTime time.Time
	ActualTime    time.Time
	Delay         int
	Platform      string
	Status        TrainStatus
}

type Departure struct {
	TrainNumber   string
	TrainCategory string
	Destination   string
	ScheduledTime time.Time
	ActualTime    time.Time
	Delay         int
	Platform      string
	Status        TrainStatus
}
