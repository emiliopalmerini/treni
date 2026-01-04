package main

import (
	"context"
	"log"

	"github.com/emiliopalmerini/treni/internal/app"
	"github.com/emiliopalmerini/treni/internal/cache"
	"github.com/emiliopalmerini/treni/internal/database"
	"github.com/emiliopalmerini/treni/internal/station"
	stationPersistence "github.com/emiliopalmerini/treni/internal/station/persistence"
	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

func main() {
	cfg, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	memCache := cache.NewMemory()
	httpClient := viaggiatreno.NewHTTPClient()
	vtClient := viaggiatreno.NewCachedClient(httpClient, memCache, cache.DefaultTTLConfig())

	stationRepo := stationPersistence.NewSQLiteRepository(db)
	stationService := station.NewService(stationRepo, vtClient)

	log.Println("starting station import...")
	if err := stationService.ImportAllStations(context.Background(), nil); err != nil {
		log.Fatalf("import failed: %v", err)
	}
}
