package station

import "time"

type Station struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Region    int       `json:"region,omitempty"`
	Latitude  float64   `json:"latitude,omitempty"`
	Longitude float64   `json:"longitude,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
