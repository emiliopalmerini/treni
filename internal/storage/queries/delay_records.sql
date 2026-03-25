-- name: InsertDelayRecord :exec
INSERT INTO delay_records (train_number, train_category, origin, destination, date, delay, cancelled, source)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(train_number, date, source) DO UPDATE SET
    delay = excluded.delay,
    cancelled = excluded.cancelled,
    recorded_at = CURRENT_TIMESTAMP;

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
GROUP BY train_number;
