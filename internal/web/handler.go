package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/emiliopalmerini/treni/internal/observation"
	"github.com/emiliopalmerini/treni/internal/station"
	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
	"github.com/emiliopalmerini/treni/internal/watchlist"
	"github.com/emiliopalmerini/treni/internal/web/views"
)

type Handler struct {
	vtClient           viaggiatreno.Client
	stationService     *station.Service
	watchlistService   *watchlist.Service
	observationService *observation.Service
}

func NewHandler(
	vtClient viaggiatreno.Client,
	stationService *station.Service,
	watchlistService *watchlist.Service,
	observationService *observation.Service,
) *Handler {
	return &Handler{
		vtClient:           vtClient,
		stationService:     stationService,
		watchlistService:   watchlistService,
		observationService: observationService,
	}
}

// Pages

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	views.Home().Render(r.Context(), w)
}

func (h *Handler) DeparturesPage(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("station")
	detectLocation := stationID == ""
	if stationID == "" {
		stationID = "S01700" // Default to Milano Centrale
	}

	stationName := h.getStationName(r, stationID)
	views.DeparturesPage(stationName, stationID, detectLocation).Render(r.Context(), w)
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

func (h *Handler) StatsPage(w http.ResponseWriter, r *http.Request) {
	globalStats, _ := h.observationService.GetGlobalStats(r.Context())
	categoryStats, _ := h.observationService.GetStatsByCategory(r.Context())
	worstTrains, _ := h.observationService.GetWorstTrains(r.Context(), 10)
	worstStations, _ := h.observationService.GetWorstStations(r.Context(), 10)
	recentObs, _ := h.observationService.GetRecentObservations(r.Context(), 20)

	statsView := views.StatsPageView{
		Global:             toGlobalStatsView(globalStats),
		Categories:         toCategoryStatsViews(categoryStats),
		WorstTrains:        toTrainStatsViews(worstTrains),
		WorstStations:      toStationStatsViews(worstStations),
		RecentObservations: toObservationViews(recentObs),
	}

	views.StatsPage(statsView).Render(r.Context(), w)
}

func (h *Handler) GetStationStats(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "stationID")

	stats, err := h.observationService.GetStatsByStation(r.Context(), stationID)
	if err != nil || stats == nil {
		views.StationStatsEmpty().Render(r.Context(), w)
		return
	}

	views.StationStatsSection(toStationStatsView(stats)).Render(r.Context(), w)
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

func (h *Handler) AddFavoriteStation(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	name := r.FormValue("name")
	if id == "" || name == "" {
		http.Error(w, "id and name required", http.StatusBadRequest)
		return
	}

	s := &station.Station{
		ID:         id,
		Name:       name,
		IsFavorite: true,
	}
	if err := h.stationService.Create(r.Context(), s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.GetFavoriteStations(w, r)
}

func (h *Handler) DeleteFavoriteStation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.stationService.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	views.EmptyRow().Render(r.Context(), w)
}

func (h *Handler) GetNearestStation(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "invalid lat", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "invalid lon", http.StatusBadRequest)
		return
	}

	station, err := h.stationService.FindNearest(r.Context(), lat, lon)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if station == nil {
		http.Error(w, "no stations found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":   station.ID,
		"name": station.Name,
	})
}

func (h *Handler) GetDepartures(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "stationID")

	departures, err := h.vtClient.Partenze(r.Context(), stationID, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stationName := h.getStationName(r, stationID)
	go h.observationService.RecordDepartures(context.Background(), stationID, stationName, departures)

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

	stationName := h.getStationName(r, stationID)
	go h.observationService.RecordArrivals(context.Background(), stationID, stationName, arrivals)

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

// Static assets

func Favicon(w http.ResponseWriter, r *http.Request) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
<rect width="32" height="32" fill="black"/>
<text x="16" y="24" font-family="Arial,sans-serif" font-size="24" font-weight="bold" fill="white" text-anchor="middle">t</text>
</svg>`
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write([]byte(svg))
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

func toGlobalStatsView(s *observation.GlobalStats) views.GlobalStatsView {
	if s == nil {
		return views.GlobalStatsView{}
	}
	return views.GlobalStatsView{
		TotalObservations: s.TotalObservations,
		AverageDelay:      s.AverageDelay,
		OnTimePercentage:  s.OnTimePercentage,
		CancelledCount:    s.CancelledCount,
	}
}

func toCategoryStatsViews(stats []*observation.CategoryStats) []views.CategoryStatsView {
	result := make([]views.CategoryStatsView, len(stats))
	for i, s := range stats {
		result[i] = views.CategoryStatsView{
			Category:         s.Category,
			ObservationCount: s.ObservationCount,
			AverageDelay:     s.AverageDelay,
			OnTimePercentage: s.OnTimePercentage,
		}
	}
	return result
}

func toTrainStatsViews(stats []*observation.TrainStats) []views.TrainStatsView {
	result := make([]views.TrainStatsView, len(stats))
	for i, s := range stats {
		result[i] = views.TrainStatsView{
			TrainNumber:      s.TrainNumber,
			Category:         s.Category,
			OriginName:       s.OriginName,
			DestinationName:  s.DestinationName,
			ObservationCount: s.ObservationCount,
			AverageDelay:     s.AverageDelay,
			MaxDelay:         s.MaxDelay,
			OnTimePercentage: s.OnTimePercentage,
		}
	}
	return result
}

func toStationStatsViews(stats []*observation.StationStats) []views.StationStatsView {
	result := make([]views.StationStatsView, len(stats))
	for i, s := range stats {
		result[i] = toStationStatsView(s)
	}
	return result
}

func toStationStatsView(s *observation.StationStats) views.StationStatsView {
	if s == nil {
		return views.StationStatsView{}
	}
	return views.StationStatsView{
		StationID:        s.StationID,
		StationName:      s.StationName,
		ObservationCount: s.ObservationCount,
		AverageDelay:     s.AverageDelay,
		OnTimePercentage: s.OnTimePercentage,
	}
}

func toObservationViews(obs []*observation.TrainObservation) []views.ObservationView {
	result := make([]views.ObservationView, len(obs))
	for i, o := range obs {
		result[i] = views.ObservationView{
			ObservedAt:      o.ObservedAt,
			StationName:     o.StationName,
			ObservationType: string(o.ObservationType),
			TrainNumber:     o.TrainNumber,
			TrainCategory:   o.TrainCategory,
			OriginName:      o.OriginName,
			DestinationName: o.DestinationName,
			Delay:           o.Delay,
			IsCancelled:     o.CirculationState == 1,
		}
	}
	return result
}
