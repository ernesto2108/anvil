---
name: rust-conventions
description: Convenciones Rust para programación de sistemas, CLIs de alto rendimiento y blockchain/crypto. Usar al escribir código Rust, revisar patrones Rust, o cuando el usuario mencione "rust conventions", "tokio", "async rust", "clap", "solana", "substrate", "no_std", "unsafe", o al trabajar con archivos .rs.
---

# Rust Conventions

> **IMPORTANTE:** Este archivo es un dispatcher ligero. NO cargar todos los archivos referenciados a la vez. Leer la tabla de enrutamiento abajo, identificar qué archivos son relevantes para la tarea actual y cargar SOLO esos usando la herramienta Read. Cada archivo pesa ~3-5KB. Cargar archivos innecesarios desperdicia tokens de contexto.

## Stack y Filosofía

- **Edition 2024** (Rust 1.85+) — `unsafe_op_in_unsafe_fn` advierte, `static mut` denegado, `unsafe extern`, resolver v3
- **Ownership sobre GC** — zero-copy cuando sea posible, Cow para ownership condicional, arena allocation para batches
- **Los errores son tipos** — thiserror para librerías, anyhow para aplicaciones, nunca panic en código de librería
- **Async con tokio** — async traits nativos (1.75+), JoinSet para concurrencia estructurada, seguridad de cancelación
- **Sin unsafe innecesario** — `#![forbid(unsafe_code)]` a nivel de crate, permitir solo en módulos auditados

## Red Flags (siempre detener el trabajo)

- `.unwrap()` en código de librería → error (usar `?` o retornar Result)
- `panic!()` en código de librería → error
- `unsafe` sin comentario `// SAFETY:` → error
- `static mut` → error (usar AtomicXxx, Mutex o LazyLock)
- `#[allow(warnings)]` → error
- `dbg!()` o `todo!()` en código commiteado → error
- `String` donde `&str` es suficiente → warning (asignación innecesaria)

## Detección de Anti-Patrones

**Detección pasiva:** Al revisar código Rust, cargar `detection/anti-patterns.md` y escanear en busca de patrones `error` y `warning`. Reportar como `[file:line] [severity] [category] anti-pattern-name`.

**Detección activa:** Cuando el usuario pide "mejorar", "refactorizar", "optimizar" o "limpiar" — también reportar patrones de nivel `suggestion`.

## Qué Cargar

Cargar **solo** los archivos relevantes para la tarea actual:

### Rules (referencia rápida, ~2-3KB cada uno)

| Trabajando en... | Cargar |
|---|---|
| Ownership, manejo de errores, naming, features de edition 2024 | `rules/coding.md` |
| Layout del workspace, Cargo.toml, feature flags, lib vs bin | `rules/architecture.md` |

### Guides (patrones detallados con código, ~3-5KB cada uno)

| Trabajando en... | Cargar |
|---|---|
| Runtime tokio, async traits, JoinSet, select!, channels | `guides/async/tokio.md` |
| Seguridad de cancelación, graceful shutdown, timeouts | `guides/async/cancellation.md` |
| clap v4 derive, tracing, indicatif, cross-compilation | `guides/cli/patterns.md` |
| Solana/Anchor, Alloy/Ethereum, crypto primitives, no_std | `guides/blockchain/patterns.md` |
| Zero-copy, arena allocation, Cow, SmallVec, Pin/Unpin | `guides/performance/memory.md` |
| Guías de unsafe, cargo-audit, cargo-deny, constant-time | `guides/performance/safety.md` |
| Tests unitarios, tests de integración, proptest, criterion, insta | `guides/testing/patterns.md` |
| Configuración de Clippy, rustfmt, pipeline CI, cargo-machete | `guides/tooling.md` |

### Detección y Checklists

| Cuándo... | Cargar |
|---|---|
| Revisión de código | `detection/anti-patterns.md` |
| Antes de escribir código Rust | `checklists/pre.md` |
| Después de escribir código Rust | `checklists/post.md` |

## Gate Post-Implementación

Después de CUALQUIER cambio de código en archivos `.rs`, ejecutar `cargo fmt --check && cargo clippy -- -D warnings` antes de considerar la tarea como terminada.
