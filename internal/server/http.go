package server

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/treni/internal/app"
	"github.com/emiliopalmerini/treni/internal/cache"
	"github.com/emiliopalmerini/treni/internal/itinerary"
	"github.com/emiliopalmerini/treni/internal/journey"
	journeyPersistence "github.com/emiliopalmerini/treni/internal/journey/persistence"
	"github.com/emiliopalmerini/treni/internal/middleware"
	"github.com/emiliopalmerini/treni/internal/observation"
	observationPersistence "github.com/emiliopalmerini/treni/internal/observation/persistence"
	"github.com/emiliopalmerini/treni/internal/preferita"
	preferitaPersistence "github.com/emiliopalmerini/treni/internal/preferita/persistence"
	"github.com/emiliopalmerini/treni/internal/realtime"
	"github.com/emiliopalmerini/treni/internal/station"
	stationPersistence "github.com/emiliopalmerini/treni/internal/station/persistence"
	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
	"github.com/emiliopalmerini/treni/internal/watchlist"
	watchlistPersistence "github.com/emiliopalmerini/treni/internal/watchlist/persistence"
	"github.com/emiliopalmerini/treni/internal/web"
)

func NewHTTPServer(cfg *app.Config, db *sql.DB) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS)

	r.Get("/health", Health)

	// Initialize cache and ViaggiaTreno client
	memCache := cache.NewMemory()
	httpClient := viaggiatreno.NewHTTPClient()
	vtClient := viaggiatreno.NewCachedClient(httpClient, memCache, cache.DefaultTTLConfig())

	// Station module
	stationRepo := stationPersistence.NewSQLiteRepository(db)
	stationService := station.NewService(stationRepo, vtClient)
	stationHandler := station.NewHandler(stationService)
	station.RegisterRoutes(r, stationHandler)

	// Import stations in background if none exist
	go func() {
		count, err := stationService.Count(context.Background())
		if err != nil {
			log.Printf("failed to count stations: %v", err)
			return
		}
		if count == 0 {
			log.Println("no stations found, importing all stations...")
			if err := stationService.ImportAllStations(context.Background(), nil); err != nil {
				log.Printf("failed to import stations: %v", err)
			}
		}
	}()

	// Journey module
	journeyRepo := journeyPersistence.NewSQLiteRepository(db)
	journeyService := journey.NewService(journeyRepo, vtClient)
	journeyHandler := journey.NewHandler(journeyService)
	journey.RegisterRoutes(r, journeyHandler)

	// Realtime module
	realtimeHandler := realtime.NewHandler(vtClient)
	realtime.RegisterRoutes(r, realtimeHandler)

	// Itinerary module
	itineraryService := itinerary.NewService(vtClient)

	// Watchlist module
	watchlistRepo := watchlistPersistence.NewSQLiteRepository(db)
	watchlistService := watchlist.NewService(watchlistRepo, vtClient)
	watchlistHandler := watchlist.NewHandler(watchlistService)
	watchlist.RegisterRoutes(r, watchlistHandler)

	// Observation module
	observationRepo := observationPersistence.NewSQLiteRepository(db)
	observationService := observation.NewService(observationRepo)

	// Preferita module
	preferitaRepo := preferitaPersistence.NewSQLiteRepository(db)
	preferitaService := preferita.NewService(preferitaRepo)

	// Web UI
	webHandler := web.NewHandler(vtClient, stationService, watchlistService, observationService, preferitaService, itineraryService)
	web.RegisterRoutes(r, webHandler)

	return &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}
}
