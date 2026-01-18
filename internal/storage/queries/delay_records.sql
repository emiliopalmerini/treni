-- name: InsertDelayRecord :exec
INSERT INTO delay_records (train_number, train_category, origin, destination, date, delay, cancelled, source)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(train_number, date, source) DO UPDATE SET
    delay = excluded.delay,
    cancelled = excluded.cancelled,
    recorded_at = CURRENT_TIMESTAMP;

-- name: GetDelayRecordsByTrain :many
SELECT * FROM delay_records
WHERE train_number = ?
ORDER BY date DESC;

-- name: GetDelayRecordsByTrainInRange :many
SELECT * FROM delay_records
WHERE train_number = ?
AND date BETWEEN ? AND ?
ORDER BY date DESC;

-- name: GetDelayRecordsByDateRange :many
SELECT * FROM delay_records
WHERE date BETWEEN ? AND ?
ORDER BY date DESC, train_number;

-- name: GetTrainStats :one
SELECT
    train_number,
    COUNT(*) as total_trips,
    SUM(CASE WHEN delay <= 5 AND cancelled = FALSE THEN 1 ELSE 0 END) as on_time_trips,
    SUM(CASE WHEN delay > 5 AND cancelled = FALSE THEN 1 ELSE 0 END) as delayed_trips,
    SUM(CASE WHEN cancelled = TRUE THEN 1 ELSE 0 END) as cancelled_trips,
    AVG(CASE WHEN cancelled = FALSE THEN delay ELSE NULL END) as average_delay,
    MAX(CASE WHEN cancelled = FALSE THEN delay ELSE NULL END) as max_delay
FROM delay_records
WHERE train_number = ?
AND date BETWEEN ? AND ?
GROUP BY train_number;

-- name: GetMostDelayedTrains :many
SELECT
    train_number,
    train_category,
    origin,
    destination,
    COUNT(*) as trip_count,
    AVG(delay) as avg_delay,
    MAX(delay) as max_delay
FROM delay_records
WHERE date BETWEEN ? AND ?
AND cancelled = FALSE
GROUP BY train_number, train_category, origin, destination
ORDER BY avg_delay DESC
LIMIT ?;

-- name: GetMostReliableTrains :many
SELECT
    train_number,
    train_category,
    origin,
    destination,
    COUNT(*) as trip_count,
    AVG(delay) as avg_delay,
    SUM(CASE WHEN delay <= 5 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as on_time_rate
FROM delay_records
WHERE date BETWEEN ? AND ?
AND cancelled = FALSE
GROUP BY train_number, train_category, origin, destination
HAVING COUNT(*) >= 5
ORDER BY on_time_rate DESC, avg_delay ASC
LIMIT ?;
