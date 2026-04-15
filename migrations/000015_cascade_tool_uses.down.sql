-- Reverses 000015: recreates tool_uses WITHOUT ON DELETE CASCADE.
-- WARNING: if rows were already deleted via cascade, those rows cannot be restored.

PRAGMA foreign_keys = OFF;

ALTER TABLE tool_uses RENAME TO tool_uses_new;

CREATE TABLE tool_uses (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id     TEXT    NOT NULL REFERENCES runs(id),
    agent_id   TEXT    NOT NULL DEFAULT '',
    tool_name  TEXT    NOT NULL,
    tool_input TEXT    NOT NULL DEFAULT '',
    timestamp  TEXT    NOT NULL
);

INSERT INTO tool_uses (id, run_id, agent_id, tool_name, tool_input, timestamp)
SELECT id, run_id, agent_id, tool_name, tool_input, timestamp
FROM tool_uses_new;

DROP TABLE tool_uses_new;

CREATE INDEX IF NOT EXISTS idx_tool_uses_run_id   ON tool_uses(run_id);
CREATE INDEX IF NOT EXISTS idx_tool_uses_tool_name ON tool_uses(run_id, tool_name);

PRAGMA foreign_keys = ON;
