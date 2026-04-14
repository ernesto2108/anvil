-- SQLite no soporta DROP COLUMN en versiones viejas; recrear la tabla.
CREATE TABLE tool_uses_backup AS SELECT id, run_id, agent_id, tool_name, timestamp FROM tool_uses;
DROP TABLE tool_uses;
CREATE TABLE tool_uses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL REFERENCES runs(id),
    agent_id TEXT NOT NULL DEFAULT '',
    tool_name TEXT NOT NULL,
    timestamp TEXT NOT NULL
);
INSERT INTO tool_uses (id, run_id, agent_id, tool_name, timestamp) SELECT id, run_id, agent_id, tool_name, timestamp FROM tool_uses_backup;
DROP TABLE tool_uses_backup;
CREATE INDEX idx_tool_uses_run_id ON tool_uses(run_id);
CREATE INDEX idx_tool_uses_tool_name ON tool_uses(run_id, tool_name);
