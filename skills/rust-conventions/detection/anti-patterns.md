# Rust Anti-Patterns — Detection Reference

## Anti-Pattern Table

| Code Pattern | Anti-Pattern | Severity | Category | Fix |
|---|---|---|---|---|
| `.unwrap()` in lib code | unwrap-in-lib | error | reliability | Return `Result`, use `?` |
| `panic!()` in lib code | panic-in-lib | error | reliability | Return `Result` |
| `unsafe` without `// SAFETY:` | undocumented-unsafe | error | safety | Document invariant |
| `static mut` | static-mut | error | safety | `AtomicXxx`, `Mutex`, or `LazyLock` |
| `#[allow(warnings)]` | blanket-allow | error | quality | Allow specific lints only |
| `dbg!()` in committed code | debug-macro | error | quality | Use `tracing::debug!` |
| `todo!()` in main branch | todo-macro | error | quality | Implement or track in issue |
| `String` param where `&str` works | unnecessary-owned | warning | performance | Accept `&str` or `impl AsRef<str>` |
| `.clone()` without need | unnecessary-clone | warning | performance | Borrow or restructure |
| `.to_string()` in hot path | hot-path-alloc | warning | performance | `Cow`, `&str`, or pre-allocate |
| `Vec` growing in loop | vec-push-loop | warning | performance | Pre-allocate with `Vec::with_capacity` |
| `Box<dyn Error>` in library API | boxed-error-lib | warning | types | `thiserror` enum |
| `impl Display` missing for error types | error-no-display | warning | types | Derive with `thiserror` |
| Bare `extern "C"` (edition 2024) | bare-extern | error | edition | `unsafe extern "C"` |
| `std::env::set_var` without unsafe (2024) | unsafe-env-set | error | edition | Wrap in `unsafe {}` |
| Large `unsafe` block | wide-unsafe | warning | safety | Minimize scope per operation |
| `#[async_trait]` on Rust 1.75+ | legacy-async-trait | suggestion | modernization | Native async fn in traits |
| `lazy_static!` on Rust 1.80+ | legacy-lazy-static | suggestion | modernization | `std::sync::LazyLock` |
| `once_cell` on Rust 1.80+ | legacy-once-cell | suggestion | modernization | `std::sync::LazyLock`/`LazyCell` |
| `expected == received` for secrets | timing-attack | error | security | `subtle::ConstantTimeEq` |
| Secret in `Debug` output | secret-in-debug | warning | security | Custom `Debug` impl with redaction |
| `println!` in lib/server code | print-in-lib | warning | quality | Use `tracing` |
| Wildcard deps `*` | wildcard-dep | error | supply-chain | Pin exact or range version |
| Missing `Cargo.lock` in binary crate | no-lockfile | warning | reproducibility | Commit `Cargo.lock` |
| `as` cast without check | unchecked-cast | warning | correctness | `try_into()` or clippy cast lints |
| Blocking IO in async context | blocking-in-async | error | reliability | `tokio::fs`, `spawn_blocking` |
| `select!` with non-cancel-safe future | cancel-unsafe-select | error | reliability | Use cancel-safe methods |
| Fire-and-forget `tokio::spawn` | untracked-spawn | warning | reliability | Use `JoinSet` |
