-- Stations table for caching station data
CREATE TABLE IF NOT EXISTS stations (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    city TEXT,
    region TEXT,
    latitude REAL,
    longitude REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Delay records for historical tracking
CREATE TABLE IF NOT EXISTS delay_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    train_number TEXT NOT NULL,
    train_category TEXT,
    origin TEXT NOT NULL,
    destination TEXT NOT NULL,
    date DATE NOT NULL,
    delay INTEGER NOT NULL,
    cancelled BOOLEAN DEFAULT FALSE,
    source TEXT DEFAULT 'viaggiatreno',
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Prevent duplicate records for same train on same day
    UNIQUE(train_number, date, source)
);

-- Index for querying by train
CREATE INDEX IF NOT EXISTS idx_delay_records_train ON delay_records(train_number);

-- Index for querying by date range
CREATE INDEX IF NOT EXISTS idx_delay_records_date ON delay_records(date);

-- Index for analytics queries
CREATE INDEX IF NOT EXISTS idx_delay_records_analytics ON delay_records(train_number, date, delay);
