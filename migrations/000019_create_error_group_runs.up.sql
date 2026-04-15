CREATE TABLE IF NOT EXISTS error_group_runs (
    error_group_id TEXT NOT NULL REFERENCES error_groups(id) ON DELETE CASCADE,
    run_id TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    agent_name TEXT NOT NULL DEFAULT '',
    error_msg TEXT NOT NULL,
    exit_code INTEGER,
    occurred_at TEXT NOT NULL,
    PRIMARY KEY (error_group_id, run_id, agent_name)
);

CREATE INDEX IF NOT EXISTS idx_error_group_runs_run_id ON error_group_runs(run_id);
