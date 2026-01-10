-- Recreate watchlist tables
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
