-- Extend the source CHECK constraint on digests to accept 'direct' in addition
-- to the existing values 'auto', 'manual', 'handoff'.
--
-- SQLite does not support ALTER TABLE ... DROP CONSTRAINT, so we use the
-- standard 4-step table-recreation pattern recommended by the SQLite docs.
-- The FTS triggers referencing this table are dropped implicitly with the
-- original table and recreated at the end.
--
-- Execution order: run BEFORE deploying code that writes source='direct'.

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
    source      TEXT NOT NULL DEFAULT 'auto' CHECK (source IN ('auto','manual','handoff','direct')),
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
