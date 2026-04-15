-- SQLite does not support ALTER FOREIGN KEY; table must be recreated to add ON DELETE CASCADE.
-- Rename → recreate with cascade → copy → drop old.

PRAGMA foreign_keys = OFF;

ALTER TABLE tool_uses RENAME TO tool_uses_old;

CREATE TABLE tool_uses (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id     TEXT    NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    agent_id   TEXT    NOT NULL DEFAULT '',
    tool_name  TEXT    NOT NULL,
    tool_input TEXT    NOT NULL DEFAULT '',
    timestamp  TEXT    NOT NULL
);

INSERT INTO tool_uses (id, run_id, agent_id, tool_name, tool_input, timestamp)
SELECT id, run_id, agent_id, tool_name, tool_input, timestamp
FROM tool_uses_old;

DROP TABLE tool_uses_old;

CREATE INDEX IF NOT EXISTS idx_tool_uses_run_id   ON tool_uses(run_id);
CREATE INDEX IF NOT EXISTS idx_tool_uses_tool_name ON tool_uses(run_id, tool_name);

PRAGMA foreign_keys = ON;
