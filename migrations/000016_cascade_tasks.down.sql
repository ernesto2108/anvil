-- Reverses 000016: recreates tasks WITHOUT ON DELETE CASCADE.
-- WARNING: if rows were already deleted via cascade, those rows cannot be restored.

PRAGMA foreign_keys = OFF;

ALTER TABLE tasks RENAME TO tasks_new;

CREATE TABLE tasks (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id       TEXT    NOT NULL REFERENCES runs(id),
    task_id      TEXT    NOT NULL,
    title        TEXT    NOT NULL DEFAULT '',
    status       TEXT    NOT NULL DEFAULT 'pending',
    created_at   TEXT    NOT NULL,
    completed_at TEXT
);

INSERT INTO tasks (id, run_id, task_id, title, status, created_at, completed_at)
SELECT id, run_id, task_id, title, status, created_at, completed_at
FROM tasks_new;

DROP TABLE tasks_new;

CREATE INDEX IF NOT EXISTS idx_tasks_run_id ON tasks(run_id);

PRAGMA foreign_keys = ON;
