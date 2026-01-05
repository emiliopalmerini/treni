package staticdata

import "time"

// Station represents a train station with import metadata.
type Station struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Region    int       `json:"region,omitempty"`
	Latitude  float64   `json:"latitude,omitempty"`
	Longitude float64   `json:"longitude,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// ImportMetadata tracks the status of data imports.
type ImportMetadata struct {
	EntityType   string
	LastImport   time.Time
	RecordCount  int
	DurationMs   int64
	Status       string
	ErrorMessage string
}

// DataFreshness indicates how current the data is.
type DataFreshness struct {
	Source      string
	LastUpdated time.Time
	IsStale     bool
}
