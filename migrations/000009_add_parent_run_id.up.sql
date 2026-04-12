ALTER TABLE runs ADD COLUMN parent_run_id TEXT;
CREATE INDEX idx_runs_parent_run_id ON runs(parent_run_id);
