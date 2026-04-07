# Architecture Rules

## Workspace Layout

```
my-project/
  Cargo.toml              # workspace root
  crates/
    core/                  # domain logic, no IO, no_std optional
    cli/                   # binary crate
    server/                # optional server binary
    sdk/                   # public API crate
  tests/                   # workspace-level integration tests
  benches/                 # workspace-level benchmarks
```

## Workspace Cargo.toml

```toml
[workspace]
members = ["crates/*"]
resolver = "3"

[workspace.package]
version = "0.1.0"
edition = "2024"
rust-version = "1.85"

[workspace.dependencies]
tokio = { version = "1", features = ["full"] }
serde = { version = "1", features = ["derive"] }
anyhow = "1"
thiserror = "2"
tracing = "0.1"

[workspace.lints.rust]
unsafe_code = "forbid"

[workspace.lints.clippy]
all = { level = "deny", priority = -1 }
pedantic = { level = "warn", priority = -1 }
unwrap_used = "deny"
```

## Member Crate Inherits Workspace

```toml
[package]
name = "my-core"
version.workspace = true
edition.workspace = true

[dependencies]
serde.workspace = true
thiserror.workspace = true

[lints]
workspace = true
```

## lib.rs vs main.rs

1. **`main.rs` is thin** — only wiring, parse args, call `lib::run()`
2. **`lib.rs` has all logic** — testable without running the binary
3. **Separate bin and lib crates** in workspace for anything non-trivial

## Feature Flags

4. **`default = []`** — minimal by default, opt-in to features
5. **`std` feature** for no_std support — `#![cfg_attr(not(feature = "std"), no_std)]`
6. **Feature-gated deps** — `[dependencies] postgres = { version = "0.19", optional = true }`

```toml
[features]
default = ["native-tls"]
native-tls = ["reqwest/native-tls"]
rustls = ["reqwest/rustls-tls"]
std = ["serde/std", "hex/std"]
```

## Dependency Hygiene

7. **Workspace deps** — centralize versions in `[workspace.dependencies]`
8. **Exact pins for crypto** — `ring = "=0.17.8"`
9. **No wildcard deps** — deny via `cargo-deny`
10. **Commit Cargo.lock for binaries** — don't commit for libraries
