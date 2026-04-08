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
4. **Present the plan to the user for approval** (see Approval Gate below)

### Approval Gate (MANDATORY before coding)

After creating the handoff with the execution plan, the developer MUST pause and present a summary to the user in Spanish:

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

¿Apruebas este plan o quieres ajustar algo?
```

**Rules:**
- Do NOT write any production code until the user approves
- The user may say: "dale", "ok", "aprobado" → proceed
- The user may say: "cambia X", "no me gusta Y", "mejor usa Z" → update the handoff plan, present again
- The user may say: "no, mejor hacemos otra cosa" → discard plan, start over
- On **continuations** (handoff already exists with approved plan) → skip the gate, resume from "Siguiente paso"

### Update (after each milestone)

Update incrementally — don't rewrite the whole file, append or update sections:
- Check off completed steps in "Estado actual"
- Add new entries to "Archivos modificados"
- Record decisions in "Decisiones tomadas"
- Update "Siguiente paso" to reflect current state

### Read (on session continuation)

The orchestrator reads the handoff and passes it inline to the developer. The developer resumes from "Siguiente paso" — does NOT re-read PRD/design/context unless the handoff explicitly says to.

### Archive (on task completion)

Called by `/task-complete`:
1. Read handoff content
2. Append `## Resumen de completacion` to the task file with: archivos modificados, decisiones tomadas, token usage summary
3. Delete the handoff file
4. If `.handoff/` is now empty, delete the directory

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

## Rules

- One file per task
- `.handoff/` MUST be in the project's `.gitignore`
- Handoff files are temporary — they do not belong in version control or documentation
- Do NOT create handoff for Small tasks
