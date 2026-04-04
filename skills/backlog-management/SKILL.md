---
name: backlog-management
description: Task creation, backlog management, and sprint board format. Defines how to break PRDs into tickets, assign agents, and track progress. Used by the PM agent after writing a PRD.
---

# Backlog Management

## When to use

After a PRD is written, the PM MUST break it into tasks before any agent starts working. No PRD without tasks. No tasks without a PRD reference.

## Task ID format

`<PROJECT>-<AREA>-<NNN>`

Areas: FEAT, SEC, BUG, TECH, INFRA, DOC, TEST

Check existing IDs in `<docs>/02-backlog/sprint-current.md` before assigning new ones.

## Breaking a PRD into tasks

Read the PRD's functional requirements and acceptance criteria. Create one task per:

1. **Each P0 requirement** → at least one task
2. **Each component that needs separate work** (backend, frontend, DB, infra)
3. **Tests** → separate task per component (developer writes code, tester writes tests)
4. **Migrations** → separate task if DB changes needed
5. **Documentation** → separate task if user-facing docs needed

### Decomposition rules

- One concern per task — if a task touches backend AND frontend, split it
- Tasks should be completable in 1-8 points (if > 8, break down further)
- Every task must reference its PRD: `PRD: <TASK-ID>`
- Every task must have an assigned agent type
- Tests are ALWAYS a separate task from implementation

### Example decomposition

PRD: `PROJ-FEAT-042` — Add password reset flow

| Task ID | Title | Agent | Points | Depends on |
|---|---|---|---|---|
| PROJ-FEAT-042-01 | Create password reset endpoint | developer | 5 | — |
| PROJ-FEAT-042-02 | Add email sending service | developer | 3 | — |
| PROJ-FEAT-042-03 | Create password reset UI | developer | 5 | 01 |
| PROJ-FEAT-042-04 | Add migration for reset tokens table | dba | 2 | — |
| PROJ-FEAT-042-05 | Tests for reset endpoint | tester | 3 | 01 |
| PROJ-FEAT-042-06 | Tests for email service | tester | 2 | 02 |
| PROJ-FEAT-042-07 | Tests for reset UI | tester | 3 | 03 |
| PROJ-FEAT-042-08 | Security review | security | 2 | 01, 02 |

## Task format

**CRITICAL:** Always read the existing `sprint-current.md` before adding tasks. Match the format that already exists — never impose a different format.

The standard format uses **tables**, not markdown headers:

### Backlog table row
```
| TASK-ID | Task description | P | Type | Agent | Pts | Repo |
```

### Section header row (for grouping related tasks)
```
| | **── Feature Name (PARENT-ID, date) ──** | | | | | |
```

### In Progress table row
```
| TASK-ID | Task | P | Agent | Start date | Branch |
```

### Done table row
```
| TASK-ID | Task | Type | Date | Notes |
```

## Sprint board format

`<docs>/02-backlog/sprint-current.md`:

```markdown
# Sprint Backlog

> Sprint #N | YYYY-MM-DD → ongoing | Goal: <sprint goal>

## Backlog
| ID | Tarea | P | Tipo | Agente | Pts | Repo |
|----|-------|---|------|--------|-----|------|
| | **── Feature Name (TASK-ID, date) ──** | | | | | |
| PROJ-FEAT-001 | Create password reset endpoint | P1 | feat | developer | 5 | my-service |
| PROJ-FEAT-002 | Password reset UI | P1 | feat | developer | 5 | my-web |
| PROJ-TEST-001 | Tests for reset endpoint | P1 | test | tester | 3 | my-service |

## TODO
| ID | Tarea | P | Tipo | Agente | Pts | Repo |
|----|-------|---|------|--------|-----|------|

## In Progress
| ID | Tarea | P | Agente | Inicio | Branch |
|----|-------|---|--------|--------|--------|

## Blocked
| ID | Tarea | P | Agente | Bloqueado por |
|----|-------|---|--------|---------------|

## In Review
| ID | Tarea | Agente | Reviewer | PR |
|----|-------|--------|----------|-----|

## Done
| ID | Tarea | Tipo | Fecha | Notas |
|----|-------|------|-------|-------|
```

## Task lifecycle

```
PM creates PRD
  → PM breaks into tasks (this skill)
  → Tasks go to Backlog column
  → Orchestrator picks task, assigns to agent
  → Agent starts → task moves to In Progress
  → Agent finishes → task moves to Done with date
  → All tasks done → PRD is complete
```

## Rules

- **No work without a ticket** — if an agent needs to do something, there must be a task for it
- **No ticket without a PRD** — every task references its parent PRD (except bugs with repro steps)
- **Dependencies must be explicit** — if task B needs task A done first, write "Depends on: A"
- **Acceptance criteria from PRD** — each task's criteria come from the PRD's Given/When/Then scenarios
- **Points are fibonacci** — 1, 2, 3, 5, 8, 13. If > 8, break it down
- **Status updates are mandatory** — agents must update task status when starting and finishing
