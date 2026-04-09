CREATE TABLE IF NOT EXISTS files (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id    TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    agent_id  TEXT NOT NULL,
    path      TEXT NOT NULL,
    operation TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_files_run_id ON files(run_id);
