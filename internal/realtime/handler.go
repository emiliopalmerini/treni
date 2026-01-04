package realtime

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

type Handler struct {
	client viaggiatreno.Client
}

func NewHandler(client viaggiatreno.Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) Departures(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "stationID")
	if stationID == "" {
		http.Error(w, "missing stationID", http.StatusBadRequest)
		return
	}

	departures, err := h.client.Partenze(r.Context(), stationID, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(departures)
}

func (h *Handler) Arrivals(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "stationID")
	if stationID == "" {
		http.Error(w, "missing stationID", http.StatusBadRequest)
		return
	}

	arrivals, err := h.client.Arrivi(r.Context(), stationID, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(arrivals)
}

func (h *Handler) TrainStatus(w http.ResponseWriter, r *http.Request) {
	trainNumber := chi.URLParam(r, "trainNumber")
	if trainNumber == "" {
		http.Error(w, "missing trainNumber", http.StatusBadRequest)
		return
	}

	matches, err := h.client.CercaNumeroTreno(r.Context(), trainNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(matches) == 0 {
		http.Error(w, "train not found", http.StatusNotFound)
		return
	}

	match := matches[0]
	status, err := h.client.AndamentoTreno(r.Context(), match.OriginID, match.Number, match.DepartureTS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if status == nil {
		http.Error(w, "train status not available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *Handler) SearchTrains(w http.ResponseWriter, r *http.Request) {
	trainNumber := r.URL.Query().Get("q")
	if trainNumber == "" {
		http.Error(w, "missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	matches, err := h.client.CercaNumeroTreno(r.Context(), trainNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

type trainStatusResponse struct {
	TrainNumber int                       `json:"trainNumber"`
	Origin      string                    `json:"origin"`
	Destination string                    `json:"destination"`
	Delay       int                       `json:"delay"`
	Stops       []viaggiatreno.Stop       `json:"stops"`
	Status      *viaggiatreno.TrainStatus `json:"raw,omitempty"`
}

func (h *Handler) TrainStatusDetailed(w http.ResponseWriter, r *http.Request) {
	trainNumber := chi.URLParam(r, "trainNumber")
	originID := r.URL.Query().Get("origin")
	tsStr := r.URL.Query().Get("ts")

	if trainNumber == "" {
		http.Error(w, "missing trainNumber", http.StatusBadRequest)
		return
	}

	var ts int64
	if tsStr != "" {
		var err error
		ts, err = strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid timestamp", http.StatusBadRequest)
			return
		}
	}

	if originID == "" || ts == 0 {
		matches, err := h.client.CercaNumeroTreno(r.Context(), trainNumber)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(matches) == 0 {
			http.Error(w, "train not found", http.StatusNotFound)
			return
		}
		originID = matches[0].OriginID
		ts = matches[0].DepartureTS
	}

	status, err := h.client.AndamentoTreno(r.Context(), originID, trainNumber, ts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if status == nil {
		http.Error(w, "train status not available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
