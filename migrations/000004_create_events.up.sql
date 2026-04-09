CREATE TABLE IF NOT EXISTS events (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id         TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    event_id       TEXT NOT NULL UNIQUE,
    event_type     TEXT NOT NULL,
    schema_version TEXT NOT NULL DEFAULT '1',
    timestamp      TEXT NOT NULL,
    payload        TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_events_run_id ON events(run_id);
CREATE INDEX IF NOT EXISTS idx_events_event_type ON events(event_type);
