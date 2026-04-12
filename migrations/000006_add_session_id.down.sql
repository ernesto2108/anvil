DROP INDEX IF EXISTS idx_runs_session_id;

ALTER TABLE runs DROP COLUMN session_id;
