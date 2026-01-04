package station

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/api/v1/stations", func(r chi.Router) {
		r.Get("/", h.List)
		r.Get("/favorites", h.ListFavorites)
		r.Get("/search", h.Search)
		r.Get("/search/live", h.SearchLive)
		r.Post("/", h.Create)
		r.Post("/import/{id}", h.Import)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.Get)
			r.Put("/", h.Update)
			r.Delete("/", h.Delete)
			r.Post("/favorite", h.SetFavorite)
			r.Delete("/favorite", h.UnsetFavorite)
		})
	})
}
