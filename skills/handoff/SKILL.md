---
name: handoff
disable-model-invocation: true
description: Session continuity for Medium+ tasks. Creates, updates, reads, and archives handoff notes so developers can resume work across sessions without re-reading everything. Invoked by developer and orchestrator — not directly by the user.
---

# Handoff Notes

Handoff files live at `.handoff/` in the project root. They enable session continuity — if a developer runs out of tokens, the next session reads the handoff and picks up exactly where it left off.

## When handoff applies

| Complexity | Handoff | Why |
|---|---|---|
| **Small (1-5 pts)** | NO | Fits in one session, not worth the overhead |
| **Medium (5-8 pts)** | YES | May span sessions, context loss is expensive |
| **Large (8-13 pts)** | YES | Will almost certainly span sessions |

Applies whether the task comes from the backlog (TASK-ID) or is invoked directly without one.

## File naming

- **With TASK-ID:** `.handoff/<TASK-ID>.md`
- **Without TASK-ID:** `.handoff/<short-slug>.md` — derive a slug from the task description (e.g., `add-auth-middleware.md`, `fix-payment-flow.md`). The orchestrator or user may provide a name; if not, generate one.

## Operations

### Create (on task start)

1. Create `.handoff/` directory if it doesn't exist
2. Read `template.md` from this skill and use it to write the handoff file
3. Fill in the execution plan and empty Token usage table
4. **Return control to the orchestrator with the plan** — do NOT present the plan directly to the user and do NOT auto-continue. The orchestrator is responsible for the user approval gate (see Approval Gate below).

### Approval Gate (MANDATORY — user must approve manually)

The plan presented in the handoff MUST be approved by the **user**, not by the orchestrator. The orchestrator is a relay, not an authority on approvals.

**Flow:**
1. Developer finishes creating the handoff with the execution plan and returns control to the orchestrator with a plan summary
2. **Orchestrator surfaces the plan to the user in Spanish** and waits for an explicit response. The orchestrator uses `AskUserQuestion` or a direct text prompt — never assumes approval from silence
3. Developer does NOT continue coding until the orchestrator resumes it with explicit user approval (flag `plan_preapproved=true` or "plan approved — proceed")

**Orchestrator presentation format (Spanish):**

```
## Plan de ejecución — <TASK-ID or slug>

**Pasos:**
1. <step description>
2. <step description>
3. <step description>

**Archivos que voy a tocar:**
- `path/to/file` — qué cambio y por qué
- `path/to/file` — qué cambio y por qué

**Enfoque técnico:**
<brief explanation of the approach>

¿Apruebas este plan, quieres ajustar algo, o prefieres otra cosa?
```

**Rules:**
- The orchestrator MUST NOT auto-approve. Phrases like "el plan coincide con las specs, apruebo y continúo" are forbidden — the user decides
- The orchestrator waits for an explicit user response before resuming the developer
- User may say: "dale", "ok", "aprobado", "sí" → orchestrator resumes developer with `plan_preapproved=true`
- User may say: "cambia X", "no me gusta Y", "mejor usa Z" → orchestrator sends feedback back to developer, developer updates the handoff plan, loop
- User may say: "no, mejor hacemos otra cosa" → orchestrator discards plan, restart with new scope
- On **continuations** (handoff already exists with approved plan) → skip the gate, resume from "Siguiente paso"

### Pre-approved plans (orchestrator shortcut)

When the orchestrator has already designed the plan in detail in the main conversation AND the user has already approved it in the main conversation (before any agent was invoked), the orchestrator may pass `plan_preapproved=true` directly in the first developer invocation. In this case:

1. Developer creates the handoff file as a progress artifact (NOT as a blocking gate)
2. Developer proceeds to implementation immediately — no second invocation, no separate approval step
3. This avoids the overhead of two developer invocations (plan + impl) when the orchestrator has already done the planning work

**Orchestrator responsibility:** be honest about pre-approval. If there is ANY doubt that the user has explicitly approved the plan (not just said "continue" or "sigue" without seeing the details), use the normal flow, not the shortcut. When in doubt, ask.

### Update (after each milestone)

