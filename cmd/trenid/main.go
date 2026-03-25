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
	"github.com/emiliopalmerini/treni/web/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	apiClient := viaggiatreno.New()
	svc := service.New(apiClient)
	h := handlers.New(svc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Pages
	r.Get("/", h.Home)
	r.Get("/add-station", h.AddStation)
	r.Get("/train/{number}", h.Train)
	r.Get("/station/{code}", h.Station)

	// HTMX API endpoints
	r.Route("/api", func(r chi.Router) {
		r.Get("/search", h.Search)
		r.Get("/pick-station", h.PickStation)
		r.Get("/remove-station", h.RemoveStation)
		r.Get("/train/{number}/status", h.TrainStatus)
		r.Get("/station/{code}/content", h.StationContent)
		r.Get("/station/{code}/departures", h.StationDepartures)
		r.Get("/station/{code}/arrivals", h.StationArrivals)
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
