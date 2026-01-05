-- Add scheduled_date column for deduplication key
ALTER TABLE train_observation ADD COLUMN scheduled_date TEXT;

-- Populate existing records
UPDATE train_observation SET scheduled_date = DATE(scheduled_time) WHERE scheduled_date IS NULL;

-- Create unique index for deduplication
-- Key: train_number + station_id + observation_type + scheduled_date
CREATE UNIQUE INDEX IF NOT EXISTS idx_observation_unique
ON train_observation(train_number, station_id, observation_type, scheduled_date);

-- Delay variation table to track delay changes over time
CREATE TABLE IF NOT EXISTS delay_variation (
    id TEXT PRIMARY KEY,
    observation_id TEXT NOT NULL,
    recorded_at DATETIME NOT NULL,
    delay INTEGER NOT NULL,
    FOREIGN KEY (observation_id) REFERENCES train_observation(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_delay_variation_observation ON delay_variation(observation_id);
CREATE INDEX IF NOT EXISTS idx_delay_variation_recorded ON delay_variation(recorded_at);
