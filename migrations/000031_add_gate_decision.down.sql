-- Migration: 000031_add_gate_decision.down.sql
-- Rolls back 000031_add_gate_decision.up.sql.
-- DROP COLUMN requires SQLite >= 3.35.0 (released 2021-03-12).
-- macOS 12+ ships SQLite 3.39+; Linux distros from 2022+ are covered.
-- If your environment has an older SQLite, this rollback will fail — upgrade SQLite
-- or perform a 4-step table-rebuild manually.
-- No data loss beyond the gate_decision column itself (was new in this migration).

ALTER TABLE agents DROP COLUMN gate_decision;
