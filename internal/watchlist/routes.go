package watchlist

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/api/v1/watchlist", func(r chi.Router) {
		r.Get("/", h.List)
		r.Get("/active", h.ListActive)
		r.Get("/checks/recent", h.GetRecentChecks)
		r.Post("/", h.Create)
		r.Post("/check-all", h.CheckAllActive)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.Get)
			r.Put("/", h.Update)
			r.Delete("/", h.Delete)
			r.Post("/activate", h.Activate)
			r.Post("/deactivate", h.Deactivate)
			r.Post("/check", h.CheckTrain)
			r.Get("/checks", h.GetCheckHistory)
		})
	})
}
