package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/database/sqlc"
	"github.com/emiliopalmerini/treni/internal/journey"
)

type SQLiteRepository struct {
	q *sqlc.Queries
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{q: sqlc.New(db)}
}

func (r *SQLiteRepository) Create(ctx context.Context, entity *journey.Journey) error {
	return r.q.CreateJourney(ctx, sqlc.CreateJourneyParams{
		ID:                 entity.ID.String(),
		TrainNumber:        int64(entity.TrainNumber),
		OriginID:           entity.OriginID,
		OriginName:         entity.OriginName,
		DestinationID:      entity.DestinationID,
		DestinationName:    entity.DestinationName,
		ScheduledDeparture: &entity.ScheduledDeparture,
		ActualDeparture:    timePtr(entity.ActualDeparture),
		Delay:              ptr(int64(entity.Delay)),
		RecordedAt:         entity.RecordedAt,
	})
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*journey.Journey, error) {
	row, err := r.q.GetJourneyByID(ctx, id.String())
	if err != nil {
		return nil, err
	}
	return sqlcJourneyToJourney(row), nil
}

func (r *SQLiteRepository) List(ctx context.Context) ([]*journey.Journey, error) {
	rows, err := r.q.ListJourneys(ctx)
	if err != nil {
		return nil, err
	}
	return sqlcJourneysToJourneys(rows), nil
}

func (r *SQLiteRepository) ListByTrain(ctx context.Context, trainNumber int) ([]*journey.Journey, error) {
	rows, err := r.q.ListJourneysByTrain(ctx, int64(trainNumber))
	if err != nil {
		return nil, err
	}
	return sqlcJourneysToJourneys(rows), nil
}

func (r *SQLiteRepository) ListByDateRange(ctx context.Context, from, to time.Time) ([]*journey.Journey, error) {
	rows, err := r.q.ListJourneysByDateRange(ctx, sqlc.ListJourneysByDateRangeParams{
		FromRecordedAt: from,
		ToRecordedAt:   to,
	})
	if err != nil {
		return nil, err
	}
	return sqlcJourneysToJourneys(rows), nil
}

func (r *SQLiteRepository) Update(ctx context.Context, entity *journey.Journey) error {
	return r.q.UpdateJourney(ctx, sqlc.UpdateJourneyParams{
		ID:                 entity.ID.String(),
		TrainNumber:        int64(entity.TrainNumber),
		OriginID:           entity.OriginID,
		OriginName:         entity.OriginName,
		DestinationID:      entity.DestinationID,
		DestinationName:    entity.DestinationName,
		ScheduledDeparture: &entity.ScheduledDeparture,
		ActualDeparture:    timePtr(entity.ActualDeparture),
		Delay:              ptr(int64(entity.Delay)),
	})
}

func (r *SQLiteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteJourney(ctx, id.String())
}

func (r *SQLiteRepository) CreateStop(ctx context.Context, stop *journey.JourneyStop) error {
	return r.q.CreateJourneyStop(ctx, sqlc.CreateJourneyStopParams{
		ID:                 stop.ID.String(),
		JourneyID:          stop.JourneyID.String(),
		StationID:          stop.StationID,
		StationName:        stop.StationName,
		ScheduledArrival:   timePtr(stop.ScheduledArrival),
		ScheduledDeparture: timePtr(stop.ScheduledDeparture),
		ActualArrival:      timePtr(stop.ActualArrival),
		ActualDeparture:    timePtr(stop.ActualDeparture),
		ArrivalDelay:       ptr(int64(stop.ArrivalDelay)),
		DepartureDelay:     ptr(int64(stop.DepartureDelay)),
		Platform:           strPtr(stop.Platform),
	})
}

func (r *SQLiteRepository) GetStopsByJourney(ctx context.Context, journeyID uuid.UUID) ([]*journey.JourneyStop, error) {
	rows, err := r.q.GetJourneyStops(ctx, journeyID.String())
	if err != nil {
		return nil, err
	}
	return sqlcStopsToStops(rows), nil
}

func ptr[T any](v T) *T {
	return &v
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func sqlcJourneyToJourney(row sqlc.Journey) *journey.Journey {
	id, _ := uuid.Parse(row.ID)
	return &journey.Journey{
		ID:                 id,
		TrainNumber:        int(row.TrainNumber),
		OriginID:           row.OriginID,
		OriginName:         row.OriginName,
		DestinationID:      row.DestinationID,
		DestinationName:    row.DestinationName,
		ScheduledDeparture: deref(row.ScheduledDeparture),
		ActualDeparture:    deref(row.ActualDeparture),
		Delay:              int(deref(row.Delay)),
		RecordedAt:         row.RecordedAt,
	}
}

func sqlcJourneysToJourneys(rows []sqlc.Journey) []*journey.Journey {
	journeys := make([]*journey.Journey, len(rows))
	for i, row := range rows {
		journeys[i] = sqlcJourneyToJourney(row)
	}
	return journeys
}

func sqlcStopsToStops(rows []sqlc.JourneyStop) []*journey.JourneyStop {
	stops := make([]*journey.JourneyStop, len(rows))
	for i, row := range rows {
		id, _ := uuid.Parse(row.ID)
		journeyID, _ := uuid.Parse(row.JourneyID)
		stops[i] = &journey.JourneyStop{
			ID:                 id,
			JourneyID:          journeyID,
			StationID:          row.StationID,
			StationName:        row.StationName,
			ScheduledArrival:   deref(row.ScheduledArrival),
			ScheduledDeparture: deref(row.ScheduledDeparture),
			ActualArrival:      deref(row.ActualArrival),
			ActualDeparture:    deref(row.ActualDeparture),
			ArrivalDelay:       int(deref(row.ArrivalDelay)),
			DepartureDelay:     int(deref(row.DepartureDelay)),
			Platform:           deref(row.Platform),
		}
	}
	return stops
}
