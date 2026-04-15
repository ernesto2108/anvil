-- SQLite does not support ALTER FOREIGN KEY; table must be recreated to add ON DELETE CASCADE.
-- Rename → recreate with cascade → copy → drop old.

PRAGMA foreign_keys = OFF;

ALTER TABLE prompts RENAME TO prompts_old;

CREATE TABLE prompts (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id    TEXT    NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    sequence  INTEGER NOT NULL,
    prompt    TEXT    NOT NULL DEFAULT '',
    timestamp TEXT    NOT NULL
);

INSERT INTO prompts (id, run_id, sequence, prompt, timestamp)
SELECT id, run_id, sequence, prompt, timestamp
FROM prompts_old;

DROP TABLE prompts_old;

CREATE INDEX IF NOT EXISTS idx_prompts_run_id             ON prompts(run_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_prompts_run_id_sequence ON prompts(run_id, sequence);

PRAGMA foreign_keys = ON;
