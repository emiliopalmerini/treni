-- Remove voyage_stop_id from train_observation
DROP INDEX IF EXISTS idx_observation_voyage_stop;
ALTER TABLE train_observation DROP COLUMN voyage_stop_id;

-- Drop voyage tables
DROP TABLE IF EXISTS voyage_stop;
DROP TABLE IF EXISTS voyage;

-- Recreate original journey tables
CREATE TABLE IF NOT EXISTS journey (
    id TEXT PRIMARY KEY,
    train_number INTEGER NOT NULL,
    origin_id TEXT NOT NULL,
    origin_name TEXT NOT NULL,
    destination_id TEXT NOT NULL,
    destination_name TEXT NOT NULL,
    scheduled_departure DATETIME,
    actual_departure DATETIME,
    delay INTEGER DEFAULT 0,
    recorded_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_journey_train_number ON journey(train_number);
CREATE INDEX IF NOT EXISTS idx_journey_recorded_at ON journey(recorded_at);

CREATE TABLE IF NOT EXISTS journey_stop (
    id TEXT PRIMARY KEY,
    journey_id TEXT NOT NULL,
    station_id TEXT NOT NULL,
    station_name TEXT NOT NULL,
    scheduled_arrival DATETIME,
    scheduled_departure DATETIME,
    actual_arrival DATETIME,
    actual_departure DATETIME,
    arrival_delay INTEGER DEFAULT 0,
    departure_delay INTEGER DEFAULT 0,
    platform TEXT,
    FOREIGN KEY (journey_id) REFERENCES journey(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_journey_stop_journey_id ON journey_stop(journey_id);
