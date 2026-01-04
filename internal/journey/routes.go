package journey

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/journeys", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Post("/record", h.RecordFromAPI)
		r.Get("/train/{trainNumber}", h.ListByTrain)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.Get)
			r.Get("/stops", h.GetStops)
			r.Put("/", h.Update)
			r.Delete("/", h.Delete)
		})
	})
}
