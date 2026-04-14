-- Hook 1 (PreToolUse): tabla para rastrear uso de herramientas
CREATE TABLE IF NOT EXISTS tool_uses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL REFERENCES runs(id),
    agent_id TEXT NOT NULL DEFAULT '',
    tool_name TEXT NOT NULL,
    timestamp TEXT NOT NULL
);
CREATE INDEX idx_tool_uses_run_id ON tool_uses(run_id);
CREATE INDEX idx_tool_uses_tool_name ON tool_uses(run_id, tool_name);

-- Hook 2 (StopFailure): razón de error en runs
ALTER TABLE runs ADD COLUMN error_reason TEXT NOT NULL DEFAULT '';

-- Hook 3 (TaskCreated/TaskCompleted): tabla de tasks
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL REFERENCES runs(id),
    task_id TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TEXT NOT NULL,
    completed_at TEXT
);
CREATE INDEX idx_tasks_run_id ON tasks(run_id);
