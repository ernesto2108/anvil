# Patrones Async

## TaskGroup (Concurrencia Estructurada, Python 3.11+)

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

## Manejadores de Contexto Async

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

## Iteradores Async para Streaming

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

# Uso
async for embedding in stream_embeddings(texts):
    store(embedding)
```

## Semáforo para Limitación de Velocidad

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

## Reglas Clave

- **TaskGroup sobre gather** — concurrencia estructurada, limpieza automática
- **`asyncio.timeout()` sobre `wait_for`** — más limpio y composable
- **Nunca `time.sleep()` en async** — usar `await asyncio.sleep()`
- **Nunca `requests.get()` en async** — usar `aiohttp` o `httpx`
- **Manejadores de contexto async para recursos** — limpieza garantizada
