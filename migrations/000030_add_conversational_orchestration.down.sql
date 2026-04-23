-- Migration: 000030_add_conversational_orchestration.down.sql
-- Rolls back 000030_add_conversational_orchestration.up.sql.
-- DROP COLUMN requires SQLite >= 3.35.0 (released 2021-03-12).
-- macOS 12+ ships SQLite 3.39+; Linux distros from 2022+ are covered.
-- If your environment has an older SQLite, this rollback will fail — upgrade SQLite
-- or perform a 4-step table-rebuild manually.
-- No data loss beyond the columns themselves (all were new in this migration).

ALTER TABLE agents DROP COLUMN files_touched_list;

ALTER TABLE runs DROP COLUMN pipeline_snapshot;
ALTER TABLE runs DROP COLUMN stack;
ALTER TABLE runs DROP COLUMN pipeline;
