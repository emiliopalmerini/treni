package itinerary

import "time"

type Station struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Leg struct {
	TrainNumber string    `json:"trainNumber"`
	TrainType   string    `json:"trainType"`
	From        Station   `json:"from"`
	To          Station   `json:"to"`
	DepartureAt time.Time `json:"departureAt"`
	ArrivalAt   time.Time `json:"arrivalAt"`
	Platform    string    `json:"platform"`
	Delay       int       `json:"delay"`
}

type Solution struct {
	Legs        []Leg         `json:"legs"`
	DepartureAt time.Time     `json:"departureAt"`
	ArrivalAt   time.Time     `json:"arrivalAt"`
	Duration    time.Duration `json:"duration"`
	Changes     int           `json:"changes"`
}

func NewSolution() *Solution {
	return &Solution{Legs: make([]Leg, 0)}
}

func (s *Solution) AddLeg(leg Leg) {
	s.Legs = append(s.Legs, leg)
	if len(s.Legs) == 1 {
		s.DepartureAt = leg.DepartureAt
	}
	s.ArrivalAt = leg.ArrivalAt
	s.Duration = s.ArrivalAt.Sub(s.DepartureAt)
	s.Changes = len(s.Legs) - 1
}

type SearchRequest struct {
	FromStationID string `json:"from"`
	ToStationID   string `json:"to"`
}
