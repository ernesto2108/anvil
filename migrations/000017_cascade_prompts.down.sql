-- Reverses 000017: recreates prompts WITHOUT ON DELETE CASCADE.
-- WARNING: if rows were already deleted via cascade, those rows cannot be restored.

PRAGMA foreign_keys = OFF;

ALTER TABLE prompts RENAME TO prompts_new;

CREATE TABLE prompts (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id    TEXT    NOT NULL REFERENCES runs(id),
    sequence  INTEGER NOT NULL,
    prompt    TEXT    NOT NULL DEFAULT '',
    timestamp TEXT    NOT NULL
);

INSERT INTO prompts (id, run_id, sequence, prompt, timestamp)
SELECT id, run_id, sequence, prompt, timestamp
FROM prompts_new;

DROP TABLE prompts_new;

CREATE INDEX IF NOT EXISTS idx_prompts_run_id             ON prompts(run_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_prompts_run_id_sequence ON prompts(run_id, sequence);

PRAGMA foreign_keys = ON;
