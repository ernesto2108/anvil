DROP INDEX IF EXISTS idx_tool_uses_run_id;
DROP INDEX IF EXISTS idx_tool_uses_tool_name;
DROP TABLE IF EXISTS tool_uses;

DROP INDEX IF EXISTS idx_tasks_run_id;
DROP TABLE IF EXISTS tasks;

-- SQLite no soporta DROP COLUMN antes de 3.35.0; recrear tabla si es necesario.
-- Para simplificar, la columna error_reason queda (DEFAULT '' no afecta).