Update incrementally — don't rewrite the whole file, append or update sections:
- Check off completed steps in "Estado actual"
- Add new entries to "Archivos modificados"
- Record decisions in "Decisiones tomadas"
- Update "Siguiente paso" to reflect current state

### Read (on session continuation)

The orchestrator reads the handoff and passes it inline to the developer. The developer resumes from "Siguiente paso" — does NOT re-read PRD/design/context unless the handoff explicitly says to.

### Archive (on task completion)

Called by `/task-complete` or manually by the orchestrator:

1. Read handoff content
2. **Append `## Resumen de completacion` to the task file in the Obsidian vault** (e.g., `anvil-knowledge-base/03-tasks/<TASK-ID>/task.md`). This is the permanent record — include: fecha, archivos creados, archivos modificados, decisiones tomadas, resultado de tests. Follow the same format as existing completed tasks in the vault.
3. Move the handoff file to `.handoff/archive/<TASK-ID>.md`
4. Update the board in the vault (`02-backlog/board.md`) — move task from current column to Done
5. Update the task frontmatter: `status: done`, `completed: <date>`

## Template

See `template.md` in this skill directory for the handoff file structure.

## Token usage tracking

At the end of every session (complete or not), append a row to the **Token usage** table:

- **Session**: sequential number
- **Tokens used**: approximate tokens consumed (from agent metadata if available, otherwise estimate)
- **Tokens available**: token budget for the model used
- **Tool calls**: number of tool invocations
- **Files read**: number of files read
- **Files written**: number of files created or modified

## Cross-stack tasks

When a task touches multiple stacks (e.g., Go backend + React frontend):

1. Use `## Fases` instead of `## Estado actual` — one phase per stack, ordered by dependency (backend first)
2. Fill `## Puente de contratos` — the exact struct/DTO/interface that connects both sides, with JSON tags and TypeScript types side by side
3. Group `### Tests requeridos — por stack` by stack — each group with its own file path, run command, and numbered test list

**Why this matters:** cross-stack bugs almost always happen at the contract boundary. If the Go struct has `json:"runId"` but the TS interface expects `run_id`, a flat handoff won't catch it. The contract bridge makes both sides visible in one place.

For single-stack tasks, use the flat `## Estado actual` checklist and omit `## Fases`, `## Puente de contratos`, and the stack grouping in tests.

## Cross-service tasks

When a task touches multiple repos/services:

1. Fill `## Dependencias cross-service` — table with service, repo, what changes, and deploy order
2. Document shared contracts (API endpoints, event schemas, DB tables that cross boundaries)
3. Flag breaking changes and migration plan

The orchestrator MUST verify deploy order before closing the task.

## Input/Output tracking

### Input recibido (on task start)

The developer fills `## Input recibido` when creating the handoff. This is a receipt of what the orchestrator provided — if the next session finds a gap, it knows what was missing vs. what was lost.

### Output entregado (before finishing)

The developer fills `## Output entregado` before reporting done. This is the delivery checklist the orchestrator verifies. Must include: build result, lint result, existing tests result, file counts, contract bridge verification (if cross-stack), and cross-service impact.

## Retro (on task completion)

After the task is done (all agents finished, QA passed or skipped), fill `## Retro` before archiving. This is NOT optional for Medium+ tasks.

**What to record:**
- **Qué funcionó** — patterns, decisions, approaches worth repeating
- **Qué no funcionó** — rework, QA bounces, wrong assumptions, wasted reads. Be specific: "assumed nullable column was NOT NULL, caused QA bounce" not "should have checked"
- **Métricas** — estimated vs actual: story points, QA bounces, developer invocations, tester invocations
- **Aprendizaje** — one concrete takeaway for future tasks (not generic)

**Who fills it:**
- Developer fills "qué funcionó" and "qué no funcionó" from their perspective
- Orchestrator fills "métricas" (has the full picture of agent invocations and bounces)
- Either fills "aprendizaje"

**How it feeds improvement:**
- The orchestrator reads retros from `.handoff/archive/` when planning similar tasks
- Patterns that repeat across 3+ retros should become memory entries or convention skill updates

## Rules

- One file per task
- `.handoff/` MUST be in the project's `.gitignore`
- Handoff files are temporary — they do not belong in version control or documentation
- Do NOT create handoff for Small tasks
