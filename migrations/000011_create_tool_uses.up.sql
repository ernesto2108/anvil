CREATE TABLE IF NOT EXISTS tool_uses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL REFERENCES runs(id),
    agent_id TEXT NOT NULL DEFAULT '',
    tool_name TEXT NOT NULL,
    tool_input TEXT NOT NULL DEFAULT '',
    timestamp TEXT NOT NULL
);
CREATE INDEX idx_tool_uses_run_id ON tool_uses(run_id);
CREATE INDEX idx_tool_uses_tool_name ON tool_uses(run_id, tool_name);
