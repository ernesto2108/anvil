ALTER TABLE tool_uses ADD COLUMN source TEXT NOT NULL DEFAULT 'native';
ALTER TABLE tool_uses ADD COLUMN mcp_server TEXT;

CREATE INDEX IF NOT EXISTS idx_tool_uses_source ON tool_uses(run_id, source);
