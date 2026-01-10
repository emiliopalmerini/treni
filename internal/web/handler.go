package web

import (
	"net/http"

	"github.com/emiliopalmerini/treni/internal/app"
	"github.com/emiliopalmerini/treni/internal/observation"
	"github.com/emiliopalmerini/treni/internal/preferita"
	"github.com/emiliopalmerini/treni/internal/station"
	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
	"github.com/emiliopalmerini/treni/internal/voyage"
	"github.com/emiliopalmerini/treni/internal/web/api"
	"github.com/emiliopalmerini/treni/internal/web/pages"
)

// Handler composes page and API handlers.
type Handler struct {
	Pages *pages.Handler
	API   *api.Handler
}

// NewHandler creates a new composite web handler.
func NewHandler(
	config *app.Config,
	vtClient viaggiatreno.Client,
	stationService *station.Service,
	observationService *observation.Service,
	preferitaService *preferita.Service,
	voyageService *voyage.Service,
) *Handler {
	return &Handler{
		Pages: pages.NewHandler(vtClient, stationService, observationService, voyageService),
		API:   api.NewHandler(config, vtClient, stationService, observationService, preferitaService, voyageService),
	}
}

// Favicon serves the favicon SVG.
func Favicon(w http.ResponseWriter, r *http.Request) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
<rect width="32" height="32" fill="black"/>
<text x="16" y="24" font-family="Arial,sans-serif" font-size="24" font-weight="bold" fill="white" text-anchor="middle">t</text>
</svg>`
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write([]byte(svg))
}
