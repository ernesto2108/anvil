# Pre-Implementation Checklist

- [ ] Edition 2024? — `edition = "2024"` in Cargo.toml, resolver = "3"
- [ ] Error strategy — thiserror for lib, anyhow for app, miette for CLI
- [ ] Ownership plan — who owns the data? Borrow where possible, `Arc` for shared async
- [ ] Async or sync? — if async, tokio runtime configured, no blocking in async context
- [ ] `#![forbid(unsafe_code)]` at crate level — allow only in audited modules
- [ ] Workspace deps — centralized in `[workspace.dependencies]`
- [ ] For CLI — clap derive, tracing for logging, indicatif for progress
- [ ] For blockchain — anchor/alloy patterns, constant-time for secrets, no_std if needed
- [ ] Test strategy — unit + integration, proptest for invariants, criterion for perf
