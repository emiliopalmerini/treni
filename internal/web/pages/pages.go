package pages

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/observation"
	"github.com/emiliopalmerini/treni/internal/station"
	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
	"github.com/emiliopalmerini/treni/internal/voyage"
	"github.com/emiliopalmerini/treni/internal/web/converters"
	"github.com/emiliopalmerini/treni/internal/web/views"
)

// Handler handles page rendering.
type Handler struct {
	vtClient           viaggiatreno.Client
	stationService     *station.Service
	observationService *observation.Service
	voyageService      *voyage.Service
}

// NewHandler creates a new pages handler.
func NewHandler(
	vtClient viaggiatreno.Client,
	stationService *station.Service,
	observationService *observation.Service,
	voyageService *voyage.Service,
) *Handler {
	return &Handler{
		vtClient:           vtClient,
		stationService:     stationService,
		observationService: observationService,
		voyageService:      voyageService,
	}
}

// StationPage renders the station picker or station view page.
func (h *Handler) StationPage(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("station")
	if stationID == "" {
		views.StationPicker().Render(r.Context(), w)
		return
	}

	stationName := h.getStationName(r, stationID)
	views.StationPage(stationName, stationID).Render(r.Context(), w)
}

// StatsPage renders the statistics dashboard.
func (h *Handler) StatsPage(w http.ResponseWriter, r *http.Request) {
	globalStats, _ := h.observationService.GetGlobalStats(r.Context())
	categoryStats, _ := h.observationService.GetStatsByCategory(r.Context())
	worstTrains, _ := h.observationService.GetWorstTrains(r.Context(), 10)
	worstStations, _ := h.observationService.GetWorstStations(r.Context(), 10)
	recentVoyages, _ := h.voyageService.GetRecentVoyages(r.Context(), 10)
	recentObs, _ := h.observationService.GetRecentObservations(r.Context(), 20)

	statsView := views.StatsPageView{
		Global:             converters.GlobalStats(globalStats),
		Categories:         converters.CategoryStats(categoryStats),
		WorstTrains:        converters.TrainStats(worstTrains),
		WorstStations:      converters.StationStats(worstStations),
		RecentVoyages:      converters.VoyageList(recentVoyages),
		RecentObservations: converters.Observations(recentObs),
	}

	views.StatsPage(statsView).Render(r.Context(), w)
}

// VoyageDetailPage renders the voyage detail page.
func (h *Handler) VoyageDetailPage(w http.ResponseWriter, r *http.Request) {
	voyageIDStr := chi.URLParam(r, "voyageID")
	voyageID, err := uuid.Parse(voyageIDStr)
	if err != nil {
		http.Error(w, "Invalid voyage ID", http.StatusBadRequest)
		return
	}

	voyageWithStops, err := h.voyageService.GetVoyageWithStops(r.Context(), voyageID)
	if err != nil {
		http.Error(w, "Voyage not found", http.StatusNotFound)
		return
	}

	voyageView := converters.VoyageDetail(voyageWithStops)
	views.VoyageDetailPage(voyageView).Render(r.Context(), w)
}

func (h *Handler) getStationName(r *http.Request, stationID string) string {
	s, err := h.stationService.GetByID(r.Context(), stationID)
	if err == nil && s != nil {
		return s.Name
	}

	stations, err := h.vtClient.AutocompletaStazione(r.Context(), stationID)
	if err == nil && len(stations) > 0 {
		return stations[0].Name
	}

	return stationID
}
