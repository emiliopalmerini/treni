package preferita

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]*Preferita, error) {
	return s.repo.List(ctx)
}

func (s *Service) Add(ctx context.Context, stationID, name string) error {
	entity := &Preferita{
		StationID: stationID,
		Name:      name,
	}
	return s.repo.Add(ctx, entity)
}

func (s *Service) Remove(ctx context.Context, stationID string) error {
	return s.repo.Remove(ctx, stationID)
}

func (s *Service) Exists(ctx context.Context, stationID string) (bool, error) {
	return s.repo.Exists(ctx, stationID)
}
