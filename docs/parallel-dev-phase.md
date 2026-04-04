# Parallel Dev Phase — Reference

When a task requires **two different stacks** (e.g., Go backend + React frontend, Go backend + Flutter mobile), The orchestrator MAY launch two developer agents in parallel. This is the ONLY agent that supports parallel execution.

## When to activate

All three conditions must be true:
1. The task requires **two different convention skills** (go + react, go + flutter)
2. The stacks write to **disjoint directories** (e.g., backend in `cmd/`, `internal/` — frontend in `src/`, `app/`)
3. `design.md` and `ui-spec.md` both exist (architect and designer already ran)

## How it works

```
Phase 1 — Parallel (two developer agents launched simultaneously):
  developer(backend-skill) → implements API, domain logic, persistence
  developer(frontend-skill) → maquetation from ui-spec.md using mock data

Phase 2 — Sequential (after both complete):
  developer(frontend-skill) → integration pass: replace mocks with real API calls

Phase 3 — Normal pipeline continues:
  tester → qa
```

## Execution rules

- Use `isolation: "worktree"` for the **frontend maquetation** agent (Phase 1) to avoid file conflicts
- Backend developer runs in the main worktree
- After Phase 1 completes, The orchestrator merges the frontend worktree branch before starting Phase 2
- Phase 2 developer receives: the backend's API contracts + the maquetation code as starting point
- If directories are NOT disjoint → fall back to sequential development

## Context passing for parallel phase

| Developer instance | Receives | Skill |
|---|---|---|
| Backend | prd.md, design.md, convention skill | go-conventions |
| Frontend (maquetation) | prd.md, design.md, ui-spec.md, convention skill, **mock data contracts from design.md** | react-conventions or flutter-conventions |
| Frontend (integration) | prd.md, design.md, backend API contracts (actual), maquetation code, convention skill | react-conventions or flutter-conventions |

## Triage question

Add to orchestrate Step 0: "Does this task need two different stacks?" → evaluate parallel dev phase.
