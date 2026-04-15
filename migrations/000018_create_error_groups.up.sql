CREATE TABLE IF NOT EXISTS error_groups (
    id TEXT PRIMARY KEY,
    fingerprint TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    normalized_msg TEXT NOT NULL,
    resolution_status TEXT NOT NULL DEFAULT 'new',
    first_seen_at TEXT NOT NULL,
    last_seen_at TEXT NOT NULL,
    occurrence_count INTEGER NOT NULL DEFAULT 1,
    notes TEXT,
    commit_link TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_error_groups_resolution_status ON error_groups(resolution_status);
CREATE INDEX IF NOT EXISTS idx_error_groups_last_seen ON error_groups(last_seen_at);
