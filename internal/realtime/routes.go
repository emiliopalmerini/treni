package realtime

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/realtime", func(r chi.Router) {
		r.Get("/departures/{stationID}", h.Departures)
		r.Get("/arrivals/{stationID}", h.Arrivals)
		r.Get("/train/{trainNumber}", h.TrainStatus)
		r.Get("/train/{trainNumber}/detailed", h.TrainStatusDetailed)
		r.Get("/search/trains", h.SearchTrains)
	})
}
