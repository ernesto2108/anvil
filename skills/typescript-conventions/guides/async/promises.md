# Async & Promises Guide

## AbortController for Cancellation

Every network request, long computation, or streaming operation must accept an `AbortSignal`. Never fire-and-forget without a cancellation path.

```typescript
// WRONG: no cancellation — request runs forever
async function fetchUser(id: string): Promise<User> {
  const res = await fetch(`/api/users/${id}`);
  return res.json();
}

// RIGHT: accept signal, pass to fetch
async function fetchUser(id: string, signal?: AbortSignal): Promise<User> {
  const res = await fetch(`/api/users/${id}`, { signal });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json() as Promise<User>;
}

// Caller creates and manages the controller
const controller = new AbortController();
const timeoutId = setTimeout(() => controller.abort(), 5_000);

try {
  const user = await fetchUser(id, controller.signal);
} catch (e) {
  if (e instanceof DOMException && e.name === "AbortError") {
    // Request was cancelled — handle gracefully
  }
  throw e;
} finally {
  clearTimeout(timeoutId);
}
```

### Timeout helper with AbortSignal

```typescript
// AbortSignal.timeout() — available in Node 17.3+ and modern browsers
async function fetchWithTimeout(url: string, ms: number): Promise<Response> {
  const signal = AbortSignal.timeout(ms);
  return fetch(url, { signal });
}

// Combine multiple signals (e.g., component unmount + timeout)
function combineSignals(...signals: AbortSignal[]): AbortSignal {
  const controller = new AbortController();
  for (const signal of signals) {
    if (signal.aborted) {
      controller.abort(signal.reason);
      break;
    }
    signal.addEventListener("abort", () => controller.abort(signal.reason), {
      signal: controller.signal,
    });
  }
  return controller.signal;
}
```

## `Promise.allSettled` vs `Promise.all`

Use `Promise.all` only when all operations MUST succeed together. Use `Promise.allSettled` when partial success is acceptable.

```typescript
// WRONG: Promise.all fails fast — one rejection cancels everything
const [users, orders, products] = await Promise.all([
  fetchUsers(),    // if this throws...
  fetchOrders(),   // ...this result is lost even if it succeeded
  fetchProducts(),
]);

// RIGHT for independent operations: Promise.allSettled
const results = await Promise.allSettled([fetchUsers(), fetchOrders(), fetchProducts()]);

const users    = results[0].status === "fulfilled" ? results[0].value : [];
const orders   = results[1].status === "fulfilled" ? results[1].value : [];
const products = results[2].status === "fulfilled" ? results[2].value : [];

// Log partial failures
results.forEach((r, i) => {
  if (r.status === "rejected") {
    console.error(`Operation ${i} failed:`, r.reason);
  }
});

// Helper for typed allSettled
async function settleAll<T extends readonly unknown[]>(
  promises: { [K in keyof T]: Promise<T[K]> }
): Promise<{ [K in keyof T]: T[K] | null }> {
  const results = await Promise.allSettled(promises);
  return results.map((r) => (r.status === "fulfilled" ? r.value : null)) as any;
}
```

## Async Iterators

Use `for await...of` over async generators for streaming data (paginated APIs, file streams, event sources).

```typescript
// Async generator — yields one page at a time
async function* paginateUsers(
  pageSize: number,
  signal?: AbortSignal
): AsyncGenerator<User[]> {
  let cursor: string | null = null;

  while (true) {
    signal?.throwIfAborted();

    const response = await fetchUserPage({ cursor, pageSize, signal });
    yield response.data;

    if (!response.nextCursor) break;
    cursor = response.nextCursor;
  }
}

// Caller — process users without loading all into memory
for await (const page of paginateUsers(100, signal)) {
  await processPage(page);
}

// Pipeline: transform async iterables
async function* map<T, U>(
  source: AsyncIterable<T>,
  fn: (item: T) => Promise<U>
): AsyncGenerator<U> {
  for await (const item of source) {
    yield fn(item);
  }
}
```

## Error Handling in Async Code

```typescript
// WRONG: unhandled rejection — silently swallowed
async function syncData() {
  fetchUsers().then(saveUsers); // no .catch()
}

// WRONG: catching unknown as any
try {
  await fetchUsers();
} catch (e: any) { // any strips type safety
  console.error(e.message); // crashes if e is not an Error
}

// RIGHT: catch unknown, narrow before use
async function fetchSafely<T>(fn: () => Promise<T>): Promise<T | null> {
  try {
    return await fn();
  } catch (e: unknown) {
    if (e instanceof DOMException && e.name === "AbortError") {
      return null; // expected cancellation
    }
    if (e instanceof TypeError) {
      console.error("Network error:", e.message);
      return null;
    }
    throw e; // re-throw unexpected errors
  }
}

// RIGHT: always attach rejection handlers
const userPromise = fetchUsers();
userPromise.catch(console.error); // even if you don't use the result yet

// Using Promise.race for timeout
function withTimeout<T>(promise: Promise<T>, ms: number): Promise<T> {
  const timeout = new Promise<never>((_, reject) =>
    setTimeout(() => reject(new Error(`Timed out after ${ms}ms`)), ms)
  );
  return Promise.race([promise, timeout]);
}
```

## Concurrent Execution Patterns

```typescript
// Sequential — when order or dependencies matter
for (const user of users) {
  await processUser(user); // each waits for the previous
}

// Concurrent (unbounded) — use only for small, known-size arrays
await Promise.all(users.map(processUser));

// Concurrent with concurrency limit — for large arrays or rate-limited APIs
async function* chunks<T>(arr: T[], size: number): AsyncGenerator<T[]> {
  for (let i = 0; i < arr.length; i += size) {
    yield arr.slice(i, i + size);
  }
}

async function processWithConcurrency<T>(
  items: T[],
  fn: (item: T) => Promise<void>,
  concurrency = 5
): Promise<void> {
  for await (const batch of chunks(items, concurrency)) {
    await Promise.all(batch.map(fn));
  }
}
```
