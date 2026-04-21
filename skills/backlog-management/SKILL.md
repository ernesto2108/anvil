---
name: backlog-management
description: Task creation, backlog management, and sprint board format. Defines how to break PRDs into tickets, assign agents, and track progress. Used by the PM agent after writing a PRD.
---

# Backlog Management

## When to use

After a PRD is written, the PM MUST break it into tasks before any agent starts working. No PRD without tasks. No tasks without a PRD reference.

## Work hierarchy

```
PROJECT (the repo/product — e.g., Anvil, Dashboard)
  └── MILESTONE (deliverable milestone — e.g., MVP, v1.0, v2.0)
       └── US (User Stories / Features — each PRD is a US)
            └── TASKS (technical tasks — backend, frontend, DB, tests)
                 └── SUB TASKS (steps within a task — tracked in handoff, not backlog)
```

- **PROJECT** — implicit from the repo. Not tracked in backlog
- **MILESTONE** — groups of related features with a shared delivery goal. Tracked as a field in PRD Scope and task frontmatter
- **US** — each PRD represents a User Story or Feature. The PRD is the US
- **TASKS** — the decomposition of a US into technical work items. These are the rows in sprint-current.md
- **SUB TASKS** — implementation steps within a task. Tracked in `.handoff/` files, not in the backlog

### Milestone management

Milestones are defined in the PRD's `## Scope` section (`Milestone` field) and propagated to every task.

**In sprint-current.md:** group tasks by milestone using section headers:
```
| | **── 🎯 MVP ──** | | | | | |
| PROJ-FEAT-001 | Create auth flow | P0 | feat | developer | 5 | my-service |
| PROJ-FEAT-002 | Auth UI | P0 | feat | developer | 5 | my-web |
| | **── 🎯 v1.0 ──** | | | | | |
| PROJ-FEAT-010 | Add analytics | P1 | feat | developer | 3 | my-service |
```

**In task frontmatter:** include `milestone: <name>` so Dataview can group/filter by milestone.

**Dashboard query:** `dashboard.md` can use `GROUP BY milestone` to show progress per milestone.

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

## Obsidian integration (MANDATORY)

The backlog lives in an Obsidian vault. Every task and sprint file must be compatible with Obsidian plugins: **Dataview** (queries) and **Kanban** (visual board).

**All templates live in `vault-template/` in the Anvil repo root.** Read them directly — never hardcode templates inline in skills or agents.

### Task file frontmatter (MANDATORY for every task.md)

Every `<docs>/03-tasks/<TASK-ID>/task.md` MUST include Dataview frontmatter. Read `vault-template/03-tasks/task-template.md` for the exact format and fields.

**Without this frontmatter, Dataview queries and the Kanban board will not work.** This is not optional.

### Sprint companion files (create alongside sprint-current.md)

When creating a new sprint, the PM MUST create 3 files based on the templates in `vault-template/02-backlog/`:

| File | Template source | Purpose |
|---|---|---|
| `<docs>/02-backlog/sprint-current.md` | `vault-template/02-backlog/sprint-current.md` | Sprint table with sections |
| `<docs>/02-backlog/board.md` | `vault-template/02-backlog/board.md` | Kanban board (Obsidian plugin) |
| `<docs>/02-backlog/dashboard.md` | `vault-template/02-backlog/dashboard.md` | Dataview dashboard queries |

**All three files must exist together.** Never create one without the others.

Each task in board.md is a checkbox item with a wiki-link and relevant tags:
```
- [ ] [[TASK-ID/task]] Titulo de la tarea #proyecto #tag
```

### Updating companion files

- When tasks move between states → update both `sprint-current.md` AND `board.md` (move the checkbox to the correct column)
- When tasks are added → add to `sprint-current.md` table AND `board.md` Backlog column
- Update the task's frontmatter `status` field to match
- The `dashboard.md` is self-updating via Dataview queries — no manual updates needed

### State transitions — the 3-places rule (CRITICAL)

Every time a task changes state, you MUST update **exactly 3 files**. Forgetting any one of them causes drift: `sprint-current.md` and `board.md` are manual views, but `dashboard.md` is built on Dataview queries that read from task frontmatters — if you only update the first two, the dashboard shows stale data and the next orchestrator session thinks the task is still pending.

**Checklist for EVERY state transition:**

1. **`<docs>/02-backlog/sprint-current.md`** — move the row to the correct section (Backlog / TODO / In Progress / Blocked / In Review / Done). Done rows include: `| ID | Title | Type | YYYY-MM-DD | Notes |`.

2. **`<docs>/02-backlog/board.md`** — move the `- [ ]` / `- [x]` checkbox line to the correct Kanban column. When moving to Done, change `- [ ]` to `- [x]`.

3. **`<docs>/03-tasks/<TASK-ID>/task.md`** frontmatter — update the `status` field. If moving to Done, ALSO add a `completed: YYYY-MM-DD` field. **This is the one that gets forgotten.**

### State → frontmatter value mapping

| Kanban column | Frontmatter `status` | Extra fields |
|---|---|---|
| Backlog | `backlog` | — |
| TODO | `todo` | — |
| In Progress | `in-progress` | `started: YYYY-MM-DD` |
| Blocked | `blocked` | `blocked_by: <TASK-ID>` |
| In Review | `in-review` | `pr: <URL>` |
| Done | `done` | `completed: YYYY-MM-DD` |

### Common mistake — Done with stale frontmatter

**Symptom:** user says "I see task X still in the backlog" even though `sprint-current.md` and `board.md` show it in Done.

**Cause:** `03-tasks/<ID>/task.md` still has `status: backlog`. The Dataview query in `dashboard.md` reads frontmatters, not the Kanban columns, so it shows the task as pending.

**Fix:** grep for stale statuses before closing a sprint:
```bash
# Any task file whose status does not match the sprint-current section it is in
grep -r "status: backlog" <docs>/03-tasks/ | grep -v "$(awk '/## Backlog/,/## TODO/' <docs>/02-backlog/sprint-current.md)"
```

**Prevention:** when moving to Done, do all 3 file edits in the same tool call batch. Never split across messages — that is where the third file gets forgotten.

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
