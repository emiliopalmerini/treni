-- name: GetImportMetadata :one
SELECT entity_type, last_import, record_count, import_duration_ms, status, error_message
FROM import_metadata WHERE entity_type = ?;

-- name: UpsertImportMetadata :exec
INSERT INTO import_metadata (entity_type, last_import, record_count, import_duration_ms, status, error_message)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(entity_type) DO UPDATE SET
    last_import = excluded.last_import,
    record_count = excluded.record_count,
    import_duration_ms = excluded.import_duration_ms,
    status = excluded.status,
    error_message = excluded.error_message;
