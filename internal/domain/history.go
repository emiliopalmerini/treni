package domain

import "time"

type DelayRecord struct {
	ID            int64
	TrainNumber   string
	TrainCategory string
	Origin        string
	Destination   string
	Date          time.Time
	Delay         int
	Cancelled     bool
	Source        string
	RecordedAt    time.Time
}

type TrainStats struct {
	TrainNumber     string
	TotalTrips      int
	OnTimeTrips     int
	DelayedTrips    int
	CancelledTrips  int
	AverageDelay    float64
	MaxDelay        int
	OnTimeRate      float64
	Period          StatsPeriod
}

type StatsPeriod struct {
	From time.Time
	To   time.Time
}
