-- Add degraded flag to runs: set when a run was aborted due to budget exhaustion.
ALTER TABLE runs ADD COLUMN degraded INTEGER NOT NULL DEFAULT 0;

-- run_plans: budget parameters and plan content written at run start.
-- content is the plan.md markdown; may be empty if the Leader did not produce one.
CREATE TABLE IF NOT EXISTS run_plans (
    id             TEXT    PRIMARY KEY,
    run_id         TEXT    NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    project        TEXT    NOT NULL,
    max_retries    INTEGER NOT NULL DEFAULT 2,
    max_cost_usd   REAL    NOT NULL DEFAULT 0.50,
    content        TEXT    NOT NULL DEFAULT '',
    created_at     TEXT    NOT NULL,
    UNIQUE(run_id)
);

CREATE INDEX IF NOT EXISTS idx_run_plans_run_id  ON run_plans(run_id);
CREATE INDEX IF NOT EXISTS idx_run_plans_project ON run_plans(project);

-- dream_insights: patterns distilled from past digests by the dreaming process.
-- status=pending_review until a human approves (applied) or dismisses the insight.
CREATE TABLE IF NOT EXISTS dream_insights (
    id           TEXT PRIMARY KEY,
    project      TEXT NOT NULL,
    insight      TEXT NOT NULL,
    source_runs  TEXT NOT NULL DEFAULT '[]',
    status       TEXT NOT NULL DEFAULT 'pending_review'
                 CHECK (status IN ('pending_review','applied','dismissed')),
    created_at   TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_dream_insights_project ON dream_insights(project);
CREATE INDEX IF NOT EXISTS idx_dream_insights_status  ON dream_insights(status);
