package domain

import "time"

type TrainStats struct {
	TrainNumber    string
	TotalTrips     int
	OnTimeTrips    int
	DelayedTrips   int
	CancelledTrips int
	AverageDelay   float64
	MaxDelay       int
	OnTimeRate     float64
	Period         StatsPeriod
}

type StatsPeriod struct {
	From time.Time
	To   time.Time
}
