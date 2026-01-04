package cache

import "time"

// TTLConfig holds TTL settings for different data types.
type TTLConfig struct {
	Static     time.Duration // Station data: 24h
	SemiStatic time.Duration // Train number lookups: 1h
	Realtime   time.Duration // Departures, arrivals, train status: 30s
}

// DefaultTTLConfig returns sensible defaults.
func DefaultTTLConfig() TTLConfig {
	return TTLConfig{
		Static:     24 * time.Hour,
		SemiStatic: 1 * time.Hour,
		Realtime:   30 * time.Second,
	}
}
