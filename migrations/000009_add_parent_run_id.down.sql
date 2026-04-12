-- WARNING: rolling back this migration drops the parent_run_id column and its data permanently.
-- Requires SQLite >= 3.35.0 for DROP COLUMN support.
DROP INDEX IF EXISTS idx_runs_parent_run_id;
ALTER TABLE runs DROP COLUMN parent_run_id;
