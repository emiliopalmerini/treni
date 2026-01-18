package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/treni/internal/service"
	"github.com/emiliopalmerini/treni/web/templates"
)

type Handlers struct {
	svc *service.Service
}

func New(svc *service.Service) *Handlers {
	return &Handlers{svc: svc}
}

// Home renders the home page
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	templates.HomePage().Render(r.Context(), w)
}

// Search handles the search API endpoint
func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		w.WriteHeader(http.StatusOK)
		return
	}

	stations, err := h.svc.SearchStations(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.SearchResults(stations, query).Render(r.Context(), w)
}

// Train renders the train page
func (h *Handlers) Train(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")

	result, err := h.svc.GetTrain(r.Context(), number)
	if err != nil {
		templates.ErrorPage("Train Not Found", err.Error()).Render(r.Context(), w)
		return
	}

	templates.TrainPage(result).Render(r.Context(), w)
}

// TrainStatus returns the train status partial for HTMX refresh
func (h *Handlers) TrainStatus(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")

	result, err := h.svc.GetTrain(r.Context(), number)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.TrainStatusPartial(result.Train).Render(r.Context(), w)
}

// Station renders the station page
func (h *Handlers) Station(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	station, err := h.svc.GetStation(r.Context(), code)
	if err != nil {
		templates.ErrorPage("Station Not Found", err.Error()).Render(r.Context(), w)
		return
	}

	templates.StationPage(station).Render(r.Context(), w)
}

// StationDepartures returns the departures partial for HTMX
func (h *Handlers) StationDepartures(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	station, err := h.svc.GetStation(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.DeparturesPartial(station.Departures, code).Render(r.Context(), w)
}

// StationArrivals returns the arrivals partial for HTMX
func (h *Handlers) StationArrivals(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	station, err := h.svc.GetStation(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.ArrivalsPartial(station.Arrivals, code).Render(r.Context(), w)
}

// Analytics renders the analytics page
func (h *Handlers) Analytics(w http.ResponseWriter, r *http.Request) {
	templates.AnalyticsPage().Render(r.Context(), w)
}

// DelayedRankings returns the most delayed trains partial
func (h *Handlers) DelayedRankings(w http.ResponseWriter, r *http.Request) {
	trains, err := h.svc.GetMostDelayedTrains(r.Context(), 30, 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.DelayedRankings(trains).Render(r.Context(), w)
}

// ReliableRankings returns the most reliable trains partial
func (h *Handlers) ReliableRankings(w http.ResponseWriter, r *http.Request) {
	trains, err := h.svc.GetMostReliableTrains(r.Context(), 30, 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.ReliableRankings(trains).Render(r.Context(), w)
}

// NotFound renders the 404 page
func (h *Handlers) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	templates.NotFoundPage().Render(r.Context(), w)
}
