-- Create vec_digests virtual table backed by sqlite-vec for fast cosine KNN
-- search over digest embeddings. The existing digests.embedding BLOB column
-- is preserved as the source-of-truth and fallback (double-write strategy).
--
-- Partition key on `project` lets us pre-filter the KNN scan by project
-- without an extra IN clause. Embedding dimension is fixed at 768 — matches
-- the Ollama nomic-embed-text model used by the runner. If the embedder ever
-- changes dimension, this table must be recreated.
CREATE VIRTUAL TABLE IF NOT EXISTS vec_digests USING vec0(
    digest_id TEXT PRIMARY KEY,
    project TEXT partition key,
    embedding float[768] distance_metric=cosine
);

-- Backfill embeddings already persisted as BLOB in digests. LENGTH = 3072
-- bytes guards against rows with malformed/legacy embeddings of a different
-- dimension — sqlite-vec rejects mismatched lengths at insert time, so the
-- filter prevents the migration from failing on stray rows.
INSERT INTO vec_digests(digest_id, project, embedding)
SELECT id, project, embedding
FROM digests
WHERE embedding IS NOT NULL
  AND LENGTH(embedding) = 3072;
