-- Tabla para almacenar los prompts de usuario de cada run
CREATE TABLE IF NOT EXISTS prompts (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id    TEXT    NOT NULL REFERENCES runs(id),
    sequence  INTEGER NOT NULL,
    prompt    TEXT    NOT NULL DEFAULT '',
    timestamp TEXT    NOT NULL
);
CREATE INDEX idx_prompts_run_id ON prompts(run_id);
CREATE UNIQUE INDEX uniq_prompts_run_id_sequence ON prompts(run_id, sequence);
