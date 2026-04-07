# Async Patterns

## TaskGroup (Structured Concurrency, Python 3.11+)

```python
# WRONG — gather leaks tasks on failure
async def fetch_all(urls: list[str]) -> list[str]:
    tasks = [asyncio.create_task(fetch(url)) for url in urls]
    return await asyncio.gather(*tasks)

# RIGHT — TaskGroup guarantees cleanup
async def fetch_all(urls: list[str]) -> list[str]:
    results: list[str] = []
    async with asyncio.TaskGroup() as tg:
        for url in urls:
            tg.create_task(fetch_and_collect(url, results))
    return results
```

## Timeouts (Python 3.11+)

```python
# WRONG — wait_for (deprecated pattern)
result = await asyncio.wait_for(fetch(url), timeout=10)

# RIGHT — asyncio.timeout context manager
async with asyncio.timeout(10.0):
    result = await fetch(url)

# RIGHT — timeout with fallback
async with asyncio.timeout_at(deadline):
    try:
        result = await fetch(url)
    except TimeoutError:
        result = cached_result
```

## Async Context Managers

```python
from contextlib import asynccontextmanager

@asynccontextmanager
async def db_pool(dsn: str):
    pool = await create_pool(dsn)
    try:
        yield pool
    finally:
        await pool.close()

async with db_pool("postgres://...") as pool:
    await pool.execute("SELECT 1")
```

## Async Iterators for Streaming

```python
from collections.abc import AsyncIterator
from itertools import batched

async def stream_embeddings(
    texts: list[str],
    batch_size: int = 32,
) -> AsyncIterator[list[float]]:
    for batch in batched(texts, batch_size):
        embeddings = await compute_embeddings(batch)
        for emb in embeddings:
            yield emb

# Usage
async for embedding in stream_embeddings(texts):
    store(embedding)
```

## Semaphore for Rate Limiting

```python
async def fetch_all_limited(urls: list[str], max_concurrent: int = 10) -> list[str]:
    semaphore = asyncio.Semaphore(max_concurrent)

    async def fetch_one(url: str) -> str:
        async with semaphore:
            return await fetch(url)

    async with asyncio.TaskGroup() as tg:
        tasks = [tg.create_task(fetch_one(url)) for url in urls]
    return [t.result() for t in tasks]
```

## Key Rules

- **TaskGroup over gather** — structured concurrency, automatic cleanup
- **`asyncio.timeout()` over `wait_for`** — cleaner, composable
- **Never `time.sleep()` in async** — use `await asyncio.sleep()`
- **Never `requests.get()` in async** — use `aiohttp` or `httpx`
- **Async context managers for resources** — guaranteed cleanup
