CREATE TABLE digests (
    id          TEXT PRIMARY KEY,
    run_id      TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    project     TEXT NOT NULL,
    summary     TEXT NOT NULL,
    decisions   TEXT NOT NULL DEFAULT '[]',
    edge_cases  TEXT NOT NULL DEFAULT '[]',
    errors      TEXT NOT NULL DEFAULT '[]',
    embedding   BLOB,
    model_used  TEXT NOT NULL DEFAULT 'claude-haiku-4-5',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    UNIQUE(run_id)
);

CREATE INDEX idx_digests_project ON digests(project);
CREATE INDEX idx_digests_created_at ON digests(created_at);
