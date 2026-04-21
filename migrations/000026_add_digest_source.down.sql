-- SQLite does not support DROP COLUMN (before 3.35) and golang-migrate on
-- SQLite wraps each migration in an implicit transaction that prevents the
-- 4-step table-recreation pattern inside a migration file.
--
-- Intentional no-op: rolling back this migration leaves the `source` column
-- in place. Data is not lost. If you need a clean rollback, recreate the
-- database from a backup taken before 000026 was applied.
SELECT 1;
