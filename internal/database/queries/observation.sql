-- name: UpsertObservation :one
INSERT INTO train_observation (id, observed_at, station_id, station_name, observation_type,
    train_number, train_category, origin_id, origin_name, destination_id, destination_name,
    scheduled_time, scheduled_date, delay, platform, circulation_state)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(train_number, station_id, observation_type, scheduled_date) DO UPDATE SET
    observed_at = excluded.observed_at,
    train_category = excluded.train_category,
    origin_id = excluded.origin_id,
    origin_name = excluded.origin_name,
    destination_id = excluded.destination_id,
    destination_name = excluded.destination_name,
    delay = excluded.delay,
    platform = excluded.platform,
    circulation_state = excluded.circulation_state
RETURNING id, delay;

-- name: GetObservationByKey :one
SELECT id, delay FROM train_observation
WHERE train_number = ? AND station_id = ? AND observation_type = ? AND scheduled_date = ?;

-- name: CreateDelayVariation :exec
INSERT INTO delay_variation (id, observation_id, recorded_at, delay)
VALUES (?, ?, ?, ?);

-- name: GetDelayVariationsByObservation :many
SELECT id, observation_id, recorded_at, delay
FROM delay_variation
WHERE observation_id = ?
ORDER BY recorded_at ASC;

-- name: GetLatestDelayVariation :one
SELECT id, observation_id, recorded_at, delay
FROM delay_variation
WHERE observation_id = ?
ORDER BY recorded_at DESC
LIMIT 1;

-- name: GetGlobalStats :one
SELECT
    COUNT(*) as total_observations,
    COALESCE(AVG(delay), 0) as average_delay,
    SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) as on_time_count,
    SUM(CASE WHEN circulation_state = 1 THEN 1 ELSE 0 END) as cancelled_count
FROM train_observation;

-- name: GetStatsByCategory :many
SELECT
    train_category as category,
    COUNT(*) as observation_count,
    COALESCE(AVG(delay), 0) as average_delay,
    SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_percentage
FROM train_observation
WHERE train_category != ''
GROUP BY train_category
ORDER BY observation_count DESC;

-- name: GetStatsByStation :one
SELECT
    station_id,
    station_name,
    COUNT(*) as observation_count,
    COALESCE(AVG(delay), 0) as average_delay,
    SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_percentage
FROM train_observation
WHERE station_id = ?
GROUP BY station_id;

-- name: GetStatsByTrain :one
SELECT
    train_number,
    train_category as category,
    origin_id,
    origin_name,
    destination_id,
    destination_name,
    COUNT(*) as observation_count,
    COALESCE(AVG(delay), 0) as average_delay,
    COALESCE(MAX(delay), 0) as max_delay,
    SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_percentage
FROM train_observation
WHERE train_number = ?
GROUP BY train_number, origin_id, destination_id;

-- name: GetWorstTrains :many
SELECT
    train_number,
    train_category as category,
    origin_id,
    origin_name,
    destination_id,
    destination_name,
    COUNT(*) as observation_count,
    COALESCE(AVG(delay), 0) as average_delay,
    COALESCE(MAX(delay), 0) as max_delay,
    SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_percentage
FROM train_observation
WHERE circulation_state != 1
GROUP BY train_number, origin_id, destination_id
HAVING observation_count >= 3
ORDER BY average_delay DESC
LIMIT ?;

-- name: GetWorstStations :many
SELECT
    station_id,
    station_name,
    COUNT(*) as observation_count,
    COALESCE(AVG(delay), 0) as average_delay,
    SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_percentage
FROM train_observation
WHERE circulation_state != 1
GROUP BY station_id
HAVING observation_count >= 3
ORDER BY average_delay DESC
LIMIT ?;

-- name: GetRecentObservations :many
SELECT id, observed_at, station_id, station_name, observation_type,
    train_number, train_category, origin_id, origin_name, destination_id, destination_name,
    scheduled_time, delay, platform, circulation_state
FROM train_observation
ORDER BY observed_at DESC
LIMIT ?;

-- name: GetRecentObservationsByStation :many
SELECT id, observed_at, station_id, station_name, observation_type,
    train_number, train_category, origin_id, origin_name, destination_id, destination_name,
    scheduled_time, delay, platform, circulation_state
FROM train_observation
WHERE station_id = ?
ORDER BY observed_at DESC
LIMIT ?;
