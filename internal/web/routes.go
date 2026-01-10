package web

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	// Static assets
	r.Get("/favicon.ico", Favicon)

	// HTML pages
	r.Get("/", h.Pages.StationPage)
	r.Get("/stats", h.Pages.StatsPage)
	r.Get("/voyage/{voyageID}", h.Pages.VoyageDetailPage)

	// HTMX API endpoints
	r.Route("/api", func(r chi.Router) {
		// Stations
		r.Get("/stations/search", h.API.SearchStations)
		r.Get("/stations/nearest", h.API.GetNearestStation)
		r.Get("/stations/favorites", h.API.GetFavoriteStations)
		r.Post("/stations/favorites", h.API.AddFavoriteStation)
		r.Delete("/stations/favorites/{id}", h.API.DeleteFavoriteStation)

		// Station trains (merged departures/arrivals)
		r.Get("/station/{stationID}", h.API.GetStationTrains)

		// Stats
		r.Get("/stats/station/{stationID}", h.API.GetStationStats)

		// Train detail (used by station page)
		r.Get("/train/{trainNumber}", h.API.GetTrainDetail)

		// Voyages
		r.Get("/voyages/recent", h.API.GetRecentVoyages)
	})
}
