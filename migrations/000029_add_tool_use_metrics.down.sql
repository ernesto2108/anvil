-- Migration: 000029_add_tool_use_metrics.down.sql
-- Rolls back 000029_add_tool_use_metrics.up.sql.
--
-- WARNING — PARTIAL ROLLBACK:
-- The partial index is dropped cleanly.
-- The columns input_size_bytes and duration_ms CANNOT be removed: SQLite versions
-- prior to 3.35.0 do not support DROP COLUMN, and rusqlite/go-sqlite3 ship against
-- older SQLite builds. Rolling back this migration leaves those columns in place but
-- harmless — no application code in earlier versions references them.
-- If a full column removal is required, use the 4-step table-rebuild pattern in a
-- subsequent migration.

DROP INDEX IF EXISTS idx_tool_uses_mcp_time;
