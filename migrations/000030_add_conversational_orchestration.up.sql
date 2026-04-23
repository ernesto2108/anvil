-- Migration: 000030_add_conversational_orchestration.up.sql
-- Adds columns to runs and agents tables for conversational orchestration persistence.
-- pipeline: preset name (e.g. "feat", "bug") — empty string for ad-hoc orchestrations.
-- stack: stack declared at orchestration start (e.g. "go", "react") — empty for ad-hoc.
-- pipeline_snapshot: JSON array of pipeline nodes serialized at start time — NULL for ad-hoc.
--   Used by load_orchestration to compute pending_roles.
-- files_touched_list: JSON array of file paths touched by this agent step — complements
--   the integer count in runs.files_touched.
-- All columns have DEFAULT or are nullable — safe to add on non-empty tables.

ALTER TABLE runs ADD COLUMN pipeline TEXT NOT NULL DEFAULT '';
ALTER TABLE runs ADD COLUMN stack TEXT NOT NULL DEFAULT '';
ALTER TABLE runs ADD COLUMN pipeline_snapshot TEXT;

ALTER TABLE agents ADD COLUMN files_touched_list TEXT NOT NULL DEFAULT '[]';
