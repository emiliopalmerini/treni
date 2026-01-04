package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/journey"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, entity *journey.Journey) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO journey (id, train_number, origin_id, origin_name, destination_id, destination_name,
		 scheduled_departure, actual_departure, delay, recorded_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entity.ID.String(), entity.TrainNumber, entity.OriginID, entity.OriginName,
		entity.DestinationID, entity.DestinationName, entity.ScheduledDeparture,
		entity.ActualDeparture, entity.Delay, entity.RecordedAt)
	return err
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*journey.Journey, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, train_number, origin_id, origin_name, destination_id, destination_name,
		 scheduled_departure, actual_departure, delay, recorded_at
		 FROM journey WHERE id = ?`, id.String())

	return scanJourney(row)
}

func (r *SQLiteRepository) List(ctx context.Context) ([]*journey.Journey, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, train_number, origin_id, origin_name, destination_id, destination_name,
		 scheduled_departure, actual_departure, delay, recorded_at
		 FROM journey ORDER BY recorded_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanJourneys(rows)
}

func (r *SQLiteRepository) ListByTrain(ctx context.Context, trainNumber int) ([]*journey.Journey, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, train_number, origin_id, origin_name, destination_id, destination_name,
		 scheduled_departure, actual_departure, delay, recorded_at
		 FROM journey WHERE train_number = ? ORDER BY recorded_at DESC`, trainNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanJourneys(rows)
}

func (r *SQLiteRepository) ListByDateRange(ctx context.Context, from, to time.Time) ([]*journey.Journey, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, train_number, origin_id, origin_name, destination_id, destination_name,
		 scheduled_departure, actual_departure, delay, recorded_at
		 FROM journey WHERE recorded_at BETWEEN ? AND ? ORDER BY recorded_at DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanJourneys(rows)
}

func (r *SQLiteRepository) Update(ctx context.Context, entity *journey.Journey) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE journey SET train_number = ?, origin_id = ?, origin_name = ?, destination_id = ?,
		 destination_name = ?, scheduled_departure = ?, actual_departure = ?, delay = ?
		 WHERE id = ?`,
		entity.TrainNumber, entity.OriginID, entity.OriginName, entity.DestinationID,
		entity.DestinationName, entity.ScheduledDeparture, entity.ActualDeparture,
		entity.Delay, entity.ID.String())
	return err
}

func (r *SQLiteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM journey WHERE id = ?", id.String())
	return err
}

func (r *SQLiteRepository) CreateStop(ctx context.Context, stop *journey.JourneyStop) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO journey_stop (id, journey_id, station_id, station_name, scheduled_arrival,
		 scheduled_departure, actual_arrival, actual_departure, arrival_delay, departure_delay, platform)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		stop.ID.String(), stop.JourneyID.String(), stop.StationID, stop.StationName,
		stop.ScheduledArrival, stop.ScheduledDeparture, stop.ActualArrival, stop.ActualDeparture,
		stop.ArrivalDelay, stop.DepartureDelay, stop.Platform)
	return err
}

func (r *SQLiteRepository) GetStopsByJourney(ctx context.Context, journeyID uuid.UUID) ([]*journey.JourneyStop, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, journey_id, station_id, station_name, scheduled_arrival, scheduled_departure,
		 actual_arrival, actual_departure, arrival_delay, departure_delay, platform
		 FROM journey_stop WHERE journey_id = ? ORDER BY scheduled_departure`, journeyID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stops []*journey.JourneyStop
	for rows.Next() {
		var stop journey.JourneyStop
		var idStr, journeyIDStr string
		var schedArr, schedDep, actArr, actDep sql.NullTime
		if err := rows.Scan(&idStr, &journeyIDStr, &stop.StationID, &stop.StationName,
			&schedArr, &schedDep, &actArr, &actDep,
			&stop.ArrivalDelay, &stop.DepartureDelay, &stop.Platform); err != nil {
			return nil, err
		}
		stop.ID, _ = uuid.Parse(idStr)
		stop.JourneyID, _ = uuid.Parse(journeyIDStr)
		if schedArr.Valid {
			stop.ScheduledArrival = schedArr.Time
		}
		if schedDep.Valid {
			stop.ScheduledDeparture = schedDep.Time
		}
		if actArr.Valid {
			stop.ActualArrival = actArr.Time
		}
		if actDep.Valid {
			stop.ActualDeparture = actDep.Time
		}
		stops = append(stops, &stop)
	}
	return stops, rows.Err()
}

func scanJourney(row *sql.Row) (*journey.Journey, error) {
	var entity journey.Journey
	var idStr string
	var actualDep sql.NullTime
	if err := row.Scan(&idStr, &entity.TrainNumber, &entity.OriginID, &entity.OriginName,
		&entity.DestinationID, &entity.DestinationName, &entity.ScheduledDeparture,
		&actualDep, &entity.Delay, &entity.RecordedAt); err != nil {
		return nil, err
	}
	entity.ID, _ = uuid.Parse(idStr)
	if actualDep.Valid {
		entity.ActualDeparture = actualDep.Time
	}
	return &entity, nil
}

func scanJourneys(rows *sql.Rows) ([]*journey.Journey, error) {
	var entities []*journey.Journey
	for rows.Next() {
		var entity journey.Journey
		var idStr string
		var actualDep sql.NullTime
		if err := rows.Scan(&idStr, &entity.TrainNumber, &entity.OriginID, &entity.OriginName,
			&entity.DestinationID, &entity.DestinationName, &entity.ScheduledDeparture,
			&actualDep, &entity.Delay, &entity.RecordedAt); err != nil {
			return nil, err
		}
		entity.ID, _ = uuid.Parse(idStr)
		if actualDep.Valid {
			entity.ActualDeparture = actualDep.Time
		}
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}
