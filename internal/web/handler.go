package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/treni/internal/station"
	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
	"github.com/emiliopalmerini/treni/internal/watchlist"
	"github.com/emiliopalmerini/treni/internal/web/views"
)

type Handler struct {
	vtClient         viaggiatreno.Client
	stationService   *station.Service
	watchlistService *watchlist.Service
}

func NewHandler(
	vtClient viaggiatreno.Client,
	stationService *station.Service,
	watchlistService *watchlist.Service,
) *Handler {
	return &Handler{
		vtClient:         vtClient,
		stationService:   stationService,
		watchlistService: watchlistService,
	}
}

// Pages

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	views.Home().Render(r.Context(), w)
}

func (h *Handler) DeparturesPage(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("station")
	if stationID == "" {
		stationID = "S01700" // Default to Milano Centrale
	}

	stationName := h.getStationName(r, stationID)
	views.DeparturesPage(stationName, stationID).Render(r.Context(), w)
}

func (h *Handler) ArrivalsPage(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("station")
	if stationID == "" {
		stationID = "S01700"
	}

	stationName := h.getStationName(r, stationID)
	views.ArrivalsPage(stationName, stationID).Render(r.Context(), w)
}

func (h *Handler) WatchlistPage(w http.ResponseWriter, r *http.Request) {
	views.WatchlistPage().Render(r.Context(), w)
}

// API endpoints for HTMX

func (h *Handler) SearchStations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		views.StationSearchResults(nil).Render(r.Context(), w)
		return
	}

	stations, err := h.vtClient.AutocompletaStazione(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	views.StationSearchResults(stations).Render(r.Context(), w)
}

func (h *Handler) GetFavoriteStations(w http.ResponseWriter, r *http.Request) {
	stations, err := h.stationService.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stationViews := make([]views.StationView, len(stations))
	for i, s := range stations {
		stationViews[i] = views.StationView{ID: s.ID, Name: s.Name}
	}

	views.FavoriteStations(stationViews).Render(r.Context(), w)
}

func (h *Handler) ImportStation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.stationService.Import(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteFavoriteStation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.stationService.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	views.EmptyRow().Render(r.Context(), w)
}

func (h *Handler) GetDepartures(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "stationID")

	departures, err := h.vtClient.Partenze(r.Context(), stationID, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	departureViews := make([]views.DepartureView, len(departures))
	for i, d := range departures {
		departureViews[i] = views.DepartureView{
			TrainNumber: d.TrainNumber,
			Category:    d.CategoryDesc,
			Destination: d.Destination,
			Time:        time.UnixMilli(d.DepartureTime),
			Delay:       d.Delay,
			Platform:    d.Platform,
			IsCancelled: d.CirculationState == 1,
		}
	}

	views.DeparturesTable(departureViews).Render(r.Context(), w)
}

func (h *Handler) GetArrivals(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "stationID")

	arrivals, err := h.vtClient.Arrivi(r.Context(), stationID, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	arrivalViews := make([]views.ArrivalView, len(arrivals))
	for i, a := range arrivals {
		arrivalViews[i] = views.ArrivalView{
			TrainNumber: a.TrainNumber,
			Category:    a.CategoryDesc,
			Origin:      a.Origin,
			Time:        time.UnixMilli(a.ArrivalTime),
			Delay:       a.Delay,
			Platform:    a.Platform,
			IsCancelled: a.CirculationState == 1,
		}
	}

	views.ArrivalsTable(arrivalViews).Render(r.Context(), w)
}

func (h *Handler) SearchTrains(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		views.TrainSearchResults(nil).Render(r.Context(), w)
		return
	}

	matches, err := h.vtClient.CercaNumeroTreno(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	trainViews := make([]views.TrainMatchView, len(matches))
	for i, m := range matches {
		trainViews[i] = views.TrainMatchView{
			Number:      m.Number,
			Origin:      m.Origin,
			OriginID:    m.OriginID,
			DepartureTS: m.DepartureTS,
		}
	}

	views.TrainSearchResults(trainViews).Render(r.Context(), w)
}

