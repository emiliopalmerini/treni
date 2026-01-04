package station

import (
	"context"

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
