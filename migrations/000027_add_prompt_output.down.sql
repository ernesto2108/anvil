-- Migration: 000027_add_prompt_output.down.sql
-- SQLite does not support DROP COLUMN (requires 3.35+ and not universally safe).
-- This down migration is a no-op: it leaves the output column in place.
-- Data stored in output will NOT be recovered if this rollback is run.
-- To fully reverse, recreate the table without the column using the 4-step pattern.

SELECT 1;
