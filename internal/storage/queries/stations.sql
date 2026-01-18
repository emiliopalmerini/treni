-- name: UpsertStation :exec
INSERT INTO stations (code, name, city, region, latitude, longitude)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(code) DO UPDATE SET
    name = excluded.name,
    city = excluded.city,
    region = excluded.region,
    latitude = excluded.latitude,
    longitude = excluded.longitude,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetStation :one
SELECT * FROM stations WHERE code = ?;

-- name: GetStationByName :many
SELECT * FROM stations WHERE name LIKE ? LIMIT 20;

-- name: ListStations :many
SELECT * FROM stations ORDER BY name;
