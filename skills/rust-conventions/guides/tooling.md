# Tooling & Linting

## Clippy Configuration

```toml
# Cargo.toml
[workspace.lints.clippy]
all = { level = "deny", priority = -1 }
pedantic = { level = "warn", priority = -1 }
nursery = { level = "warn", priority = -1 }
unwrap_used = "deny"
expect_used = "warn"
panic = "deny"
todo = "deny"
dbg_macro = "deny"
print_stdout = "warn"
print_stderr = "warn"
cast_possible_truncation = "deny"
cast_sign_loss = "deny"
redundant_clone = "deny"
needless_pass_by_value = "warn"
module_name_repetitions = "allow"
must_use_candidate = "allow"
```

## rustfmt Configuration

```toml
# rustfmt.toml
edition = "2024"
max_width = 100
tab_spaces = 4
use_field_init_shorthand = true
use_try_shorthand = true
imports_granularity = "Crate"
group_imports = "StdExternalCrate"
newline_style = "Unix"
```

## CI Pipeline

```yaml
- name: Format
  run: cargo fmt --all -- --check

- name: Clippy
  run: cargo clippy --workspace --all-targets --all-features -- -D warnings

- name: Audit
  run: cargo audit

- name: Deny
  run: cargo deny check

- name: Unused deps
  run: cargo machete

- name: Test
  run: cargo test --workspace --all-features
```

## cargo-deny (deny.toml)

```toml
[advisories]
vulnerability = "deny"
unmaintained = "warn"
yanked = "deny"

[licenses]
unlicensed = "deny"
allow = ["MIT", "Apache-2.0", "BSD-2-Clause", "BSD-3-Clause", "ISC"]

[bans]
multiple-versions = "warn"
wildcards = "deny"

[sources]
unknown-registry = "deny"
unknown-git = "deny"
```

## Key Tools

- `rustfmt` — formatting (ships with rustup)
- `clippy` — linting (ships with rustup)
- `cargo-audit` — vulnerability scanning (RustSec)
- `cargo-deny` — license/ban/advisory/source checking
- `cargo-machete` — unused dependency detection
- `cargo-vet` — supply chain verification
- `cargo-bloat` — binary size analysis
- `cargo-expand` — macro expansion viewer
- `rust-analyzer` — IDE support (LSP)
