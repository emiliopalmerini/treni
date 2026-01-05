-- Import metadata table to track data import status
CREATE TABLE IF NOT EXISTS import_metadata (
    entity_type TEXT PRIMARY KEY,
    last_import DATETIME NOT NULL,
    record_count INTEGER DEFAULT 0,
    import_duration_ms INTEGER,
    status TEXT DEFAULT 'success',
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_import_metadata_last ON import_metadata(last_import);

-- Add updated_at column to station table for staleness tracking
ALTER TABLE station ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP;
