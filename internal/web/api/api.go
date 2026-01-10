package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/errgroup"

	"github.com/emiliopalmerini/treni/internal/app"
	"github.com/emiliopalmerini/treni/internal/observation"
	"github.com/emiliopalmerini/treni/internal/preferita"
	"github.com/emiliopalmerini/treni/internal/station"
	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
	"github.com/emiliopalmerini/treni/internal/voyage"
	"github.com/emiliopalmerini/treni/internal/web/converters"
	"github.com/emiliopalmerini/treni/internal/web/views"
)

// Handler handles HTMX API requests.
type Handler struct {
	config             *app.Config
	vtClient           viaggiatreno.Client
	stationService     *station.Service
	observationService *observation.Service
	preferitaService   *preferita.Service
	voyageService      *voyage.Service
}

// NewHandler creates a new API handler.
func NewHandler(
	config *app.Config,
	vtClient viaggiatreno.Client,
	stationService *station.Service,
	observationService *observation.Service,
	preferitaService *preferita.Service,
	voyageService *voyage.Service,
) *Handler {
	return &Handler{
		config:             config,
		vtClient:           vtClient,
		stationService:     stationService,
		observationService: observationService,
		preferitaService:   preferitaService,
		voyageService:      voyageService,
	}
}

// SearchStations handles station search requests.
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

	stationResults := make([]views.StationResult, len(stations))
	for i, s := range stations {
		stationResults[i] = views.StationResult{ID: s.ID, Name: s.Name}
	}

	views.StationSearchResults(stationResults).Render(r.Context(), w)
}

// GetFavoriteStations returns the list of favorite stations.
func (h *Handler) GetFavoriteStations(w http.ResponseWriter, r *http.Request) {
	preferite, err := h.preferitaService.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stationViews := make([]views.StationView, len(preferite))
	for i, p := range preferite {
		stationViews[i] = views.StationView{ID: p.StationID, Name: p.Name}
	}

	views.FavoriteStations(stationViews).Render(r.Context(), w)
}

// AddFavoriteStation adds a station to favorites.
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

	if err := h.preferitaService.Add(r.Context(), id, name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.GetFavoriteStations(w, r)
}

// DeleteFavoriteStation removes a station from favorites.
func (h *Handler) DeleteFavoriteStation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.preferitaService.Remove(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	views.EmptyRow().Render(r.Context(), w)
}

// GetNearestStation finds the nearest station to coordinates.
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

	// Check if previous position was provided
	prevLatStr := r.URL.Query().Get("prevLat")
	prevLonStr := r.URL.Query().Get("prevLon")
	prevStationID := r.URL.Query().Get("prevStationId")
	prevStationName := r.URL.Query().Get("prevStationName")

	if prevLatStr != "" && prevLonStr != "" && prevStationID != "" {
		prevLat, err1 := strconv.ParseFloat(prevLatStr, 64)
		prevLon, err2 := strconv.ParseFloat(prevLonStr, 64)

		if err1 == nil && err2 == nil {
			dist := station.Haversine(prevLat, prevLon, lat, lon)
			if dist <= h.config.NearbyStationDistanceKm {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"id":     prevStationID,
					"name":   prevStationName,
					"cached": "true",
				})
				return
			}
		}
	}

	st, err := h.stationService.FindNearest(r.Context(), lat, lon)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if st == nil {
		http.Error(w, "no stations found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":   st.ID,
		"name": st.Name,
	})
}

// GetStationTrains returns merged departures/arrivals for a station.
func (h *Handler) GetStationTrains(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "stationID")
	now := time.Now()

	var departures []viaggiatreno.Departure
	var arrivals []viaggiatreno.Arrival

	g, ctx := errgroup.WithContext(r.Context())

	g.Go(func() error {
		var err error
		departures, err = h.vtClient.Partenze(ctx, stationID, now)
		return err
	})

	g.Go(func() error {
		var err error
		arrivals, err = h.vtClient.Arrivi(ctx, stationID, now)
		return err
	})

	if err := g.Wait(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stationName := h.getStationName(r, stationID)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		h.observationService.RecordDepartures(ctx, stationID, stationName, departures)
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		h.observationService.RecordArrivals(ctx, stationID, stationName, arrivals)
	}()

	trains := mergeTrains(departures, arrivals)
	views.StationTable(trains).Render(r.Context(), w)
}

// GetStationStats returns statistics for a station.
func (h *Handler) GetStationStats(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "stationID")

	stats, err := h.observationService.GetStatsByStation(r.Context(), stationID)
	if err != nil || stats == nil {
		views.StationStatsEmpty().Render(r.Context(), w)
		return
	}

	views.StationStatsSection(converters.StationStat(stats)).Render(r.Context(), w)
}

