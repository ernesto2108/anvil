# Post-Implementation Checklist

- [ ] `cargo fmt --check` — formatting correct
- [ ] `cargo clippy -- -D warnings` — no lint errors
- [ ] `cargo test --workspace` — all tests pass
- [ ] No `.unwrap()` or `panic!()` in library code
- [ ] Every `unsafe` block has `// SAFETY:` comment
- [ ] Every `?` has `.context()` or `.with_context()`
- [ ] No blocking IO in async functions
- [ ] `JoinSet` for spawned tasks, not fire-and-forget
- [ ] `select!` only with cancellation-safe futures
- [ ] No `static mut` — using atomics/Mutex/LazyLock
- [ ] `cargo audit` clean — no known vulnerabilities
- [ ] `cargo machete` clean — no unused dependencies
- [ ] For crypto: constant-time comparisons, secrets redacted from Debug
