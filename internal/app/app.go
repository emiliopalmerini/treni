package app

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr            string        `envconfig:"ADDR" default:":8080"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"5s"`
	DatabasePath    string        `envconfig:"DATABASE_PATH" default:"treni.db"`

	// Static data settings
	StationStalenessAge   time.Duration `envconfig:"STATION_STALENESS_AGE" default:"720h"` // 30 days
	AutoImportEnabled     bool          `envconfig:"AUTO_IMPORT_ENABLED" default:"true"`
	ImportRefreshInterval time.Duration `envconfig:"IMPORT_REFRESH_INTERVAL" default:"168h"` // 7 days

	// Geolocation settings
	NearbyStationDistanceKm float64 `envconfig:"NEARBY_STATION_DISTANCE_KM" default:"5.0"`
}

func New() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
