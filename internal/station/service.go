package station

import (
	"context"
	"log"
	"math"

	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

type Service struct {
	repo   StationRepository
	client viaggiatreno.Client
}

func NewService(repo StationRepository, client viaggiatreno.Client) *Service {
	return &Service{repo: repo, client: client}
}

func (s *Service) Create(ctx context.Context, entity *Station) error {
	return s.repo.Create(ctx, entity)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Station, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*Station, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListFavorites(ctx context.Context) ([]*Station, error) {
	return s.repo.ListFavorites(ctx)
}

func (s *Service) Update(ctx context.Context, entity *Station) error {
	return s.repo.Update(ctx, entity)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) SetFavorite(ctx context.Context, id string, favorite bool) error {
	station, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	station.IsFavorite = favorite
	return s.repo.Update(ctx, station)
}

// SearchLive searches stations from the ViaggiaTreno API.
func (s *Service) SearchLive(ctx context.Context, query string) ([]*Station, error) {
	results, err := s.client.AutocompletaStazione(ctx, query)
	if err != nil {
		return nil, err
	}

	stations := make([]*Station, len(results))
	for i, r := range results {
		stations[i] = &Station{
			ID:   r.ID,
			Name: r.Name,
		}
	}
	return stations, nil
}

// Search searches stations from the local database.
func (s *Service) Search(ctx context.Context, query string) ([]*Station, error) {
	return s.repo.Search(ctx, query)
}

// Count returns the number of stations in the database.
func (s *Service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// ImportStats tracks import progress.
type ImportStats struct {
	TotalRegions     int
	CompletedRegions int
	TotalStations    int
	FailedStations   int
	InProgress       bool
}

// ImportAllStations fetches all stations from all regions and stores them locally.
func (s *Service) ImportAllStations(ctx context.Context, progress chan<- ImportStats) error {
	const totalRegions = 22
	stats := ImportStats{TotalRegions: totalRegions, InProgress: true}

	sendProgress := func() {
		if progress != nil {
			select {
			case progress <- stats:
			default:
			}
		}
	}

	sendProgress()

	for region := 1; region <= totalRegions; region++ {
		stations, err := s.client.ElencoStazioni(ctx, region)
		if err != nil {
			log.Printf("failed to fetch stations for region %d: %v", region, err)
			stats.CompletedRegions++
			sendProgress()
			continue
		}

		for _, rs := range stations {
			station := &Station{
				ID:        rs.ID,
				Name:      rs.Name,
				Region:    rs.Region,
				Latitude:  rs.Latitude,
				Longitude: rs.Longitude,
			}
			if err := s.repo.Upsert(ctx, station); err != nil {
				log.Printf("failed to upsert station %s: %v", rs.ID, err)
				stats.FailedStations++
			} else {
				stats.TotalStations++
			}
		}
		stats.CompletedRegions++
		log.Printf("imported %d stations from region %d (%d/%d)", len(stations), region, stats.CompletedRegions, totalRegions)
		sendProgress()
	}

	stats.InProgress = false
	sendProgress()
	log.Printf("import complete: %d stations imported, %d failed", stats.TotalStations, stats.FailedStations)
	return nil
}

// FindNearest finds the station closest to the given coordinates.
func (s *Service) FindNearest(ctx context.Context, lat, lon float64) (*Station, error) {
	stations, err := s.repo.ListWithCoordinates(ctx)
	if err != nil {
		return nil, err
	}

	if len(stations) == 0 {
		return nil, nil
	}

	var nearest *Station
	minDist := math.MaxFloat64

	for _, st := range stations {
		dist := haversine(lat, lon, st.Latitude, st.Longitude)
		if dist < minDist {
			minDist = dist
			nearest = st
		}
	}

	return nearest, nil
}

// haversine calculates the distance in km between two coordinates.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// Import fetches a station from the API and saves it locally.
func (s *Service) Import(ctx context.Context, id string) (*Station, error) {
	results, err := s.client.AutocompletaStazione(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, r := range results {
		if r.ID == id {
			station := &Station{
				ID:   r.ID,
				Name: r.Name,
			}
			if err := s.repo.Upsert(ctx, station); err != nil {
				return nil, err
			}
			return station, nil
		}
	}

	return nil, nil
}
