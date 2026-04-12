ALTER TABLE runs ADD COLUMN session_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_runs_session_id
    ON runs(session_id) WHERE session_id IS NOT NULL;
