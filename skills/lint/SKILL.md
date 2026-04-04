---
name: lint
description: Run linters and formatters. Auto-detects stack (Go, React, Flutter). Use when user says "lint", "check code style", "format code", "run eslint", "run prettier", "golangci-lint", "dart analyze", or after writing code that should be validated. MANDATORY after any code modification — invoke before considering any code task done.
---

# Lint

## Mandatory Post-Code Gate

This skill MUST be invoked after any code modification (Write, Edit on source files) before a task is considered done. This is not optional.

1. Run the linter for the detected stack
2. Run the formatter for the detected stack
3. If new lint errors → fix immediately, do not leave for a separate task
4. Zero new violations is the bar — never increase the lint error count

This skill handles **code style and static analysis only**. For running tests, use `/run-tests`.

This applies to all agents (developer, tester) and direct edits. Never ship code without passing this gate.

## Auto-Detection

Detect stack by checking for marker files:

| File | Stack | Linter | Formatter |
|------|-------|--------|-----------|
| `go.mod` | Go | `golangci-lint run ./...` | `gofmt` (built-in) |
| `package.json` | React/Node | `npx eslint .` | `npx prettier --check .` |
| `pubspec.yaml` | Flutter | `dart analyze` | `dart format --set-exit-if-changed .` |

If multiple stacks detected, lint each separately.

## Execution

### Go

```bash
# If .golangci.yml or .golangci.yaml exists, it will be used automatically
golangci-lint run ./...
```

If `golangci-lint` is not installed:
```bash
go vet ./...
```

Auto-fix: `golangci-lint run --fix ./...` — then report what couldn't be fixed.

### React/Node

```bash
# Lint
npx eslint . --ext .ts,.tsx,.js,.jsx

# Format check
npx prettier --check "src/**/*.{ts,tsx,js,jsx}"

# Auto-fix
npx eslint . --ext .ts,.tsx,.js,.jsx --fix
npx prettier --write "src/**/*.{ts,tsx,js,jsx}"
```

Check `package.json` scripts for project-specific lint commands (e.g., `npm run lint`).

Configuration files: `.eslintrc.*` or `eslint.config.*`, `.prettierrc`. Recommended: `eslint-config-react-app` or `@typescript-eslint`.

### Flutter

```bash
# Analyze
dart analyze

# Format check
dart format --set-exit-if-changed .

# Auto-fix
dart fix --apply
dart format .
```

Configuration: `analysis_options.yaml`. Recommended: `flutter_lints` or `very_good_analysis`.

Common auto-fixable issues: unused imports, missing `const` constructors, prefer `final` for immutable variables.

## Workflow

1. **Detect stack** — check for marker files (`go.mod`, `package.json`, `pubspec.yaml`)
2. **Check for project-specific commands** — read `package.json` scripts or `Makefile` for custom lint commands
3. **Auto-fix first** — run the fix command before reporting
4. **Then check** — run the check command to find remaining issues
5. **If errors > 0** — report errors with file:line, do NOT proceed to next task until errors are resolved
6. **If only warnings** — report warnings, proceed unless user wants them fixed
7. **Report counts** — `Errors: 3 | Warnings: 7 | Fixed: 12`

## Output Format

```
Stack: Go
Linter: golangci-lint (config: .golangci.yml)

Auto-fixed: 5 issues
Remaining:
  Errors (2):
    - internal/user/service.go:45 — ineffectual assignment to err (ineffassign)
    - internal/order/handler.go:12 — error return value not checked (errcheck)
  Warnings (1):
    - internal/billing/calc.go:78 — function too complex, cyclomatic complexity 15 (cyclop)
```

If clean: `Lint passed. No issues found.`
