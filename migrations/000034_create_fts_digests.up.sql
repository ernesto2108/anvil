-- FTS5 virtual table for keyword search over digest text fields.
-- Complements vec_digests (vector KNN) with exact-term BM25 ranking.
-- digest_id and project are UNINDEXED — stored for retrieval/filtering, not tokenized.
CREATE VIRTUAL TABLE IF NOT EXISTS fts_digests USING fts5(
    digest_id  UNINDEXED,
    project    UNINDEXED,
    summary,
    decisions,
    edge_cases,
    errors,
    tokenize='unicode61 remove_diacritics 1'
);

-- Backfill existing rows.
INSERT INTO fts_digests(digest_id, project, summary, decisions, edge_cases, errors)
SELECT id, project, summary, decisions, edge_cases, errors FROM digests;

-- Keep fts_digests in sync with digests automatically.
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
