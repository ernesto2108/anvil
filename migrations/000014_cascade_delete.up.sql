-- Recrear tablas tool_uses, tasks y prompts con ON DELETE CASCADE
-- para que DELETE FROM runs cascade correctamente a todas las tablas hijas.

-- 1. tool_uses
CREATE TABLE tool_uses_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    agent_id TEXT NOT NULL DEFAULT '',
    tool_name TEXT NOT NULL,
    tool_input TEXT NOT NULL DEFAULT '',
    timestamp TEXT NOT NULL
);
INSERT INTO tool_uses_new SELECT id, run_id, agent_id, tool_name, tool_input, timestamp FROM tool_uses;
DROP TABLE tool_uses;
ALTER TABLE tool_uses_new RENAME TO tool_uses;
CREATE INDEX idx_tool_uses_run_id ON tool_uses(run_id);
CREATE INDEX idx_tool_uses_tool_name ON tool_uses(run_id, tool_name);

-- 2. tasks
CREATE TABLE tasks_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    task_id TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TEXT NOT NULL,
    completed_at TEXT
);
INSERT INTO tasks_new SELECT id, run_id, task_id, title, status, created_at, completed_at FROM tasks;
DROP TABLE tasks;
ALTER TABLE tasks_new RENAME TO tasks;
CREATE INDEX idx_tasks_run_id ON tasks(run_id);

-- 3. prompts
CREATE TABLE prompts_new (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id    TEXT    NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    sequence  INTEGER NOT NULL,
    prompt    TEXT    NOT NULL DEFAULT '',
    timestamp TEXT    NOT NULL
);
INSERT INTO prompts_new SELECT id, run_id, sequence, prompt, timestamp FROM prompts;
DROP TABLE prompts;
ALTER TABLE prompts_new RENAME TO prompts;
CREATE INDEX idx_prompts_run_id ON prompts(run_id);
CREATE UNIQUE INDEX uniq_prompts_run_id_sequence ON prompts(run_id, sequence);
