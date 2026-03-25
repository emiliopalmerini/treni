package handlers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/emiliopalmerini/treni/internal/service"
	"github.com/emiliopalmerini/treni/web/templates"
)

type Handlers struct {
	svc *service.Service
}

func New(svc *service.Service) *Handlers {
	return &Handlers{svc: svc}
}

// Home renders the station picker or the multi-station dashboard.
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	saved := parseStations(r)
	if len(saved) == 0 {
		templates.StationPickerPage(false).Render(r.Context(), w)
		return
	}

	// Load the first station's data for initial render
	active := saved[0]
	station, err := h.svc.GetStation(r.Context(), active.Code)
	if err != nil {
		log.Printf("failed to load station %s: %v", active.Code, err)
		// Skip bad station, try remaining
		remaining := removeStation(saved, active.Code)
		writeStationsCookie(w, remaining)
		if len(remaining) == 0 {
			templates.StationPickerPage(false).Render(r.Context(), w)
			return
		}
		// Redirect to retry with remaining stations
		w.Header().Set("HX-Redirect", "/")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	templates.DashboardPage(saved, active.Code, station).Render(r.Context(), w)
}

// AddStation renders the station picker for adding a new station.
func (h *Handlers) AddStation(w http.ResponseWriter, r *http.Request) {
	saved := parseStations(r)
	if len(saved) >= maxStations {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	templates.StationPickerPage(len(saved) > 0).Render(r.Context(), w)
}

// PickStation adds a station to the saved list and redirects to home.
func (h *Handlers) PickStation(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	name := r.URL.Query().Get("name")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	saved := parseStations(r)
	updated, ok := addStation(saved, domain.StationRef{Code: code, Name: name})
	if !ok {
		http.Error(w, "maximum stations reached", http.StatusConflict)
		return
	}

	writeStationsCookie(w, updated)
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusNoContent)
}

// RemoveStation removes a station from the saved list.
func (h *Handlers) RemoveStation(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	saved := parseStations(r)
	updated := removeStation(saved, code)
	writeStationsCookie(w, updated)
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusNoContent)
}

// StationContent returns the board content partial for a station (HTMX swap).
func (h *Handlers) StationContent(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	station, err := h.svc.GetStation(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.StationContentPartial(station).Render(r.Context(), w)
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

	mode := r.URL.Query().Get("mode")
	if mode == "picker" {
		templates.PickerSearchResults(stations, query).Render(r.Context(), w)
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

// NotFound renders the 404 page
func (h *Handlers) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	templates.NotFoundPage().Render(r.Context(), w)
}
