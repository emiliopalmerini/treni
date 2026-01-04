-- Station table
CREATE TABLE IF NOT EXISTS station (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    region INTEGER DEFAULT 0,
    latitude REAL DEFAULT 0,
    longitude REAL DEFAULT 0,
    is_favorite INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_station_name ON station(name);

-- Journey table
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

-- Journey stops table
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

-- Watched trains table
CREATE TABLE IF NOT EXISTS watched_train (
    id TEXT PRIMARY KEY,
    train_number INTEGER NOT NULL,
    origin_id TEXT NOT NULL,
    origin_name TEXT NOT NULL,
    destination TEXT NOT NULL,
    days_of_week TEXT,
    notes TEXT,
    active INTEGER DEFAULT 1,
    created_at DATETIME NOT NULL
);

-- Train checks table
CREATE TABLE IF NOT EXISTS train_check (
    id TEXT PRIMARY KEY,
    watched_id TEXT NOT NULL,
    train_number INTEGER NOT NULL,
    delay INTEGER DEFAULT 0,
    status TEXT NOT NULL,
    checked_at DATETIME NOT NULL,
    FOREIGN KEY (watched_id) REFERENCES watched_train(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_train_check_watched_id ON train_check(watched_id);
CREATE INDEX IF NOT EXISTS idx_train_check_checked_at ON train_check(checked_at);

-- Train observation table
CREATE TABLE IF NOT EXISTS train_observation (
    id TEXT PRIMARY KEY,
    observed_at DATETIME NOT NULL,
    station_id TEXT NOT NULL,
    station_name TEXT NOT NULL,
    observation_type TEXT NOT NULL,
    train_number INTEGER NOT NULL,
    train_category TEXT,
    origin_id TEXT,
    origin_name TEXT,
    destination_id TEXT,
    destination_name TEXT,
    scheduled_time DATETIME,
    delay INTEGER DEFAULT 0,
    platform TEXT,
    circulation_state INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_observation_train ON train_observation(train_number);
CREATE INDEX IF NOT EXISTS idx_observation_station ON train_observation(station_id);
CREATE INDEX IF NOT EXISTS idx_observation_observed ON train_observation(observed_at);
CREATE INDEX IF NOT EXISTS idx_observation_category ON train_observation(train_category);

-- Preferita table
CREATE TABLE IF NOT EXISTS preferita (
    station_id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);
