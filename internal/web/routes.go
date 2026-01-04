package web

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	// HTML pages
	r.Get("/", h.Home)
	r.Get("/departures", h.DeparturesPage)
	r.Get("/arrivals", h.ArrivalsPage)
	r.Get("/watchlist", h.WatchlistPage)

	// HTMX API endpoints
	r.Route("/api", func(r chi.Router) {
		// Stations
		r.Get("/stations/search", h.SearchStations)
		r.Get("/stations/favorites", h.GetFavoriteStations)
		r.Post("/stations/import/{id}", h.ImportStation)
		r.Delete("/stations/favorites/{id}", h.DeleteFavoriteStation)

		// Departures/Arrivals
		r.Get("/departures/{stationID}", h.GetDepartures)
		r.Get("/arrivals/{stationID}", h.GetArrivals)

		// Trains
		r.Get("/train/search", h.SearchTrains)
		r.Get("/train/{trainNumber}", h.GetTrainDetail)

		// Watchlist
		r.Get("/watchlist", h.GetWatchlist)
		r.Post("/watchlist", h.AddToWatchlist)
		r.Post("/watchlist/check-all", h.CheckAllWatched)
		r.Post("/watchlist/{id}/check", h.CheckWatchedTrain)
		r.Delete("/watchlist/{id}", h.DeleteFromWatchlist)
	})
}
