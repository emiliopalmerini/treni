-- Voyage CRUD operations

-- name: CreateVoyage :exec
INSERT INTO voyage (id, train_number, train_category, origin_id, origin_name,
    destination_id, destination_name, scheduled_date, scheduled_departure,
    circulation_state, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetVoyageByID :one
SELECT id, train_number, train_category, origin_id, origin_name,
    destination_id, destination_name, scheduled_date, scheduled_departure,
    circulation_state, created_at, updated_at
FROM voyage
WHERE id = ?;

-- name: FindVoyageByKey :one
SELECT id, train_number, train_category, origin_id, origin_name,
    destination_id, destination_name, scheduled_date, scheduled_departure,
    circulation_state, created_at, updated_at
FROM voyage
WHERE train_number = ? AND origin_id = ? AND scheduled_date = ?
LIMIT 1;

-- name: UpdateVoyage :exec
UPDATE voyage
SET train_category = ?,
    circulation_state = ?,
    updated_at = ?
WHERE id = ?;

-- name: GetVoyagesByTrain :many
SELECT id, train_number, train_category, origin_id, origin_name,
    destination_id, destination_name, scheduled_date, scheduled_departure,
    circulation_state, created_at, updated_at
FROM voyage
WHERE train_number = ?
ORDER BY scheduled_date DESC, scheduled_departure DESC
LIMIT ?;

-- name: GetVoyagesByDate :many
SELECT id, train_number, train_category, origin_id, origin_name,
    destination_id, destination_name, scheduled_date, scheduled_departure,
    circulation_state, created_at, updated_at
FROM voyage
WHERE scheduled_date = ?
ORDER BY scheduled_departure ASC
LIMIT ?;

-- name: GetRecentVoyages :many
SELECT id, train_number, train_category, origin_id, origin_name,
    destination_id, destination_name, scheduled_date, scheduled_departure,
    circulation_state, created_at, updated_at
FROM voyage
ORDER BY updated_at DESC
LIMIT ?;

-- Voyage Stop CRUD operations

-- name: CreateVoyageStop :exec
INSERT INTO voyage_stop (id, voyage_id, station_id, station_name, stop_sequence,
    stop_type, scheduled_arrival, scheduled_departure, actual_arrival, actual_departure,
    arrival_delay, departure_delay, platform, is_suppressed, last_observation_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateVoyageStop :exec
UPDATE voyage_stop
SET actual_arrival = ?,
    actual_departure = ?,
    arrival_delay = ?,
    departure_delay = ?,
    platform = ?,
    is_suppressed = ?,
    last_observation_at = ?
WHERE id = ?;

-- name: GetVoyageStops :many
SELECT id, voyage_id, station_id, station_name, stop_sequence,
    stop_type, scheduled_arrival, scheduled_departure, actual_arrival, actual_departure,
    arrival_delay, departure_delay, platform, is_suppressed, last_observation_at
FROM voyage_stop
WHERE voyage_id = ?
ORDER BY stop_sequence ASC;

-- name: FindVoyageStopByStation :one
SELECT id, voyage_id, station_id, station_name, stop_sequence,
    stop_type, scheduled_arrival, scheduled_departure, actual_arrival, actual_departure,
    arrival_delay, departure_delay, platform, is_suppressed, last_observation_at
FROM voyage_stop
WHERE voyage_id = ? AND station_id = ?
LIMIT 1;

-- Voyage statistics and queries

-- name: GetVoyageDelayProgression :many
SELECT stop_sequence, station_name, scheduled_arrival, scheduled_departure,
    actual_arrival, actual_departure, arrival_delay, departure_delay, is_suppressed
FROM voyage_stop
WHERE voyage_id = ?
ORDER BY stop_sequence ASC;

-- name: GetStationStopsFromVoyages :many
SELECT
    v.id as voyage_id,
    v.train_number,
    v.train_category,
    v.origin_name,
    v.destination_name,
    v.scheduled_date,
    vs.arrival_delay,
    vs.departure_delay,
    vs.scheduled_arrival,
    vs.scheduled_departure,
    vs.actual_arrival,
    vs.actual_departure,
    vs.is_suppressed
FROM voyage_stop vs
JOIN voyage v ON vs.voyage_id = v.id
WHERE vs.station_id = ?
  AND vs.last_observation_at IS NOT NULL
ORDER BY v.scheduled_date DESC, vs.scheduled_arrival DESC
LIMIT ?;

-- name: GetStationStatsFromVoyages :one
SELECT
    COUNT(*) as observation_count,
    COALESCE(AVG(COALESCE(arrival_delay, departure_delay)), 0) as average_delay,
    SUM(CASE WHEN COALESCE(arrival_delay, departure_delay) = 0 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_percentage,
    SUM(CASE WHEN is_suppressed = 1 THEN 1 ELSE 0 END) as suppressed_count
FROM voyage_stop
WHERE station_id = ?
  AND last_observation_at IS NOT NULL;

-- name: GetVoyagesByStation :many
SELECT DISTINCT v.id, v.train_number, v.train_category, v.origin_id, v.origin_name,
    v.destination_id, v.destination_name, v.scheduled_date, v.scheduled_departure,
    v.circulation_state, v.created_at, v.updated_at
FROM voyage v
JOIN voyage_stop vs ON v.id = vs.voyage_id
WHERE vs.station_id = ?
  AND vs.last_observation_at IS NOT NULL
ORDER BY v.scheduled_date DESC, v.scheduled_departure DESC
LIMIT ?;

-- name: UpdateObservationVoyageStopLink :exec
UPDATE train_observation
SET voyage_stop_id = ?
WHERE id = ?;
