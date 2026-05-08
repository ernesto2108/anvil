-- Rollback: restore the source CHECK constraint to ('auto','manual','handoff'),
-- removing 'direct' from the allowed values.
--
-- SAFETY GUARD: if any row has source='direct' this migration aborts cleanly
-- with RAISE(ABORT). No rows are deleted or modified — the caller must remove
-- or reclassify those rows before running the rollback.

-- Hard guard: abort if 'direct' rows exist to prevent data loss / silent
-- constraint violations after the table is recreated without that value.
SELECT RAISE(ABORT, 'Cannot roll back 000035: rows with source=''direct'' exist. Remove or reclassify them first.')
WHERE EXISTS (SELECT 1 FROM digests WHERE source = 'direct');

CREATE TABLE digests_new (
    id          TEXT PRIMARY KEY,
    run_id      TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    project     TEXT NOT NULL,
    summary     TEXT NOT NULL,
    decisions   TEXT NOT NULL DEFAULT '[]',
    edge_cases  TEXT NOT NULL DEFAULT '[]',
    errors      TEXT NOT NULL DEFAULT '[]',
    embedding   BLOB,
    model_used  TEXT NOT NULL DEFAULT 'claude-haiku-4-5',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    source      TEXT NOT NULL DEFAULT 'auto' CHECK (source IN ('auto','manual','handoff')),
    UNIQUE(run_id)
);

INSERT INTO digests_new
SELECT id, run_id, project, summary, decisions, edge_cases, errors,
       embedding, model_used, created_at, updated_at, source
FROM digests;

DROP TABLE digests;

ALTER TABLE digests_new RENAME TO digests;

CREATE INDEX idx_digests_project    ON digests(project);
CREATE INDEX idx_digests_created_at ON digests(created_at);

CREATE TRIGGER IF NOT EXISTS fts_digests_ai AFTER INSERT ON digests BEGIN
    INSERT INTO fts_digests(digest_id, project, summary, decisions, edge_cases, errors)
    VALUES (new.id, new.project, new.summary, new.decisions, new.edge_cases, new.errors);
END;

CREATE TRIGGER IF NOT EXISTS fts_digests_au AFTER UPDATE ON digests BEGIN
    DELETE FROM fts_digests WHERE digest_id = old.id;
    INSERT INTO fts_digests(digest_id, project, summary, decisions, edge_cases, errors)
    VALUES (new.id, new.project, new.summary, new.decisions, new.edge_cases, new.errors);
END;

CREATE TRIGGER IF NOT EXISTS fts_digests_ad AFTER DELETE ON digests BEGIN
    DELETE FROM fts_digests WHERE digest_id = old.id;
END;
