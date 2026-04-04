---
name: scan-project
disable-model-invocation: true
description: Scan repo structure and write context.md to the vault with product objective and technical snapshot. Use when starting a new session, user says "scan the project", "what stack is this", "analyze the repo", or when context.md is missing or stale.
---

# Scan Project

Discover the REAL repository structure and tooling. Do NOT assume any architecture or folder naming. Mirror the project exactly as it exists.

## Step 1: Product Objective

If product objective context is missing or outdated in ``<vault>/01-project/context.md``, ask these questions first:
1. "What is the project objective in 3-6 lines?"
2. "What non-negotiable rules must I always respect?"

## Step 2: Detect Stacks

Check for these marker files to determine which stacks are present:

| File | Stack | What to collect |
|------|-------|-----------------|
| `go.mod` | Go | Go version, modules, `*_test.go` locations, `.golangci.*` |
| `package.json` | Node/React | Node version, framework (Next/Vite/CRA), `*.test.tsx` locations, eslint config |
| `pubspec.yaml` | Flutter | Dart version, dependencies, `*_test.dart` locations, `analysis_options.yaml` |
| `Cargo.toml` | Rust | Edition, dependencies |
| `requirements.txt` / `pyproject.toml` | Python | Python version, framework |

Multiple stacks can coexist (e.g., Go backend + React frontend).

## Step 3: Collect Information

For ALL stacks:
1. Directory tree (depth 3)
2. CI / runtime hints — `Dockerfile`, `docker-compose.*`, `.github/workflows/*`
3. Config files — `Makefile`, `taskfile`, scripts
4. Ignore: `<vault>/`, `.git/`, `vendor/`, `node_modules/`, `dist/`, `build/`, `tmp/`, `.next/`

### Go-specific
- Read `go.mod` — version and dependencies
- Search `*_test.go` — list test locations
- Search `.golangci.yml` / `.golangci.yaml`
- Search `internal/`, `cmd/`, `pkg/` structure

### React/Node-specific
- Read `package.json` — scripts, dependencies, devDependencies
- Detect framework: Next.js (`next.config.*`), Vite (`vite.config.*`), CRA (`react-scripts`)
- Search `*.test.tsx`, `*.test.ts`, `*.spec.tsx`
- Search eslint config (`.eslintrc.*`, `eslint.config.*`)
- Search prettier config (`.prettierrc.*`)
- Check for `tsconfig.json`

### Flutter-specific
- Read `pubspec.yaml` — dependencies, dev_dependencies
- Search `*_test.dart`
- Check `analysis_options.yaml`
- Check for `l10n.yaml` (localization)
- Check `lib/`, `test/`, `integration_test/` structure

## Step 4: Write Output

Write ONLY: ``<vault>/01-project/context.md`` (overwrite if exists).
Never drop technical sections when adding product objective context; keep both.

Use the template from `output-template.md` — include only sections for detected stacks.

## Actions Checklist

- [ ] Read marker files (`go.mod`, `package.json`, `pubspec.yaml`)
- [ ] List directories (depth 3)
- [ ] Per detected stack: collect version, deps, test files, linter config
- [ ] Search CI files (`.github/workflows/*`, `Dockerfile`, `docker-compose.*`)
- [ ] Search build tools (`Makefile`, `taskfile.*`)
- [ ] Ask product objective questions if missing
- [ ] Write ``<vault>/01-project/context.md``
