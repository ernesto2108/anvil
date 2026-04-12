ALTER TABLE runs ADD COLUMN project TEXT NOT NULL DEFAULT '';
CREATE INDEX idx_runs_project ON runs(project);