// GetTrainDetail returns detailed train information.
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
	foundCurrent := false
	trainDelay := status.Delay
	for i, s := range status.Stops {
		stop := views.StopView{
			Name:     s.StationName,
			DelayArr: s.ArrivalDelay,
			DelayDep: s.DepartureDelay,
			Platform: s.EffectivePlatform(),
		}
		if s.ScheduledArrival > 0 {
			t := time.UnixMilli(s.ScheduledArrival)
			stop.ScheduledArr = t.Format("15:04")
			if s.ActualArrival > 0 {
				stop.IsPassed = true
				if s.ArrivalDelay > 0 {
					stop.ActualArr = t.Add(time.Duration(s.ArrivalDelay) * time.Minute).Format("15:04")
				}
			} else if trainDelay > 0 {
				stop.DelayArr = trainDelay
				stop.ActualArr = t.Add(time.Duration(trainDelay) * time.Minute).Format("15:04")
			}
		}
		if s.ScheduledDeparture > 0 {
			t := time.UnixMilli(s.ScheduledDeparture)
			stop.ScheduledDep = t.Format("15:04")
			if s.ActualDeparture > 0 || t.Before(now) {
				stop.IsPassed = true
				if s.DepartureDelay > 0 {
					stop.ActualDep = t.Add(time.Duration(s.DepartureDelay) * time.Minute).Format("15:04")
				}
			} else if trainDelay > 0 {
				stop.DelayDep = trainDelay
				stop.ActualDep = t.Add(time.Duration(trainDelay) * time.Minute).Format("15:04")
			}
		}
		if !stop.IsPassed && !foundCurrent {
			stop.IsCurrent = true
			foundCurrent = true
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

// GetRecentVoyages returns recent voyages.
func (h *Handler) GetRecentVoyages(w http.ResponseWriter, r *http.Request) {
	voyages, err := h.voyageService.GetRecentVoyages(r.Context(), 20)
	if err != nil {
		http.Error(w, "Failed to fetch voyages", http.StatusInternalServerError)
		return
	}

	voyageViews := converters.VoyageList(voyages)
	views.VoyageList(voyageViews).Render(r.Context(), w)
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

func mergeTrains(departures []viaggiatreno.Departure, arrivals []viaggiatreno.Arrival) []views.StationTrainView {
	trainMap := make(map[int]*views.StationTrainView, len(departures)+len(arrivals))

	for _, d := range departures {
		if d.DepartureTime <= 0 {
			continue
		}
		t := time.UnixMilli(d.DepartureTime)
		trainMap[d.TrainNumber] = &views.StationTrainView{
			TrainNumber:    d.TrainNumber,
			Category:       d.CategoryDesc,
			Origin:         d.Origin,
			Destination:    d.Destination,
			DepartureTime:  &t,
			DepartureDelay: d.Delay,
			Platform:       d.EffectivePlatform(),
			IsCancelled:    d.IsCancelled(),
		}
	}

	for _, a := range arrivals {
		if a.ArrivalTime <= 0 {
			continue
		}
		t := time.UnixMilli(a.ArrivalTime)
		if existing, ok := trainMap[a.TrainNumber]; ok {
			existing.ArrivalTime = &t
			existing.ArrivalDelay = a.Delay
			existing.IsCancelled = existing.IsCancelled || a.CirculationState == 1
			if existing.Origin == "" {
				existing.Origin = a.Origin
			}
			if existing.Destination == "" {
				existing.Destination = a.Destination
			}
			if existing.Platform == "" {
				existing.Platform = a.EffectivePlatform()
			}
		} else {
			trainMap[a.TrainNumber] = &views.StationTrainView{
				TrainNumber:  a.TrainNumber,
				Category:     a.CategoryDesc,
				Origin:       a.Origin,
				Destination:  a.Destination,
				ArrivalTime:  &t,
				ArrivalDelay: a.Delay,
				Platform:     a.EffectivePlatform(),
				IsCancelled:  a.CirculationState == 1,
			}
		}
	}

	trains := make([]views.StationTrainView, 0, len(trainMap))
	for _, t := range trainMap {
		trains = append(trains, *t)
	}

	sort.Slice(trains, func(i, j int) bool {
		return earliestTime(&trains[i]).Before(earliestTime(&trains[j]))
	})

	return trains
}

func earliestTime(t *views.StationTrainView) time.Time {
	if t.ArrivalTime != nil {
		return *t.ArrivalTime
	}
	if t.DepartureTime != nil {
		return *t.DepartureTime
	}
	return time.Time{}
}
