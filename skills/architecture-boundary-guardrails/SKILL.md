---
name: architecture-boundary-guardrails
description: Prevent architectural drift by enforcing bounded contexts and use-case-per-file structure. Use when creating new services, moving code across domain boundaries, or detecting cross-context imports, god-interfaces, or monolithic service files.
user-invocable: false
---

Prevent architectural drift by enforcing bounded contexts and use-case-per-file structure.

Use when:
- creating or moving code across domain boundaries
- implementing application/domain/ports in a clean architecture project
- adding new services, handlers, workers, or repositories

## Detection

Before applying rules, detect the project's bounded contexts:
1. Read the project structure to identify domain modules/folders
2. If `<vault>/01-project/context.md` exists, use it for context boundaries
3. Otherwise infer contexts from top-level domain directories (e.g., `internal/`, `src/domains/`, `packages/`)

## Core Rules

- do not mix bounded contexts in one module/folder
- each domain context owns its own entities, value objects, and ports
- application layer must be use-case oriented:
  - one use case per file
  - avoid large service files accumulating unrelated operations
- keep ports scoped per context; avoid shared god-interfaces
- if a new file touches two contexts, stop and split

## Pre-Implementation Checklist

1. Identify target bounded context(s) for the task
2. Confirm destination folder belongs to that context
3. If code spans contexts, define explicit ports/contracts between them
4. Ensure each use case has its own file
5. Flag architecture-impact changes before coding refactors

## Validation Checks

- no cross-context entity leakage without port contracts
- no monolithic application files handling multiple unrelated use cases
- imports remain directional (domain <- application <- infrastructure)

## Output

- If violation exists: report file, violation type, and split plan
- If compliant: report "architecture guardrails OK"
