package domain

import "time"

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
	TrainNumber     string
	TrainCategory   string
	Origin          string
	ScheduledTime   time.Time
	ActualTime      time.Time
	Delay           int
	Platform        string
	Status          TrainStatus
}

type Departure struct {
	TrainNumber     string
	TrainCategory   string
	Destination     string
	ScheduledTime   time.Time
	ActualTime      time.Time
	Delay           int
	Platform        string
	Status          TrainStatus
}
