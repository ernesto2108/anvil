# Tokio Async Patterns

## Runtime Setup

```rust
// RIGHT — application entry point
#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::fmt().with_env_filter("info").init();
    run().await
}

// RIGHT — explicit config for servers
fn main() -> anyhow::Result<()> {
    let runtime = tokio::runtime::Builder::new_multi_thread()
        .worker_threads(4)
        .enable_all()
        .build()?;
    runtime.block_on(run())
}
```

## Blocking Work

```rust
// WRONG — blocking IO in async context
async fn read_file(path: &str) -> Result<String> {
    std::fs::read_to_string(path).context("read failed")  // blocks!
}

// RIGHT — use tokio async IO
async fn read_file(path: &str) -> Result<String> {
    tokio::fs::read_to_string(path).await.context("read failed")
}

// RIGHT — spawn_blocking for CPU-bound work
let hash = tokio::task::spawn_blocking(move || {
    compute_hash(&large_data)
}).await?;
```

## JoinSet (Structured Concurrency)

```rust
use tokio::task::JoinSet;

let mut set = JoinSet::new();
for url in urls {
    set.spawn(fetch(url));
}

while let Some(result) = set.join_next().await {
    match result {
        Ok(Ok(data)) => process(data),
        Ok(Err(e)) => tracing::error!("task failed: {e}"),
        Err(e) => tracing::error!("task panicked: {e}"),
    }
}
```

## Bounded Concurrency

```rust
use tokio::sync::Semaphore;
use std::sync::Arc;

let sem = Arc::new(Semaphore::new(10));
let mut set = JoinSet::new();

for url in urls {
    let permit = sem.clone().acquire_owned().await?;
    set.spawn(async move {
        let result = fetch(url).await;
        drop(permit);
        result
    });
}
```

## Channels

```rust
// mpsc — multiple producers, single consumer
let (tx, mut rx) = tokio::sync::mpsc::channel(100);

// oneshot — single request-response
let (tx, rx) = tokio::sync::oneshot::channel();

// broadcast — pub/sub
let (tx, _) = tokio::sync::broadcast::channel(100);
let mut rx = tx.subscribe();

// watch — latest-value (config reloading)
let (tx, rx) = tokio::sync::watch::channel(initial_config);
```

## Native Async Traits (1.75+)

```rust
// RIGHT — no #[async_trait] needed
trait Repository: Send + Sync {
    async fn find(&self, id: Uuid) -> Result<Option<Entity>>;
    async fn save(&self, entity: &Entity) -> Result<()>;
}

// When you need dyn dispatch
trait Repository: Send + Sync {
    fn find(&self, id: Uuid) -> Pin<Box<dyn Future<Output = Result<Option<Entity>>> + Send + '_>>;
}
```
