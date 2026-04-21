-- Add source column to digests to distinguish automatic checkpoints from
-- manually-triggered digests generated via the dashboard.
--
-- Valid values: 'auto' (digestCheckpointer post-run), 'manual' (dashboard).
-- DEFAULT 'auto' ensures all existing rows are classified correctly without a
-- data migration.
ALTER TABLE digests ADD COLUMN source TEXT NOT NULL DEFAULT 'auto';
