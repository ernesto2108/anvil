# Guía de Async y Promises

## AbortController para Cancelación

Toda solicitud de red, cómputo prolongado u operación de streaming debe aceptar un `AbortSignal`. Nunca lanzar operaciones sin una ruta de cancelación.

```typescript
// INCORRECTO: sin cancelación — la solicitud se ejecuta indefinidamente
async function fetchUser(id: string): Promise<User> {
  const res = await fetch(`/api/users/${id}`);
  return res.json();
}

// CORRECTO: aceptar signal, pasarlo a fetch
async function fetchUser(id: string, signal?: AbortSignal): Promise<User> {
  const res = await fetch(`/api/users/${id}`, { signal });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json() as Promise<User>;
}

// El caller crea y gestiona el controller
const controller = new AbortController();
const timeoutId = setTimeout(() => controller.abort(), 5_000);

try {
  const user = await fetchUser(id, controller.signal);
} catch (e) {
  if (e instanceof DOMException && e.name === "AbortError") {
    // La solicitud fue cancelada — manejar con gracia
  }
  throw e;
} finally {
  clearTimeout(timeoutId);
}
```

### Helper de timeout con AbortSignal

```typescript
// AbortSignal.timeout() — disponible en Node 17.3+ y navegadores modernos
async function fetchWithTimeout(url: string, ms: number): Promise<Response> {
  const signal = AbortSignal.timeout(ms);
  return fetch(url, { signal });
}

// Combinar múltiples signals (e.g., desmontaje de componente + timeout)
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

Usar `Promise.all` solo cuando todas las operaciones DEBEN tener éxito en conjunto. Usar `Promise.allSettled` cuando el éxito parcial es aceptable.

```typescript
// INCORRECTO: Promise.all falla rápido — un rechazo cancela todo
const [users, orders, products] = await Promise.all([
  fetchUsers(),    // si esto lanza...
  fetchOrders(),   // ...este resultado se pierde aunque haya tenido éxito
  fetchProducts(),
]);

// CORRECTO para operaciones independientes: Promise.allSettled
const results = await Promise.allSettled([fetchUsers(), fetchOrders(), fetchProducts()]);

const users    = results[0].status === "fulfilled" ? results[0].value : [];
const orders   = results[1].status === "fulfilled" ? results[1].value : [];
const products = results[2].status === "fulfilled" ? results[2].value : [];

// Registrar fallos parciales
results.forEach((r, i) => {
  if (r.status === "rejected") {
    console.error(`Operation ${i} failed:`, r.reason);
  }
});

// Helper para allSettled tipado
async function settleAll<T extends readonly unknown[]>(
  promises: { [K in keyof T]: Promise<T[K]> }
): Promise<{ [K in keyof T]: T[K] | null }> {
  const results = await Promise.allSettled(promises);
  return results.map((r) => (r.status === "fulfilled" ? r.value : null)) as any;
}
```

## Iteradores Asíncronos

Usar `for await...of` sobre generadores asíncronos para datos en streaming (APIs paginadas, streams de archivos, event sources).

```typescript
// Generador asíncrono — produce una página a la vez
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

// Caller — procesar usuarios sin cargar todo en memoria
for await (const page of paginateUsers(100, signal)) {
  await processPage(page);
}

// Pipeline: transformar iterables asíncronos
async function* map<T, U>(
  source: AsyncIterable<T>,
  fn: (item: T) => Promise<U>
): AsyncGenerator<U> {
  for await (const item of source) {
    yield fn(item);
  }
}
```

## Manejo de Errores en Código Asíncrono

```typescript
// INCORRECTO: rechazo no manejado — silenciosamente ignorado
async function syncData() {
  fetchUsers().then(saveUsers); // sin .catch()
}

// INCORRECTO: capturar unknown como any
try {
  await fetchUsers();
} catch (e: any) { // any elimina la seguridad de tipos
  console.error(e.message); // falla si e no es un Error
}

// CORRECTO: capturar unknown, estrechar antes de usar
async function fetchSafely<T>(fn: () => Promise<T>): Promise<T | null> {
  try {
    return await fn();
  } catch (e: unknown) {
    if (e instanceof DOMException && e.name === "AbortError") {
      return null; // cancelación esperada
    }
    if (e instanceof TypeError) {
      console.error("Network error:", e.message);
      return null;
    }
    throw e; // re-lanzar errores inesperados
  }
}

// CORRECTO: siempre adjuntar handlers de rechazo
const userPromise = fetchUsers();
userPromise.catch(console.error); // incluso si aún no usas el resultado

// Usar Promise.race para timeout
function withTimeout<T>(promise: Promise<T>, ms: number): Promise<T> {
  const timeout = new Promise<never>((_, reject) =>
    setTimeout(() => reject(new Error(`Timed out after ${ms}ms`)), ms)
  );
  return Promise.race([promise, timeout]);
}
```

## Patrones de Ejecución Concurrente

```typescript
// Secuencial — cuando el orden o las dependencias importan
for (const user of users) {
  await processUser(user); // cada uno espera al anterior
}

// Concurrente (sin límite) — usar solo para arrays pequeños de tamaño conocido
await Promise.all(users.map(processUser));

// Concurrente con límite de concurrencia — para arrays grandes o APIs con rate limit
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
