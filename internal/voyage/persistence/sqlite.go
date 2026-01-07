package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/database/sqlc"
	"github.com/emiliopalmerini/treni/internal/voyage"
)

type SQLiteRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db, q: sqlc.New(db)}
}

func (r *SQLiteRepository) Create(ctx context.Context, v *voyage.Voyage) error {
	return r.q.CreateVoyage(ctx, sqlc.CreateVoyageParams{
		ID:                 v.ID.String(),
		TrainNumber:        int64(v.TrainNumber),
		TrainCategory:      strPtr(v.TrainCategory),
		OriginID:           v.OriginID,
		OriginName:         v.OriginName,
		DestinationID:      v.DestinationID,
		DestinationName:    v.DestinationName,
		ScheduledDate:      v.ScheduledDate,
		ScheduledDeparture: v.ScheduledDeparture,
		CirculationState:   int64Ptr(v.CirculationState),
		CreatedAt:          v.CreatedAt,
		UpdatedAt:          v.UpdatedAt,
	})
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*voyage.Voyage, error) {
	row, err := r.q.GetVoyageByID(ctx, id.String())
	if err != nil {
		return nil, err
	}
	return rowToVoyage(row), nil
}

func (r *SQLiteRepository) FindByKey(ctx context.Context, trainNumber int, originID string, scheduledDate string) (*voyage.Voyage, error) {
	row, err := r.q.FindVoyageByKey(ctx, sqlc.FindVoyageByKeyParams{
		TrainNumber:   int64(trainNumber),
		OriginID:      originID,
		ScheduledDate: scheduledDate,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return rowToVoyage(row), nil
}

func (r *SQLiteRepository) Update(ctx context.Context, v *voyage.Voyage) error {
	return r.q.UpdateVoyage(ctx, sqlc.UpdateVoyageParams{
		TrainCategory:    strPtr(v.TrainCategory),
		CirculationState: int64Ptr(v.CirculationState),
		UpdatedAt:        v.UpdatedAt,
		ID:               v.ID.String(),
	})
}

func (r *SQLiteRepository) CreateStop(ctx context.Context, stop *voyage.VoyageStop) error {
	return r.q.CreateVoyageStop(ctx, stopToCreateParams(stop))
}

func (r *SQLiteRepository) CreateStopsBatch(ctx context.Context, stops []*voyage.VoyageStop) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := r.q.WithTx(tx)
	for _, stop := range stops {
		if err := q.CreateVoyageStop(ctx, stopToCreateParams(stop)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SQLiteRepository) UpdateStop(ctx context.Context, stop *voyage.VoyageStop) error {
	return r.q.UpdateVoyageStop(ctx, sqlc.UpdateVoyageStopParams{
		ActualArrival:     timePtr(stop.ActualArrival),
		ActualDeparture:   timePtr(stop.ActualDeparture),
		ArrivalDelay:      int64Ptr(stop.ArrivalDelay),
		DepartureDelay:    int64Ptr(stop.DepartureDelay),
		Platform:          strPtr(stop.Platform),
		IsSuppressed:      boolToInt64Ptr(stop.IsSuppressed),
		LastObservationAt: timePtr(stop.LastObservationAt),
		ID:                stop.ID.String(),
	})
}

func (r *SQLiteRepository) GetStopsByVoyage(ctx context.Context, voyageID uuid.UUID) ([]*voyage.VoyageStop, error) {
	rows, err := r.q.GetVoyageStops(ctx, voyageID.String())
	if err != nil {
		return nil, err
	}

	stops := make([]*voyage.VoyageStop, len(rows))
	for i, row := range rows {
		stops[i] = rowToVoyageStop(row)
	}
	return stops, nil
}

func (r *SQLiteRepository) FindStopByStation(ctx context.Context, voyageID uuid.UUID, stationID string) (*voyage.VoyageStop, error) {
	row, err := r.q.FindVoyageStopByStation(ctx, sqlc.FindVoyageStopByStationParams{
		VoyageID:  voyageID.String(),
		StationID: stationID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return rowToVoyageStop(row), nil
}

func (r *SQLiteRepository) GetVoyageWithStops(ctx context.Context, voyageID uuid.UUID) (*voyage.VoyageWithStops, error) {
	v, err := r.GetByID(ctx, voyageID)
	if err != nil {
		return nil, err
	}

	stops, err := r.GetStopsByVoyage(ctx, voyageID)
	if err != nil {
		return nil, err
	}

	return &voyage.VoyageWithStops{
		Voyage: *v,
		Stops:  convertStopsToValue(stops),
	}, nil
}

func (r *SQLiteRepository) GetVoyagesByTrain(ctx context.Context, trainNumber int, limit int) ([]*voyage.Voyage, error) {
	rows, err := r.q.GetVoyagesByTrain(ctx, sqlc.GetVoyagesByTrainParams{
		TrainNumber: int64(trainNumber),
		Limit:       int64(limit),
	})
	if err != nil {
		return nil, err
	}

	voyages := make([]*voyage.Voyage, len(rows))
	for i, row := range rows {
		voyages[i] = rowToVoyage(row)
	}
	return voyages, nil
}

func (r *SQLiteRepository) GetVoyagesByDate(ctx context.Context, date string, limit int) ([]*voyage.Voyage, error) {
	rows, err := r.q.GetVoyagesByDate(ctx, sqlc.GetVoyagesByDateParams{
		ScheduledDate: date,
		Limit:         int64(limit),
	})
	if err != nil {
		return nil, err
	}

	voyages := make([]*voyage.Voyage, len(rows))
	for i, row := range rows {
		voyages[i] = rowToVoyage(row)
	}
	return voyages, nil
}

func (r *SQLiteRepository) GetRecentVoyages(ctx context.Context, limit int) ([]*voyage.Voyage, error) {
	rows, err := r.q.GetRecentVoyages(ctx, int64(limit))
	if err != nil {
		return nil, err
	}

	voyages := make([]*voyage.Voyage, len(rows))
	for i, row := range rows {
		voyages[i] = rowToVoyage(row)
	}
	return voyages, nil
}

// Helper functions

func rowToVoyage(row sqlc.Voyage) *voyage.Voyage {
	return &voyage.Voyage{
		ID:                 uuid.MustParse(row.ID),
		TrainNumber:        int(row.TrainNumber),
		TrainCategory:      derefStr(row.TrainCategory),
		OriginID:           row.OriginID,
		OriginName:         row.OriginName,
		DestinationID:      row.DestinationID,
		DestinationName:    row.DestinationName,
		ScheduledDate:      row.ScheduledDate,
		ScheduledDeparture: row.ScheduledDeparture,
		CirculationState:   derefInt64(row.CirculationState),
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

func rowToVoyageStop(row sqlc.VoyageStop) *voyage.VoyageStop {
	return &voyage.VoyageStop{
		ID:                 uuid.MustParse(row.ID),
		VoyageID:           uuid.MustParse(row.VoyageID),
		StationID:          row.StationID,
		StationName:        row.StationName,
		StopSequence:       int(row.StopSequence),
		StopType:           derefStr(row.StopType),
		ScheduledArrival:   row.ScheduledArrival,
		ScheduledDeparture: row.ScheduledDeparture,
		ActualArrival:      row.ActualArrival,
		ActualDeparture:    row.ActualDeparture,
		ArrivalDelay:       derefInt64(row.ArrivalDelay),
		DepartureDelay:     derefInt64(row.DepartureDelay),
		Platform:           derefStr(row.Platform),
		IsSuppressed:       derefInt64Bool(row.IsSuppressed),
		LastObservationAt:  row.LastObservationAt,
	}
}

func stopToCreateParams(stop *voyage.VoyageStop) sqlc.CreateVoyageStopParams {
	return sqlc.CreateVoyageStopParams{
		ID:                 stop.ID.String(),
		VoyageID:           stop.VoyageID.String(),
		StationID:          stop.StationID,
		StationName:        stop.StationName,
		StopSequence:       int64(stop.StopSequence),
		StopType:           strPtr(stop.StopType),
		ScheduledArrival:   timePtr(stop.ScheduledArrival),
		ScheduledDeparture: timePtr(stop.ScheduledDeparture),
		ActualArrival:      timePtr(stop.ActualArrival),
		ActualDeparture:    timePtr(stop.ActualDeparture),
		ArrivalDelay:       int64Ptr(stop.ArrivalDelay),
		DepartureDelay:     int64Ptr(stop.DepartureDelay),
		Platform:           strPtr(stop.Platform),
		IsSuppressed:       boolToInt64Ptr(stop.IsSuppressed),
		LastObservationAt:  timePtr(stop.LastObservationAt),
	}
}

func convertStopsToValue(stops []*voyage.VoyageStop) []voyage.VoyageStop {
	result := make([]voyage.VoyageStop, len(stops))
	for i, stop := range stops {
		result[i] = *stop
	}
	return result
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func timePtr(t *time.Time) *time.Time {
	return t
}

func int64Ptr(i int) *int64 {
	v := int64(i)
	return &v
}

func derefInt64(i *int64) int {
	if i == nil {
		return 0
	}
	return int(*i)
}

func boolToInt64Ptr(b bool) *int64 {
	v := int64(0)
	if b {
		v = 1
	}
	return &v
}

func derefInt64Bool(i *int64) bool {
	if i == nil {
		return false
	}
	return *i != 0
}