func (h *Handler) GetTrainDetail(w http.ResponseWriter, r *http.Request) {
	trainNumber := chi.URLParam(r, "trainNumber")
	originID := r.URL.Query().Get("origin")
	tsStr := r.URL.Query().Get("ts")

	var ts int64
	if tsStr != "" {
		ts, _ = strconv.ParseInt(tsStr, 10, 64)
	}

	if originID == "" || ts == 0 {
		matches, err := h.vtClient.CercaNumeroTreno(r.Context(), trainNumber)
		if err != nil || len(matches) == 0 {
			http.Error(w, "train not found", http.StatusNotFound)
			return
		}
		originID = matches[0].OriginID
		ts = matches[0].DepartureTS
	}

	status, err := h.vtClient.AndamentoTreno(r.Context(), originID, trainNumber, ts)
	if err != nil || status == nil {
		http.Error(w, "train status not available", http.StatusNotFound)
		return
	}

	stops := make([]views.StopView, len(status.Stops))
	now := time.Now()
	for i, s := range status.Stops {
		stop := views.StopView{
			Name:     s.StationName,
			DelayArr: s.ArrivalDelay,
			DelayDep: s.DepartureDelay,
			Platform: s.Platform,
		}
		if s.ScheduledArrival > 0 {
			t := time.UnixMilli(s.ScheduledArrival)
			stop.ScheduledArr = t.Format("15:04")
			if s.ActualArrival > 0 {
				stop.IsPassed = true
			}
		}
		if s.ScheduledDeparture > 0 {
			t := time.UnixMilli(s.ScheduledDeparture)
			stop.ScheduledDep = t.Format("15:04")
			if s.ActualDeparture > 0 || t.Before(now) {
				stop.IsPassed = true
			}
		}
		stops[i] = stop
	}

	trainView := views.TrainDetailView{
		Number:      status.TrainNumber,
		Category:    status.Category,
		Origin:      status.Origin,
		Destination: status.Destination,
		Delay:       status.Delay,
		Stops:       stops,
	}

	views.TrainDetail(trainView).Render(r.Context(), w)
}

// Watchlist API

func (h *Handler) GetWatchlist(w http.ResponseWriter, r *http.Request) {
	trains, err := h.watchlistService.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	trainViews := make([]views.WatchedTrainView, len(trains))
	for i, t := range trains {
		view := views.WatchedTrainView{
			ID:          t.ID.String(),
			TrainNumber: t.TrainNumber,
			OriginName:  t.OriginName,
			Destination: t.Destination,
			Active:      t.Active,
		}

		checks, _ := h.watchlistService.GetCheckHistory(r.Context(), t.ID)
		if len(checks) > 0 {
			view.LastCheck = &views.TrainCheckView{
				Delay:     checks[0].Delay,
				Status:    checks[0].Status,
				CheckedAt: checks[0].CheckedAt,
			}
		}

		trainViews[i] = view
	}

	views.WatchlistTable(trainViews).Render(r.Context(), w)
}

func (h *Handler) AddToWatchlist(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trainNumber, _ := strconv.Atoi(r.FormValue("trainNumber"))
	train := &watchlist.WatchedTrain{
		TrainNumber: trainNumber,
		OriginID:    r.FormValue("originId"),
		OriginName:  r.FormValue("originName"),
		Destination: r.FormValue("destination"),
	}

	if err := h.watchlistService.Create(r.Context(), train); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.GetWatchlist(w, r)
}

func (h *Handler) CheckWatchedTrain(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uid, err := parseUUID(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	check, err := h.watchlistService.CheckTrain(r.Context(), uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	train, _ := h.watchlistService.GetByID(r.Context(), uid)
	view := views.WatchedTrainView{
		ID:          train.ID.String(),
		TrainNumber: train.TrainNumber,
		OriginName:  train.OriginName,
		Destination: train.Destination,
		Active:      train.Active,
		LastCheck: &views.TrainCheckView{
			Delay:     check.Delay,
			Status:    check.Status,
			CheckedAt: check.CheckedAt,
		},
	}

	views.WatchedTrainRow(view).Render(r.Context(), w)
}

func (h *Handler) CheckAllWatched(w http.ResponseWriter, r *http.Request) {
	h.watchlistService.CheckAllActive(r.Context())
	h.GetWatchlist(w, r)
}

func (h *Handler) DeleteFromWatchlist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uid, err := parseUUID(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.watchlistService.Delete(r.Context(), uid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	views.EmptyRow().Render(r.Context(), w)
}

// Helpers

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

func parseUUID(s string) (watchlist.UUID, error) {
	return watchlist.ParseUUID(s)
}
