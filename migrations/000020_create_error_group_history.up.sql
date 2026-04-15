CREATE TABLE IF NOT EXISTS error_group_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    error_group_id TEXT NOT NULL REFERENCES error_groups(id) ON DELETE CASCADE,
    old_status TEXT,
    new_status TEXT NOT NULL,
    note TEXT,
    commit_link TEXT,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_error_group_history_group ON error_group_history(error_group_id);
