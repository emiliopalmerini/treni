package staticdata

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

// ImportScheduler handles background refresh of static data.
type ImportScheduler struct {
	metaRepo        ImportMetadataRepository
	vtClient        viaggiatreno.Client
	stationRepo     StationRepository
	refreshInterval time.Duration
	maxAge          time.Duration

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewImportScheduler creates a new import scheduler.
func NewImportScheduler(
	metaRepo ImportMetadataRepository,
	vtClient viaggiatreno.Client,
	stationRepo StationRepository,
	refreshInterval time.Duration,
	maxAge time.Duration,
) *ImportScheduler {
	return &ImportScheduler{
		metaRepo:        metaRepo,
		vtClient:        vtClient,
		stationRepo:     stationRepo,
		refreshInterval: refreshInterval,
		maxAge:          maxAge,
		stopCh:          make(chan struct{}),
	}
}

// Start begins the background import scheduler.
func (s *ImportScheduler) Start(ctx context.Context) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		// Initial check on startup
		s.checkAndRefresh(ctx)

		ticker := time.NewTicker(s.refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkAndRefresh(ctx)
			case <-s.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop gracefully stops the scheduler.
func (s *ImportScheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

// ForceRefresh triggers an immediate import.
func (s *ImportScheduler) ForceRefresh(ctx context.Context) {
	s.importAllStations(ctx)
}

func (s *ImportScheduler) checkAndRefresh(ctx context.Context) {
	shouldRefresh, err := s.metaRepo.ShouldRefresh(ctx, "stations", s.maxAge)
	if err != nil {
		log.Printf("failed to check refresh status: %v", err)
		return
	}

	if shouldRefresh {
		log.Println("starting scheduled station import...")
		s.importAllStations(ctx)
	}
}

func (s *ImportScheduler) importAllStations(ctx context.Context) {
	start := time.Now()
	meta := &ImportMetadata{
		EntityType: "stations",
		LastImport: start,
		Status:     "in_progress",
	}

	var totalCount int
	var failedRegions int

	for region := 1; region <= 22; region++ {
		stations, err := s.vtClient.ElencoStazioni(ctx, region)
		if err != nil {
			log.Printf("failed to fetch region %d: %v", region, err)
			failedRegions++
			continue
		}

		now := time.Now()
		for _, rs := range stations {
			station := &Station{
				ID:        rs.ID,
				Name:      rs.Name,
				Region:    rs.Region,
				Latitude:  rs.Latitude,
				Longitude: rs.Longitude,
				UpdatedAt: now,
			}
			if err := s.stationRepo.Upsert(ctx, station); err != nil {
				log.Printf("failed to upsert station %s: %v", rs.ID, err)
			} else {
				totalCount++
			}
		}
		log.Printf("imported %d stations from region %d", len(stations), region)
	}

	meta.RecordCount = totalCount
	meta.DurationMs = time.Since(start).Milliseconds()
	if failedRegions > 0 {
		meta.Status = "partial_failure"
		meta.ErrorMessage = "some regions failed to import"
	} else {
		meta.Status = "success"
	}

	if err := s.metaRepo.Upsert(ctx, meta); err != nil {
		log.Printf("failed to save import metadata: %v", err)
	}

	log.Printf("import complete: %d stations in %dms (status: %s)", totalCount, meta.DurationMs, meta.Status)
}

// GetLastImportStatus returns the last import metadata.
func (s *ImportScheduler) GetLastImportStatus(ctx context.Context) (*ImportMetadata, error) {
	return s.metaRepo.Get(ctx, "stations")
}
