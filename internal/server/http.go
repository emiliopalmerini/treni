package server

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/treni/internal/app"
	"github.com/emiliopalmerini/treni/internal/journey"
	journeyPersistence "github.com/emiliopalmerini/treni/internal/journey/persistence"
	"github.com/emiliopalmerini/treni/internal/middleware"
	"github.com/emiliopalmerini/treni/internal/realtime"
	"github.com/emiliopalmerini/treni/internal/station"
	stationPersistence "github.com/emiliopalmerini/treni/internal/station/persistence"
	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
	"github.com/emiliopalmerini/treni/internal/watchlist"
	watchlistPersistence "github.com/emiliopalmerini/treni/internal/watchlist/persistence"
)

func NewHTTPServer(cfg *app.Config, db *sql.DB) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS)

	r.Get("/health", Health)

	// Initialize ViaggiaTreno client
	vtClient := viaggiatreno.NewClient()

	// Station module
	stationRepo := stationPersistence.NewSQLiteRepository(db)
	stationService := station.NewService(stationRepo, vtClient)
	stationHandler := station.NewHandler(stationService)
	station.RegisterRoutes(r, stationHandler)

	// Journey module
	journeyRepo := journeyPersistence.NewSQLiteRepository(db)
	journeyService := journey.NewService(journeyRepo, vtClient)
	journeyHandler := journey.NewHandler(journeyService)
	journey.RegisterRoutes(r, journeyHandler)

	// Realtime module
	realtimeHandler := realtime.NewHandler(vtClient)
	realtime.RegisterRoutes(r, realtimeHandler)

	// Watchlist module
	watchlistRepo := watchlistPersistence.NewSQLiteRepository(db)
	watchlistService := watchlist.NewService(watchlistRepo, vtClient)
	watchlistHandler := watchlist.NewHandler(watchlistService)
	watchlist.RegisterRoutes(r, watchlistHandler)

	return &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}
}
