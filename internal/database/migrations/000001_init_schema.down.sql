DROP INDEX IF EXISTS idx_observation_category;
DROP INDEX IF EXISTS idx_observation_observed;
DROP INDEX IF EXISTS idx_observation_station;
DROP INDEX IF EXISTS idx_observation_train;
DROP INDEX IF EXISTS idx_train_check_checked_at;
DROP INDEX IF EXISTS idx_train_check_watched_id;
DROP INDEX IF EXISTS idx_journey_stop_journey_id;
DROP INDEX IF EXISTS idx_journey_recorded_at;
DROP INDEX IF EXISTS idx_journey_train_number;
DROP INDEX IF EXISTS idx_station_name;

DROP TABLE IF EXISTS preferita;
DROP TABLE IF EXISTS train_observation;
DROP TABLE IF EXISTS train_check;
DROP TABLE IF EXISTS watched_train;
DROP TABLE IF EXISTS journey_stop;
DROP TABLE IF EXISTS journey;
DROP TABLE IF EXISTS station;
