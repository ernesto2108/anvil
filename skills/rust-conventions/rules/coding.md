# Coding Rules

## Edition 2024 (Rust 1.85+)

1. **`unsafe_op_in_unsafe_fn` warns** — unsafe ops inside `unsafe fn` require explicit `unsafe {}` block
2. **`static mut` denied** — use `AtomicXxx`, `Mutex`, or `LazyLock` instead
3. **`unsafe extern` required** — FFI blocks must be `unsafe extern "C"` with safe/unsafe annotations
4. **`gen` is reserved** — use `r#gen` if you need it as an identifier
5. **`set_var`/`remove_var` are unsafe** — env mutation is now `unsafe` due to thread safety
6. **Resolver v3 default** — `resolver = "3"` in workspace Cargo.toml
7. **RPIT captures all lifetimes** — use `+ use<>` to opt out of implicit lifetime capture

```rust
// WRONG — static mut
static mut COUNTER: u64 = 0;

// RIGHT — atomic
use std::sync::atomic::{AtomicU64, Ordering};
static COUNTER: AtomicU64 = AtomicU64::new(0);

// RIGHT — LazyLock for complex init (1.80+)
use std::sync::LazyLock;
static CONFIG: LazyLock<Config> = LazyLock::new(|| Config::load());
```

## Error Handling

8. **thiserror for libraries** — derive Error for domain types, never expose anyhow
9. **anyhow for applications** — ergonomic `Result<T>`, `.context()` for wrapping
10. **miette for CLIs** — rich diagnostic errors with source spans and help text
11. **Never panic in library code** — return `Result`, no `.unwrap()`, no `panic!()`
12. **Context on every `?`** — `.context("what failed")` or `.with_context(|| format!(...))`

```rust
// WRONG — bare ? with no context
let data = std::fs::read_to_string(path)?;

// RIGHT — context wrapping
let data = std::fs::read_to_string(path)
    .with_context(|| format!("failed to read {path}"))?;
```

## Ownership & Borrowing

13. **`&str` over `String` in params** — accept borrows, return owned only when needed
14. **`Cow<'_, str>`** for conditional ownership — avoids allocation when not needed
15. **`impl AsRef<T>` for flexible APIs** — accept both owned and borrowed
16. **Clone only when necessary** — prefer borrows, use `Arc` for shared ownership in async

```rust
// WRONG — takes ownership unnecessarily
fn process(data: String) -> String { data.to_uppercase() }

// RIGHT — borrows input
fn process(data: &str) -> String { data.to_uppercase() }
```

## Naming

17. **snake_case** for functions, methods, variables, modules, crates
18. **PascalCase** for types, traits, enums, structs
19. **SCREAMING_SNAKE** for constants and statics
20. **`is_`, `has_`, `can_`** prefix for bool-returning methods
21. **`into_`, `as_`, `to_`** for conversions (ownership-taking, borrowing, expensive respectively)
22. **`try_`** prefix for fallible operations that return `Result`

## Async Traits (1.75+)

23. **Native async fn in traits** — no more `#[async_trait]` proc macro needed
24. **`-> impl Future + Send`** when you need Send bound for trait objects
25. **`Pin<Box<dyn Future + Send>>`** when you need dyn dispatch

```rust
// WRONG — async-trait crate (legacy)
#[async_trait]
trait Repo { async fn find(&self, id: Uuid) -> Result<Entity>; }

// RIGHT — native (1.75+)
trait Repo: Send + Sync {
    async fn find(&self, id: Uuid) -> Result<Entity>;
}
```
