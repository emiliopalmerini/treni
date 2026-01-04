-- name: CreateObservation :exec
INSERT INTO train_observation (id, observed_at, station_id, station_name, observation_type,
    train_number, train_category, origin_id, origin_name, destination_id, destination_name,
    scheduled_time, delay, platform, circulation_state)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

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
    origin_name,
    destination_name,
    COUNT(*) as observation_count,
    COALESCE(AVG(delay), 0) as average_delay,
    COALESCE(MAX(delay), 0) as max_delay,
    SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_percentage
FROM train_observation
WHERE train_number = ?
GROUP BY train_number;

-- name: GetWorstTrains :many
SELECT
    train_number,
    train_category as category,
    origin_name,
    destination_name,
    COUNT(*) as observation_count,
    COALESCE(AVG(delay), 0) as average_delay,
    COALESCE(MAX(delay), 0) as max_delay,
    SUM(CASE WHEN delay = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_percentage
FROM train_observation
WHERE circulation_state != 1
GROUP BY train_number
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
