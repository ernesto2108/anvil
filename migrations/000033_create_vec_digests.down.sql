-- Rolling back the vec_digests table is safe: digests.embedding (the BLOB
-- source-of-truth) is untouched, so SearchSimilar can fall back to the
-- in-Go cosine path that existed before this migration.
DROP TABLE IF EXISTS vec_digests;
