package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/treni/internal/observation"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, entity *observation.TrainObservation) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO train_observation (id, observed_at, station_id, station_name, observation_type,
		 train_number, train_category, origin_id, origin_name, destination_id, destination_name,
		 scheduled_time, delay, platform, circulation_state)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entity.ID.String(), entity.ObservedAt, entity.StationID, entity.StationName, entity.ObservationType,
		entity.TrainNumber, entity.TrainCategory, entity.OriginID, entity.OriginName,
		entity.DestinationID, entity.DestinationName, entity.ScheduledTime, entity.Delay,
		entity.Platform, entity.CirculationState)
	return err
}

func (r *SQLiteRepository) CreateBatch(ctx context.Context, entities []*observation.TrainObservation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO train_observation (id, observed_at, station_id, station_name, observation_type,
		 train_number, train_category, origin_id, origin_name, destination_id, destination_name,
		 scheduled_time, delay, platform, circulation_state)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, entity := range entities {
		_, err := stmt.ExecContext(ctx,
			entity.ID.String(), entity.ObservedAt, entity.StationID, entity.StationName, entity.ObservationType,
			entity.TrainNumber, entity.TrainCategory, entity.OriginID, entity.OriginName,
			entity.DestinationID, entity.DestinationName, entity.ScheduledTime, entity.Delay,
			entity.Platform, entity.CirculationState)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SQLiteRepository) GetGlobalStats(ctx context.Context) (*observation.GlobalStats, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT
			COUNT(*) as total,
			COALESCE(AVG(delay), 0) as avg_delay,
			SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) as on_time,
			SUM(CASE WHEN circulation_state = 1 THEN 1 ELSE 0 END) as cancelled
		FROM train_observation`)

	var stats observation.GlobalStats
	if err := row.Scan(&stats.TotalObservations, &stats.AverageDelay, &stats.OnTimeCount, &stats.CancelledCount); err != nil {
		return nil, err
	}

	if stats.TotalObservations > 0 {
		stats.OnTimePercentage = float64(stats.OnTimeCount) / float64(stats.TotalObservations) * 100
	}

	return &stats, nil
}

func (r *SQLiteRepository) GetStatsByCategory(ctx context.Context) ([]*observation.CategoryStats, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			train_category,
			COUNT(*) as count,
			COALESCE(AVG(delay), 0) as avg_delay,
			SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_pct
		FROM train_observation
		WHERE train_category != ''
		GROUP BY train_category
		ORDER BY count DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*observation.CategoryStats
	for rows.Next() {
		var s observation.CategoryStats
		if err := rows.Scan(&s.Category, &s.ObservationCount, &s.AverageDelay, &s.OnTimePercentage); err != nil {
			return nil, err
		}
		stats = append(stats, &s)
	}
	return stats, rows.Err()
}

func (r *SQLiteRepository) GetStatsByStation(ctx context.Context, stationID string) (*observation.StationStats, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT
			station_id,
			station_name,
			COUNT(*) as count,
			COALESCE(AVG(delay), 0) as avg_delay,
			SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_pct
		FROM train_observation
		WHERE station_id = ?
		GROUP BY station_id`, stationID)

	var s observation.StationStats
	if err := row.Scan(&s.StationID, &s.StationName, &s.ObservationCount, &s.AverageDelay, &s.OnTimePercentage); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SQLiteRepository) GetStatsByTrain(ctx context.Context, trainNumber int) (*observation.TrainStats, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT
			train_number,
			train_category,
			origin_name,
			destination_name,
			COUNT(*) as count,
			COALESCE(AVG(delay), 0) as avg_delay,
			COALESCE(MAX(delay), 0) as max_delay,
			SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_pct
		FROM train_observation
		WHERE train_number = ?
		GROUP BY train_number`, trainNumber)

	var s observation.TrainStats
	var category, origin, destination sql.NullString
	if err := row.Scan(&s.TrainNumber, &category, &origin, &destination,
		&s.ObservationCount, &s.AverageDelay, &s.MaxDelay, &s.OnTimePercentage); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	s.Category = category.String
	s.OriginName = origin.String
	s.DestinationName = destination.String
	return &s, nil
}

func (r *SQLiteRepository) GetWorstTrains(ctx context.Context, limit int) ([]*observation.TrainStats, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			train_number,
			train_category,
			origin_name,
			destination_name,
			COUNT(*) as count,
			COALESCE(AVG(delay), 0) as avg_delay,
			COALESCE(MAX(delay), 0) as max_delay,
			SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_pct
		FROM train_observation
		WHERE circulation_state != 1
		GROUP BY train_number
		HAVING count >= 3
		ORDER BY avg_delay DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*observation.TrainStats
	for rows.Next() {
		var s observation.TrainStats
		var category, origin, destination sql.NullString
		if err := rows.Scan(&s.TrainNumber, &category, &origin, &destination,
			&s.ObservationCount, &s.AverageDelay, &s.MaxDelay, &s.OnTimePercentage); err != nil {
			return nil, err
		}
		s.Category = category.String
		s.OriginName = origin.String
		s.DestinationName = destination.String
		stats = append(stats, &s)
	}
	return stats, rows.Err()
}

func (r *SQLiteRepository) GetWorstStations(ctx context.Context, limit int) ([]*observation.StationStats, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			station_id,
			station_name,
			COUNT(*) as count,
			COALESCE(AVG(delay), 0) as avg_delay,
			SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_pct
		FROM train_observation
		WHERE circulation_state != 1
		GROUP BY station_id
		HAVING count >= 3
		ORDER BY avg_delay DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*observation.StationStats
	for rows.Next() {
		var s observation.StationStats
		if err := rows.Scan(&s.StationID, &s.StationName, &s.ObservationCount, &s.AverageDelay, &s.OnTimePercentage); err != nil {
			return nil, err
		}
		stats = append(stats, &s)
	}
	return stats, rows.Err()
}

func (r *SQLiteRepository) GetRecentObservations(ctx context.Context, limit int) ([]*observation.TrainObservation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, observed_at, station_id, station_name, observation_type,
		 train_number, train_category, origin_id, origin_name, destination_id, destination_name,
		 scheduled_time, delay, platform, circulation_state
		 FROM train_observation
		 ORDER BY observed_at DESC
		 LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanObservations(rows)
}

func (r *SQLiteRepository) GetRecentByStation(ctx context.Context, stationID string, limit int) ([]*observation.TrainObservation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, observed_at, station_id, station_name, observation_type,
		 train_number, train_category, origin_id, origin_name, destination_id, destination_name,
		 scheduled_time, delay, platform, circulation_state
		 FROM train_observation
		 WHERE station_id = ?
		 ORDER BY observed_at DESC
		 LIMIT ?`, stationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanObservations(rows)
}

func scanObservations(rows *sql.Rows) ([]*observation.TrainObservation, error) {
	var entities []*observation.TrainObservation
	for rows.Next() {
		var entity observation.TrainObservation
		var idStr string
		var obsType string
		var category, originID, originName, destID, destName, platform sql.NullString
		var scheduledTime sql.NullTime
		if err := rows.Scan(&idStr, &entity.ObservedAt, &entity.StationID, &entity.StationName, &obsType,
			&entity.TrainNumber, &category, &originID, &originName, &destID, &destName,
			&scheduledTime, &entity.Delay, &platform, &entity.CirculationState); err != nil {
			return nil, err
		}
		entity.ID, _ = uuid.Parse(idStr)
		entity.ObservationType = observation.ObservationType(obsType)
		entity.TrainCategory = category.String
		entity.OriginID = originID.String
		entity.OriginName = originName.String
		entity.DestinationID = destID.String
		entity.DestinationName = destName.String
		entity.Platform = platform.String
		if scheduledTime.Valid {
			entity.ScheduledTime = scheduledTime.Time
		}
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}
