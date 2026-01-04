-- name: CreateWatchedTrain :exec
INSERT INTO watched_train (id, train_number, origin_id, origin_name, destination, days_of_week, notes, active, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetWatchedTrainByID :one
SELECT id, train_number, origin_id, origin_name, destination, days_of_week, notes, active, created_at
FROM watched_train WHERE id = ?;

-- name: ListWatchedTrains :many
SELECT id, train_number, origin_id, origin_name, destination, days_of_week, notes, active, created_at
FROM watched_train ORDER BY created_at DESC;

-- name: ListActiveWatchedTrains :many
SELECT id, train_number, origin_id, origin_name, destination, days_of_week, notes, active, created_at
FROM watched_train WHERE active = 1 ORDER BY created_at DESC;

-- name: UpdateWatchedTrain :exec
UPDATE watched_train SET train_number = ?, origin_id = ?, origin_name = ?, destination = ?,
    days_of_week = ?, notes = ?, active = ? WHERE id = ?;

-- name: DeleteWatchedTrain :exec
DELETE FROM watched_train WHERE id = ?;

-- name: CreateTrainCheck :exec
INSERT INTO train_check (id, watched_id, train_number, delay, status, checked_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetTrainChecksByWatched :many
SELECT id, watched_id, train_number, delay, status, checked_at
FROM train_check WHERE watched_id = ? ORDER BY checked_at DESC LIMIT 100;

-- name: GetRecentTrainChecks :many
SELECT id, watched_id, train_number, delay, status, checked_at
FROM train_check ORDER BY checked_at DESC LIMIT ?;
