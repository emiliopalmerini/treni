-- name: CreateStation :exec
INSERT INTO station (id, name, region, latitude, longitude)
VALUES (?, ?, ?, ?, ?);

-- name: GetStationByID :one
SELECT id, name, region, latitude, longitude
FROM station WHERE id = ?;

-- name: ListStations :many
SELECT id, name, region, latitude, longitude
FROM station ORDER BY name;

-- name: SearchStations :many
SELECT id, name, region, latitude, longitude
FROM station WHERE name LIKE ? ORDER BY name LIMIT 20;

-- name: ListStationsWithCoordinates :many
SELECT id, name, region, latitude, longitude
FROM station WHERE latitude != 0 AND longitude != 0;

-- name: CountStations :one
SELECT COUNT(*) FROM station;

-- name: UpdateStation :exec
UPDATE station SET name = ?, region = ?, latitude = ?, longitude = ?
WHERE id = ?;

-- name: DeleteStation :exec
DELETE FROM station WHERE id = ?;

-- name: UpsertStation :exec
INSERT INTO station (id, name, region, latitude, longitude)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    region = excluded.region,
    latitude = excluded.latitude,
    longitude = excluded.longitude;
