package domain

import "time"

type Train struct {
	Number        string
	Category      string
	Origin        string
	Destination   string
	DepartureTime time.Time
	ArrivalTime   time.Time
	Delay         int
	Status        TrainStatus
	Stops         []Stop
	LastUpdate    time.Time
}

type TrainStatus string

const (
	TrainStatusOnTime    TrainStatus = "on_time"
	TrainStatusDelayed   TrainStatus = "delayed"
	TrainStatusCancelled TrainStatus = "cancelled"
	TrainStatusUnknown   TrainStatus = "unknown"
)

type Stop struct {
	StationCode       string
	StationName       string
	ScheduledArrival  time.Time
	ActualArrival     time.Time
	ScheduledDepart   time.Time
	ActualDepart      time.Time
	ArrivalDelay      int
	DepartureDelay    int
	Platform          string
	PlatformConfirmed bool
}
