# CLI Patterns

## clap v4 Derive API

```rust
use clap::{Parser, Subcommand, Args, ValueEnum};

#[derive(Parser)]
#[command(name = "mytool", version, about)]
struct Cli {
    #[command(subcommand)]
    command: Commands,

    #[arg(short, long, global = true, action = clap::ArgAction::Count)]
    verbose: u8,

    #[arg(short, long, value_enum, default_value_t = Format::Text)]
    format: Format,
}

#[derive(Subcommand)]
enum Commands {
    Init(InitArgs),
    Build { #[arg(short, long)] release: bool },
}

#[derive(Args)]
struct InitArgs {
    name: String,
    #[arg(short, long, default_value = "default")]
    template: String,
}

#[derive(Clone, ValueEnum)]
enum Format { Text, Json, Table }
```

## tracing (Structured Logging)

```rust
use tracing::{info, warn, instrument};
use tracing_subscriber::EnvFilter;

fn setup_tracing(verbosity: u8) {
    let filter = match verbosity {
        0 => "warn", 1 => "info", 2 => "debug", _ => "trace",
    };
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::new(filter))
        .with_target(false)
        .init();
}

#[instrument(skip(content))]
fn process_file(path: &str, content: &[u8]) -> Result<()> {
    info!(path, size = content.len(), "processing file");
    Ok(())
}
```

## indicatif (Progress)

```rust
use indicatif::{ProgressBar, ProgressStyle};

let pb = ProgressBar::new(total);
pb.set_style(ProgressStyle::default_bar()
    .template("{spinner:.green} [{bar:40}] {pos}/{len} ({eta})")
    .unwrap());

for item in items {
    process(item)?;
    pb.inc(1);
}
pb.finish_with_message("done");
```

## Error Display (miette for rich diagnostics)

```rust
use miette::{Diagnostic, SourceSpan};
use thiserror::Error;

#[derive(Error, Diagnostic, Debug)]
#[error("invalid config")]
#[diagnostic(code(app::config), help("check docs"))]
pub struct ConfigError {
    #[source_code]
    src: String,
    #[label("invalid here")]
    span: SourceSpan,
}
```

## Key Crates

- `clap` 4.x — arg parsing (derive or builder)
- `tracing` + `tracing-subscriber` — structured logging
- `indicatif` 0.17.x — progress bars
- `dialoguer` 0.11.x — interactive prompts
- `console` / `owo-colors` — terminal colors
- `miette` 7.x / `color-eyre` 0.6.x — rich error display
- `assert_cmd` + `predicates` — CLI integration testing
