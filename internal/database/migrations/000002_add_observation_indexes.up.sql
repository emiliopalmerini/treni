-- Add indexes for origin_id and destination_id to improve query performance
CREATE INDEX IF NOT EXISTS idx_observation_origin ON train_observation(origin_id);
CREATE INDEX IF NOT EXISTS idx_observation_destination ON train_observation(destination_id);

-- Add compound index for stats queries that group by train route
CREATE INDEX IF NOT EXISTS idx_observation_train_route ON train_observation(train_number, origin_id, destination_id);

-- Add compound index for filtering by circulation state and grouping
CREATE INDEX IF NOT EXISTS idx_observation_stats ON train_observation(circulation_state, train_number, observed_at);
