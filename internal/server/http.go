package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/treni/internal/app"
	"github.com/emiliopalmerini/treni/internal/middleware"
)

func NewHTTPServer(cfg *app.Config) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS)

	r.Get("/health", Health)

	return &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}
}
