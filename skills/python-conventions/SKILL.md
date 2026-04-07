---
name: python-conventions
description: Python conventions and coding standards for 3.12+. Use when writing Python code, reviewing Python patterns, or user mentions "python conventions", "type hints", "pydantic", "pytest", "embeddings", "numpy", "async python", or working with .py files.
---

# Python Conventions

> **IMPORTANT:** This file is a lightweight dispatcher. Do NOT load all referenced files at once. Read the routing table below, identify which files are relevant to the current task, and load ONLY those using the Read tool. Each file is ~3-5KB. Loading unnecessary files wastes context tokens.

## Stack & Philosophy

- **Python 3.12+ features first** — type aliases, `X | Y` unions, `match`, `itertools.batched`
- **Type hints everywhere** — Pydantic v2, Protocol, TypedDict, `NDArray[np.float32]`
- **Ruff replaces everything** — one tool for lint + format (no flake8, black, isort)
- **uv over pip** — faster, lockfile built-in, single tool
- **Structured over stringly** — structlog, Pydantic models, not f-strings and dicts
- **Vectorize, don't loop** — numpy/pandas operations, preallocate, batch everything

## Red Flags (always stop work)

- `from typing import Union, Optional` on Python 3.10+ → use `X | Y`, `X | None`
- Mutable default arguments (`def f(x=[])`) → error
- `eval()` or `exec()` with user input → error
- f-strings in SQL queries → SQL injection
- `import *` → error
- Bare `except:` or `except Exception:` without re-raise → error
- `os.system()` or `subprocess.call(shell=True)` with user input → error

## Anti-Pattern Detection

**Passive detection:** When reviewing Python code, load `detection/anti-patterns.md` and scan for `error` and `warning` patterns. Report as `[file:line] [severity] [category] anti-pattern-name`.

**Active detection:** When user asks to "improve", "refactor", "optimize", or "clean" — also report `suggestion` level patterns and propose fixes.

## What to Load

Load **only** the files relevant to the current task:

### Rules (quick reference, ~2-3KB each)

| Working on... | Load |
|---|---|
| Type hints, naming, error handling, modern syntax | `rules/coding.md` |
| Project structure, imports, DI, Pydantic models | `rules/architecture.md` |
| NumPy, Pandas, vectorization, memory management | `rules/data.md` |

### Guides (detailed patterns with code, ~3-5KB each)

| Working on... | Load |
|---|---|
| asyncio, TaskGroup, structured concurrency | `guides/async/patterns.md` |
| Async context managers, timeouts, streaming | `guides/async/resources.md` |
| pytest fixtures, parametrize, mocking | `guides/testing/pytest.md` |
| Async testing, factory fixtures | `guides/testing/async-testing.md` |
| Embeddings, batch processing, vector ops | `guides/ml/embeddings.md` |
| Memory-mapped arrays, large datasets, GPU memory | `guides/ml/memory.md` |
| Context managers, connection pools, cleanup | `guides/cleanup/resources.md` |
| Input validation, SQL injection, secrets | `guides/security.md` |

### Detection & Checklists

| When... | Load |
|---|---|
| Code review | `detection/anti-patterns.md` |
| Before writing Python code | `checklists/pre.md` |
| After writing Python code | `checklists/post.md` |

### Examples (good + bad patterns by domain, ~2-3KB each)

| Working on... | Load |
|---|---|
| Type hints, modern syntax, Pydantic v2 | `examples/types.md` |
| Error handling, exception groups, logging | `examples/errors.md` |
| Testing patterns, fixtures, mocking | `examples/testing.md` |
| NumPy/Pandas, vectorization, batch processing | `examples/data.md` |
| Async patterns, TaskGroup, timeouts | `examples/async.md` |

## Post-Implementation Gate

After ANY code change to `.py` files, run `ruff check` and `ruff format --check` before considering the task done.
