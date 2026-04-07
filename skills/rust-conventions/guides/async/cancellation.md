# Cancellation Safety & Shutdown

## select! Safety

```rust
// WRONG — read_buf is NOT cancellation safe
loop {
    tokio::select! {
        n = reader.read_buf(&mut buf) => { /* partial read lost on cancel */ }
        _ = tokio::time::sleep(timeout) => break,
    }
}

// RIGHT — use cancellation-safe methods
loop {
    tokio::select! {
        Some(msg) = rx.recv() => handle(msg),  // recv is safe
        _ = shutdown.recv() => break,
    }
}

// RIGHT — pin! for reusable futures
let sleep = tokio::time::sleep(Duration::from_secs(30));
tokio::pin!(sleep);
loop {
    tokio::select! {
        Some(msg) = rx.recv() => handle(msg),
        _ = &mut sleep => { break; }
    }
}
```

## Cancellation-Safe Methods

| Safe | Unsafe |
|------|--------|
| `mpsc::Receiver::recv()` | `AsyncReadExt::read_buf()` |
| `oneshot::Receiver::recv()` | `AsyncBufReadExt::read_line()` |
| `broadcast::Receiver::recv()` | `tokio::io::split` read half |
| `TcpListener::accept()` | Partial reads/writes |
| `JoinSet::join_next()` | `BufWriter::write` without flush |

## Graceful Shutdown

```rust
use tokio::signal;
use tokio::sync::broadcast;

#[tokio::main]
async fn main() -> Result<()> {
    let (shutdown_tx, _) = broadcast::channel(1);

    let server = tokio::spawn({
        let mut rx = shutdown_tx.subscribe();
        async move {
            loop {
                tokio::select! {
                    conn = listener.accept() => handle(conn?).await,
                    _ = rx.recv() => {
                        tracing::info!("shutting down server");
                        break;
                    }
                }
            }
        }
    });

    signal::ctrl_c().await?;
    tracing::info!("shutdown signal received");
    let _ = shutdown_tx.send(());
    server.await??;
    Ok(())
}
```

## Timeouts

```rust
use tokio::time::{timeout, Duration};

// RIGHT — timeout on any async operation
let result = timeout(Duration::from_secs(5), fetch(url)).await;
match result {
    Ok(Ok(data)) => process(data),
    Ok(Err(e)) => tracing::error!("fetch error: {e}"),
    Err(_) => tracing::warn!("fetch timed out"),
}
```
