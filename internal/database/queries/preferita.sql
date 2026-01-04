-- name: ListPreferite :many
SELECT station_id, name FROM preferita ORDER BY name;

-- name: AddPreferita :exec
INSERT OR REPLACE INTO preferita (station_id, name) VALUES (?, ?);

-- name: RemovePreferita :exec
DELETE FROM preferita WHERE station_id = ?;

-- name: PreferitaExists :one
SELECT COUNT(*) FROM preferita WHERE station_id = ?;
