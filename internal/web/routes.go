package web

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	// Static assets
	r.Get("/favicon.ico", Favicon)

	// HTML pages
	r.Get("/", h.Home)
	r.Get("/station", h.StationPage)
	r.Get("/watchlist", h.WatchlistPage)
	r.Get("/stats", h.StatsPage)
	r.Get("/voyage/{voyageID}", h.GetVoyageDetail)

	// HTMX API endpoints
	r.Route("/api", func(r chi.Router) {
		// Stations
		r.Get("/stations/search", h.SearchStations)
		r.Get("/stations/nearest", h.GetNearestStation)
		r.Get("/stations/favorites", h.GetFavoriteStations)
		r.Post("/stations/favorites", h.AddFavoriteStation)
		r.Delete("/stations/favorites/{id}", h.DeleteFavoriteStation)

		// Station trains (merged departures/arrivals)
		r.Get("/station/{stationID}", h.GetStationTrains)

		// Stats
		r.Get("/stats/station/{stationID}", h.GetStationStats)

		// Trains
		r.Get("/train/search", h.SearchTrains)
		r.Get("/train/{trainNumber}", h.GetTrainDetail)

		// Itinerary
		r.Get("/itinerary/search", h.SearchItinerary)

		// Watchlist
		r.Get("/watchlist", h.GetWatchlist)
		r.Post("/watchlist", h.AddToWatchlist)
		r.Post("/watchlist/check-all", h.CheckAllWatched)
		r.Post("/watchlist/{id}/check", h.CheckWatchedTrain)
		r.Delete("/watchlist/{id}", h.DeleteFromWatchlist)

		// Voyages
		r.Get("/voyages/recent", h.GetRecentVoyages)
	})
}
