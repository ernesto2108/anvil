# Reglas de Arquitectura

## Layout del Workspace

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

## Cargo.toml del Workspace

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

## Crate Miembro Hereda del Workspace

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

1. **`main.rs` es delgado** — solo cableado, parsear args, llamar `lib::run()`
2. **`lib.rs` tiene toda la lógica** — testeable sin ejecutar el binario
3. **Separar crates bin y lib** en el workspace para cualquier cosa no trivial

## Feature Flags

4. **`default = []`** — mínimo por defecto, opt-in a features
5. **Feature `std`** para soporte no_std — `#![cfg_attr(not(feature = "std"), no_std)]`
6. **Deps con feature gate** — `[dependencies] postgres = { version = "0.19", optional = true }`

```toml
[features]
default = ["native-tls"]
native-tls = ["reqwest/native-tls"]
rustls = ["reqwest/rustls-tls"]
std = ["serde/std", "hex/std"]
```

## Higiene de Dependencias

7. **Deps del workspace** — centralizar versiones en `[workspace.dependencies]`
8. **Pins exactos para crypto** — `ring = "=0.17.8"`
9. **Sin deps wildcard** — denegar via `cargo-deny`
10. **Commitear Cargo.lock para binarios** — no commitear para librerías
