CREATE TABLE IF NOT EXISTS agents (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id        TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    agent_id      TEXT NOT NULL,
    agent_role    TEXT NOT NULL,
    sequence      INTEGER,
    depends_on    TEXT,
    model         TEXT,
    status        TEXT NOT NULL DEFAULT 'pending',
    started_at    TEXT,
    ended_at      TEXT,
    duration_ms   INTEGER,
    tokens_input  INTEGER DEFAULT 0,
    tokens_output INTEGER DEFAULT 0,
    tokens_total  INTEGER DEFAULT 0,
    exit_code     INTEGER,
    error_msg     TEXT
);

CREATE INDEX IF NOT EXISTS idx_agents_run_id ON agents(run_id);
