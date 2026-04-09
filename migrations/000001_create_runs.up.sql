CREATE TABLE IF NOT EXISTS runs (
    id            TEXT PRIMARY KEY,
    task_id       TEXT NOT NULL,
    task_desc     TEXT,
    status        TEXT NOT NULL DEFAULT 'running',
    complexity    TEXT,
    provider      TEXT,
    started_at    TEXT NOT NULL,
    ended_at      TEXT,
    duration_ms   INTEGER,
    total_tokens  INTEGER DEFAULT 0,
    files_touched INTEGER DEFAULT 0,
    agents_count  INTEGER DEFAULT 0,
    qa_score      REAL
);

CREATE INDEX IF NOT EXISTS idx_runs_started_at ON runs(started_at);
