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
	"github.com/emiliopalmerini/treni/internal/middleware"
	"github.com/emiliopalmerini/treni/internal/observation"
	observationPersistence "github.com/emiliopalmerini/treni/internal/observation/persistence"
	"github.com/emiliopalmerini/treni/internal/preferita"
	preferitaPersistence "github.com/emiliopalmerini/treni/internal/preferita/persistence"
	"github.com/emiliopalmerini/treni/internal/realtime"
	"github.com/emiliopalmerini/treni/internal/staticdata"
	staticdataPersistence "github.com/emiliopalmerini/treni/internal/staticdata/persistence"
	"github.com/emiliopalmerini/treni/internal/station"
	stationPersistence "github.com/emiliopalmerini/treni/internal/station/persistence"
	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
	"github.com/emiliopalmerini/treni/internal/voyage"
	voyagePersistence "github.com/emiliopalmerini/treni/internal/voyage/persistence"
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
	stationRepoAdapter := staticdata.NewStationRepositoryAdapter(stationRepo)

	// Create composite provider for station lookups with fallback
	sqliteSource := staticdata.NewSQLiteStationSource(stationRepoAdapter)
	apiSource := staticdata.NewAPIStationSource(httpClient)
	compositeProvider := staticdata.NewCompositeStationProvider(
		cfg.StationStalenessAge,
		stationRepoAdapter,
		sqliteSource,
		apiSource,
	)
	stationProvider := staticdata.NewStationProviderAdapter(compositeProvider)

	stationService := station.NewService(stationRepo, vtClient).WithProvider(stationProvider)
	stationHandler := station.NewHandler(stationService)
	station.RegisterRoutes(r, stationHandler)

	// Static data import scheduler
	metadataRepo := staticdataPersistence.NewSQLiteMetadataRepository(db)

	if cfg.AutoImportEnabled {
		scheduler := staticdata.NewImportScheduler(
			metadataRepo,
			httpClient, // Use raw client for imports, not cached
			stationRepoAdapter,
			cfg.ImportRefreshInterval,
			cfg.StationStalenessAge,
		)
		scheduler.Start(context.Background())
		log.Printf("static data import scheduler started (refresh: %v, max age: %v)", cfg.ImportRefreshInterval, cfg.StationStalenessAge)
	}

	// Voyage module
	voyageRepo := voyagePersistence.NewSQLiteRepository(db)
	voyageService := voyage.NewService(voyageRepo, vtClient)

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
	observationService := observation.NewService(observationRepo, voyageService)

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
