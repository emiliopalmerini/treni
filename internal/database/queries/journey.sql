-- name: CreateJourney :exec
INSERT INTO journey (id, train_number, origin_id, origin_name, destination_id, destination_name,
    scheduled_departure, actual_departure, delay, recorded_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetJourneyByID :one
SELECT id, train_number, origin_id, origin_name, destination_id, destination_name,
    scheduled_departure, actual_departure, delay, recorded_at
FROM journey WHERE id = ?;

-- name: ListJourneys :many
SELECT id, train_number, origin_id, origin_name, destination_id, destination_name,
    scheduled_departure, actual_departure, delay, recorded_at
FROM journey ORDER BY recorded_at DESC LIMIT 100;

-- name: ListJourneysByTrain :many
SELECT id, train_number, origin_id, origin_name, destination_id, destination_name,
    scheduled_departure, actual_departure, delay, recorded_at
FROM journey WHERE train_number = ? ORDER BY recorded_at DESC;

-- name: ListJourneysByDateRange :many
SELECT id, train_number, origin_id, origin_name, destination_id, destination_name,
    scheduled_departure, actual_departure, delay, recorded_at
FROM journey WHERE recorded_at BETWEEN ? AND ? ORDER BY recorded_at DESC;

-- name: UpdateJourney :exec
UPDATE journey SET train_number = ?, origin_id = ?, origin_name = ?, destination_id = ?,
    destination_name = ?, scheduled_departure = ?, actual_departure = ?, delay = ?
WHERE id = ?;

-- name: DeleteJourney :exec
DELETE FROM journey WHERE id = ?;

-- name: CreateJourneyStop :exec
INSERT INTO journey_stop (id, journey_id, station_id, station_name, scheduled_arrival,
    scheduled_departure, actual_arrival, actual_departure, arrival_delay, departure_delay, platform)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetJourneyStops :many
SELECT id, journey_id, station_id, station_name, scheduled_arrival, scheduled_departure,
    actual_arrival, actual_departure, arrival_delay, departure_delay, platform
FROM journey_stop WHERE journey_id = ? ORDER BY scheduled_departure;
