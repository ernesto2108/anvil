# Embeddings & ML Patterns

## Batch Embedding Computation

```python
from itertools import batched
import numpy as np
from numpy.typing import NDArray

def compute_embeddings(
    texts: list[str],
    model: EmbeddingModel,
    dim: int = 768,
    batch_size: int = 256,
) -> NDArray[np.float32]:
    n = len(texts)
    result = np.empty((n, dim), dtype=np.float32)

    for i, batch in enumerate(batched(texts, batch_size)):
        start = i * batch_size
        result[start:start + len(batch)] = model.encode(list(batch))

    return result
```

## Cosine Similarity (Vectorized)

```python
def cosine_similarity(
    query: NDArray[np.float32],
    corpus: NDArray[np.float32],
) -> NDArray[np.float32]:
    """Batch cosine similarity: query (dim,) vs corpus (n, dim)."""
    query_norm = np.linalg.norm(query)
    corpus_norms = np.linalg.norm(corpus, axis=1)
    return corpus @ query / (corpus_norms * query_norm + 1e-12)
```

## Normalization

```python
def normalize(vectors: NDArray[np.float32]) -> NDArray[np.float32]:
    """L2-normalize vectors. Shape: (n, dim) -> (n, dim)."""
    norms = np.linalg.norm(vectors, axis=1, keepdims=True)
    return vectors / np.maximum(norms, 1e-12)
```

## Top-K Retrieval

```python
def top_k(
    query: NDArray[np.float32],
    corpus: NDArray[np.float32],
    k: int = 10,
) -> tuple[NDArray[np.int64], NDArray[np.float32]]:
    """Return top-k indices and scores."""
    scores = cosine_similarity(query, corpus)
    # argpartition is O(n) vs O(n log n) for argsort
    top_indices = np.argpartition(scores, -k)[-k:]
    top_indices = top_indices[np.argsort(scores[top_indices])[::-1]]
    return top_indices, scores[top_indices]
```

## Memory-Mapped Storage

```python
def create_store(path: str, n: int, dim: int) -> np.memmap:
    """Create a new memory-mapped embedding store."""
    return np.memmap(path, dtype="float32", mode="w+", shape=(n, dim))

def load_store(path: str, n: int, dim: int) -> np.memmap:
    """Load existing memory-mapped embedding store (read-only)."""
    return np.memmap(path, dtype="float32", mode="r", shape=(n, dim))
```

## Async Batch Processing

```python
async def embed_async(
    texts: list[str],
    client: AsyncEmbeddingClient,
    batch_size: int = 64,
    max_concurrent: int = 4,
) -> NDArray[np.float32]:
    """Async embedding with controlled concurrency."""
    semaphore = asyncio.Semaphore(max_concurrent)
    results: dict[int, NDArray[np.float32]] = {}

    async def process_batch(idx: int, batch: list[str]) -> None:
        async with semaphore:
            results[idx] = await client.embed(batch)

    async with asyncio.TaskGroup() as tg:
        for i, batch in enumerate(batched(texts, batch_size)):
            tg.create_task(process_batch(i, list(batch)))

    return np.vstack([results[i] for i in sorted(results)])
```

## Key Rules

- **Always `float32`** — float64 doubles memory for zero benefit in similarity search
- **Preallocate with `np.empty`** — never grow lists
- **`argpartition` for top-k** — O(n) vs O(n log n) for full sort
- **Batch API calls** — 32-256 items per batch, never one-by-one
- **Semaphore for concurrency** — limit parallel API calls to avoid rate limits
- **Memory-map for large stores** — datasets that don't fit in RAM
