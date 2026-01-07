-- Drop old journey tables (repurposing for voyage model)
DROP TABLE IF EXISTS journey_stop;
DROP TABLE IF EXISTS journey;

-- Create voyage table (enhanced journey concept)
CREATE TABLE IF NOT EXISTS voyage (
    id TEXT PRIMARY KEY,
    train_number INTEGER NOT NULL,
    train_category TEXT,
    origin_id TEXT NOT NULL,
    origin_name TEXT NOT NULL,
    destination_id TEXT NOT NULL,
    destination_name TEXT NOT NULL,
    scheduled_date TEXT NOT NULL,
    scheduled_departure DATETIME NOT NULL,
    circulation_state INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(train_number, origin_id, scheduled_date)
);

CREATE INDEX idx_voyage_train_number ON voyage(train_number);
CREATE INDEX idx_voyage_scheduled_date ON voyage(scheduled_date);
CREATE INDEX idx_voyage_origin ON voyage(origin_id);
CREATE INDEX idx_voyage_updated ON voyage(updated_at);

-- Create voyage_stop table (enhanced journey_stop)
CREATE TABLE IF NOT EXISTS voyage_stop (
    id TEXT PRIMARY KEY,
    voyage_id TEXT NOT NULL,
    station_id TEXT NOT NULL,
    station_name TEXT NOT NULL,
    stop_sequence INTEGER NOT NULL,
    stop_type TEXT,
    scheduled_arrival DATETIME,
    scheduled_departure DATETIME,
    actual_arrival DATETIME,
    actual_departure DATETIME,
    arrival_delay INTEGER DEFAULT 0,
    departure_delay INTEGER DEFAULT 0,
    platform TEXT,
    is_suppressed INTEGER DEFAULT 0,
    last_observation_at DATETIME,
    FOREIGN KEY (voyage_id) REFERENCES voyage(id) ON DELETE CASCADE,
    UNIQUE(voyage_id, station_id)
);

CREATE INDEX idx_voyage_stop_voyage_id ON voyage_stop(voyage_id);
CREATE INDEX idx_voyage_stop_station_id ON voyage_stop(station_id);
CREATE INDEX idx_voyage_stop_sequence ON voyage_stop(voyage_id, stop_sequence);
CREATE INDEX idx_voyage_stop_observed ON voyage_stop(last_observation_at);

-- Add voyage_stop_id to train_observation table
ALTER TABLE train_observation ADD COLUMN voyage_stop_id TEXT REFERENCES voyage_stop(id);
CREATE INDEX idx_observation_voyage_stop ON train_observation(voyage_stop_id);
