---
name: rust-conventions
description: Rust conventions for systems programming, high-performance CLIs, and blockchain/crypto. Use when writing Rust code, reviewing Rust patterns, or user mentions "rust conventions", "tokio", "async rust", "clap", "solana", "substrate", "no_std", "unsafe", or working with .rs files.
---

# Rust Conventions

> **IMPORTANT:** This file is a lightweight dispatcher. Do NOT load all referenced files at once. Read the routing table below, identify which files are relevant to the current task, and load ONLY those using the Read tool. Each file is ~3-5KB. Loading unnecessary files wastes context tokens.

## Stack & Philosophy

- **Edition 2024** (Rust 1.85+) — `unsafe_op_in_unsafe_fn` warns, `static mut` denied, `unsafe extern`, resolver v3
- **Ownership over GC** — zero-copy when possible, Cow for conditional ownership, arena allocation for batches
- **Errors are types** — thiserror for libraries, anyhow for applications, never panic in library code
- **Async with tokio** — native async traits (1.75+), JoinSet for structured concurrency, cancellation safety
- **No unnecessary unsafe** — `#![forbid(unsafe_code)]` at crate level, allow only in audited modules

## Red Flags (always stop work)

- `.unwrap()` in library code → error (use `?` or return Result)
- `panic!()` in library code → error
- `unsafe` without `// SAFETY:` comment → error
- `static mut` → error (use AtomicXxx, Mutex, or LazyLock)
- `#[allow(warnings)]` → error
- `dbg!()` or `todo!()` in committed code → error
- `String` where `&str` suffices → warning (unnecessary allocation)

## Anti-Pattern Detection

**Passive detection:** When reviewing Rust code, load `detection/anti-patterns.md` and scan for `error` and `warning` patterns. Report as `[file:line] [severity] [category] anti-pattern-name`.

**Active detection:** When user asks to "improve", "refactor", "optimize", or "clean" — also report `suggestion` level patterns.

## What to Load

Load **only** the files relevant to the current task:

### Rules (quick reference, ~2-3KB each)

| Working on... | Load |
|---|---|
| Ownership, error handling, naming, edition 2024 features | `rules/coding.md` |
| Workspace layout, Cargo.toml, feature flags, lib vs bin | `rules/architecture.md` |

### Guides (detailed patterns with code, ~3-5KB each)

| Working on... | Load |
|---|---|
| tokio runtime, async traits, JoinSet, select!, channels | `guides/async/tokio.md` |
| Cancellation safety, graceful shutdown, timeouts | `guides/async/cancellation.md` |
| clap v4 derive, tracing, indicatif, cross-compilation | `guides/cli/patterns.md` |
| Solana/Anchor, Alloy/Ethereum, crypto primitives, no_std | `guides/blockchain/patterns.md` |
| Zero-copy, arena allocation, Cow, SmallVec, Pin/Unpin | `guides/performance/memory.md` |
| Unsafe guidelines, cargo-audit, cargo-deny, constant-time | `guides/performance/safety.md` |
| Unit tests, integration tests, proptest, criterion, insta | `guides/testing/patterns.md` |
| Clippy config, rustfmt, CI pipeline, cargo-machete | `guides/tooling.md` |

### Detection & Checklists

| When... | Load |
|---|---|
| Code review | `detection/anti-patterns.md` |
| Before writing Rust code | `checklists/pre.md` |
| After writing Rust code | `checklists/post.md` |

## Post-Implementation Gate

After ANY code change to `.rs` files, run `cargo fmt --check && cargo clippy -- -D warnings` before considering the task done.
