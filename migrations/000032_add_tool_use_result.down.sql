-- Migration: 000032_add_tool_use_result.down.sql
-- Rolls back 000032_add_tool_use_result.up.sql.
-- DROP COLUMN requires SQLite >= 3.35.0 (released 2021-03-12).
-- macOS 12+ ships SQLite 3.39+; Linux distros from 2022+ are covered.

DROP INDEX IF EXISTS idx_tool_uses_bash_result;
ALTER TABLE tool_uses DROP COLUMN tool_output_excerpt;
ALTER TABLE tool_uses DROP COLUMN exit_code;
