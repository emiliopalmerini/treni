package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/emiliopalmerini/treni/internal/api/viaggiatreno"
	"github.com/emiliopalmerini/treni/internal/service"
	"github.com/emiliopalmerini/treni/internal/storage"
	"github.com/emiliopalmerini/treni/internal/storage/sqlc"
	"github.com/emiliopalmerini/treni/web/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize API client
	apiClient := viaggiatreno.New()

	// Initialize database (optional - works without it)
	var queries *sqlc.Queries
	db, err := initDB()
	if err != nil {
		log.Printf("Warning: database not available: %v", err)
	} else {
		defer db.Close()
		queries = sqlc.New(db.DB)
	}

	// Initialize service and handlers
	svc := service.New(apiClient, queries)
	h := handlers.New(svc)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Pages
	r.Get("/", h.Home)
	r.Get("/train/{number}", h.Train)
	r.Get("/station/{code}", h.Station)
	r.Get("/analytics", h.Analytics)

	// HTMX API endpoints
	r.Route("/api", func(r chi.Router) {
		r.Get("/search", h.Search)
		r.Get("/train/{number}/status", h.TrainStatus)
		r.Get("/station/{code}/departures", h.StationDepartures)
		r.Get("/station/{code}/arrivals", h.StationArrivals)
		r.Get("/analytics/delayed", h.DelayedRankings)
		r.Get("/analytics/reliable", h.ReliableRankings)
	})

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// 404 handler
	r.NotFound(h.NotFound)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func initDB() (*storage.DB, error) {
	// Try Turso first
	if os.Getenv("TRENI_DATABASE_URL") != "" {
		return storage.New()
	}

	// Try local database
	localPath := os.Getenv("TRENI_DB_PATH")
	if localPath == "" {
		// Default location
		home, _ := os.UserHomeDir()
		localPath = home + "/.local/share/treni/treni.db"
	}

	// Check if local db exists
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no database configured")
	}

	return storage.NewLocal(localPath)
}
