-- Migration: 000032_add_tool_use_result.up.sql
-- Adds exit_code and tool_output_excerpt to tool_uses for direct-mode verify gate.
-- Captured from PostToolUse.tool_response for native Bash invocations.
-- exit_code: nullable — NULL means PostToolUse never arrived or tool wasn't Bash.
-- tool_output_excerpt: nullable — first 4KB of stdout+stderr for debugging the gate.

ALTER TABLE tool_uses ADD COLUMN exit_code INTEGER;
ALTER TABLE tool_uses ADD COLUMN tool_output_excerpt TEXT;

-- Partial index for fast verify-direct lookups: most queries filter by run_id +
-- tool_name = 'Bash' + exit_code IS NOT NULL (only completed Bash invocations).
CREATE INDEX idx_tool_uses_bash_result ON tool_uses(run_id, timestamp)
    WHERE tool_name = 'Bash' AND exit_code IS NOT NULL;
