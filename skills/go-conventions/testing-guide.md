# Go Testing Guide

> **Dispatcher:** Load ONLY the files relevant to your test scope. Each file is ~3-5KB.

## Routing Table

| Scope | File |
|---|---|
| Test file structure + table-driven tests | `guides/testing/structure-tables.md` |
| HTTP handler tests | `guides/testing/http-handlers.md` |
| Repository / DB tests | `guides/testing/repositories.md` |
| Manual mocks + test helpers | `guides/testing/helpers-mocking.md` |
| Fixtures + integration test setup | `guides/testing/fixtures-integration.md` |
| Coverage targets + benchmarks | `guides/testing/coverage-benchmarks.md` |

## Always load

- `guides/testing/structure-tables.md` — naming, package suffix, table-driven pattern (required for ALL Go tests)
- `guides/testing/helpers-mocking.md` — manual mocks, `t.Helper()`, no mockery/gomock

## Load when relevant

- `guides/testing/http-handlers.md` — if testing HTTP handlers or middleware
- `guides/testing/repositories.md` — if testing DB queries or repositories
- `guides/testing/fixtures-integration.md` — if writing integration tests or using testdata fixtures
- `guides/testing/coverage-benchmarks.md` — if coverage is part of the task

## Key rules (summary)

- Package: use `package foo_test` (black-box) unless testing unexported logic
- Naming: `Test_FunctionName` or `Test_Type_Method` (underscore after `Test`)
- Table-driven by default — any function with >1 scenario gets a `tests []struct{name string, ...}` loop
- Assertions: stdlib (`t.Fatalf`/`t.Errorf`) OR testify (`require`/`assert`) — match what the project already uses, never mix
- `require` (testify) / `t.Fatalf` (stdlib) for fatal checks; `assert` / `t.Errorf` for non-fatal
- No mockery, no gomock — all mocks are written manually as structs implementing the interface
- `t.Helper()` on every test helper function
- Error assertions check message/type, not just `err != nil`
