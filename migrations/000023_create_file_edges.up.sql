CREATE TABLE file_edges (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id      TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    source_path TEXT NOT NULL,
    target_path TEXT NOT NULL,
    relation    TEXT NOT NULL DEFAULT 'imports',
    agent_id    TEXT,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX idx_file_edges_run_id ON file_edges(run_id);

CREATE UNIQUE INDEX uniq_file_edges_run_source_target ON file_edges(run_id, source_path, target_path);
