DROP INDEX IF EXISTS idx_delay_variation_recorded;
DROP INDEX IF EXISTS idx_delay_variation_observation;
DROP TABLE IF EXISTS delay_variation;
DROP INDEX IF EXISTS idx_observation_unique;
-- Note: SQLite doesn't support DROP COLUMN in older versions
-- The scheduled_date column will remain but can be ignored
