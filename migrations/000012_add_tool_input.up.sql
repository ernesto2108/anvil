-- Almacenar el input de cada tool call (e.g. comando Bash)
ALTER TABLE tool_uses ADD COLUMN tool_input TEXT NOT NULL DEFAULT '';
