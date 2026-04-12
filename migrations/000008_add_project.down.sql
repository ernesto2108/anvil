DROP INDEX IF EXISTS idx_runs_project;
-- SQLite doesn't support DROP COLUMN before 3.35.0; recreate table if needed.
-- For simplicity, we just leave the column in place on rollback.
