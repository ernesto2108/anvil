-- Migration: 000029_add_tool_use_metrics.up.sql
-- Adds input_size_bytes and duration_ms to tool_uses for MCP Analytics (DASH-FEAT-022).
-- input_size_bytes: NOT NULL DEFAULT 0 — safe to add on non-empty tables (SQLite allows ADD COLUMN with DEFAULT).
-- duration_ms: nullable — no default required, existing rows get NULL.

ALTER TABLE tool_uses ADD COLUMN input_size_bytes INTEGER NOT NULL DEFAULT 0;
ALTER TABLE tool_uses ADD COLUMN duration_ms INTEGER;

-- Partial index for MCP-scoped temporal queries.
-- Covers: latency histograms, error-rate trends, server-level aggregation filtered to source='mcp'.
-- SQLite does not support CREATE INDEX CONCURRENTLY — index is built synchronously.
CREATE INDEX idx_tool_uses_mcp_time ON tool_uses(source, mcp_server, timestamp)
    WHERE source = 'mcp';
