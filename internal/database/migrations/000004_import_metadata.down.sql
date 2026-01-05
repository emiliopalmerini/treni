-- Remove updated_at from station (SQLite doesn't support DROP COLUMN before 3.35)
-- We'll recreate the table without the column
CREATE TABLE station_backup AS SELECT id, name, region, latitude, longitude, is_favorite FROM station;
DROP TABLE station;
CREATE TABLE station (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    region INTEGER DEFAULT 0,
    latitude REAL DEFAULT 0,
    longitude REAL DEFAULT 0,
    is_favorite INTEGER DEFAULT 0
);
INSERT INTO station SELECT * FROM station_backup;
DROP TABLE station_backup;
CREATE INDEX IF NOT EXISTS idx_station_name ON station(name);

DROP INDEX IF EXISTS idx_import_metadata_last;
DROP TABLE IF EXISTS import_metadata;
