---
name: run-tests
description: Run project tests with race detector and coverage. Auto-detects stack (Go, React, Flutter). Use when user says "run tests", "test this", "check coverage", "run vitest", "go test", "flutter test", or after implementing code that needs verification.
---

# Run Tests

## Auto-Detection

Detect stack by checking for marker files in the project root:

| File | Stack | Command |
|------|-------|---------|
| `go.mod` | Go | `go test ./... -race -cover -count=1` |
| `package.json` | Node/React | `<pm> exec vitest run --coverage` or `<pm> test -- --coverage` (detect `<pm>` per CLAUDE.md — prefer `pnpm`) |
| `pubspec.yaml` | Flutter | `flutter test --coverage` |

If multiple stacks detected, run tests for each stack separately.

For Node/React: check `package.json` for test runner — prefer `vitest` if configured, fall back to `jest`, then `<pm> test`. Detect the package manager from the lockfile per CLAUDE.md (`pnpm-lock.yaml` → pnpm, `yarn.lock` → yarn, `package-lock.json` → npm) and use it consistently.

## Execution

### Go
```bash
go test ./... -race -cover -count=1
```

Flags:
- `-race`: detect race conditions (always on)
- `-cover`: show coverage per package
- `-count=1`: disable test caching (ensures fresh run)

For integration tests (if requested):
```bash
go test ./... -race -cover -count=1 -tags integration
```

For a specific package:
```bash
go test -race -cover -count=1 ./internal/user/...
```

### React/Node

Detect package manager from lockfile per CLAUDE.md. Examples below use `pnpm` (default); swap to `npm` / `yarn` as detected.

```bash
# Vitest (preferred)
pnpm exec vitest run --coverage        # npm: npx vitest run --coverage
                                       # yarn: yarn exec vitest run --coverage

# Jest fallback
pnpm exec jest --coverage --passWithNoTests

# Generic fallback (runs the "test" script from package.json)
pnpm test -- --coverage                # npm: npm test -- --coverage
                                       # yarn: yarn test --coverage

# Specific file
pnpm exec vitest run src/path/Component.test.tsx

# Watch mode (if requested)
pnpm exec vitest
```

Troubleshooting: act warnings — wrap state updates in `act()`. Async issues — use `waitFor()` instead of manual timeouts. Missing providers — wrap component in necessary context providers.

### Flutter
```bash
# Unit + widget tests
flutter test --coverage

# Specific file
flutter test test/path/to_test.dart

# Integration tests
flutter test integration_test/

# With expanded reporter
flutter test --reporter expanded
```

Troubleshooting: golden failures — `flutter test --update-goldens`. Flaky tests — check for missing `pumpAndSettle()` or unresolved futures. View coverage: `genhtml coverage/lcov.info -o coverage/html`.

## Analyze Results

### Categorize failures

When tests fail, categorize each failure:

| Category | Signal | Action |
|----------|--------|--------|
| **Compilation error** | `cannot find`, `undefined`, `syntax error` | Fix code first |
| **Assertion failure** | `expected X got Y`, `assert`, `require` | Check logic or update test |
| **Timeout** | `context deadline exceeded`, `test timed out` | Check for deadlocks, increase timeout |
| **Race condition** | `DATA RACE`, `concurrent map` | Add synchronization |
| **Flaky** | Passes on retry | Investigate timing dependencies |

### Coverage check

Report coverage percentage. Flag if below 80% on business logic packages:

```
Coverage: 73.2% — BELOW THRESHOLD (80%)
Low coverage:
  - internal/billing: 45.2%
  - internal/notification: 61.0%
```

### Re-run failures

Suggest command to re-run only failed tests:

```bash
# Go — run specific failing test
go test -race -run TestFailingName ./internal/pkg/...

# Vitest (swap `pnpm exec` for `npx` / `yarn exec` per lockfile)
pnpm exec vitest run --reporter=verbose path/to/failing.test.ts

# Flutter
flutter test test/failing_test.dart
```

## Output Format

Summarize results in a table:

```
| Metric    | Result |
|-----------|--------|
| Passed    | 142    |
| Failed    | 3      |
| Skipped   | 5      |
| Coverage  | 84.1%  |
| Duration  | 12.3s  |
| Races     | 0      |
```

## Workflow

1. **Detect stack** — check for marker files
2. **Check for project-specific commands** — read `package.json` scripts or `Makefile` for custom test commands
3. **Run tests** — execute the appropriate command for the detected stack
4. **Categorize failures** — use the failure table above
5. **If compilation errors** — stop, fix code first, then re-run
6. **If assertion failures** — report with context, ask user: "Should I fix the code or update the test?"
7. **If coverage < 80%** — flag low-coverage packages, ask user if coverage improvement is needed
8. **Report results** — output the summary table

If all pass: report table only.
If failures: report table + categorized failure details + re-run commands.
