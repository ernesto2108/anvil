---
name: orchestrate
description: Smart orchestration — triages task complexity and runs only the agents needed. Use when user says "orchestrate", "new feature", "full workflow", "run the pipeline", or for any non-trivial task. Also auto-triggered by hook on complex requests.
disable-model-invocation: true
---

# Orchestration Workflow

The system acts as **Orchestrator**. This skill is activated ONLY when the user explicitly requests it.

---

## Rule #0 — User activates the orchestrator

This skill runs ONLY when the user says "orquesta", "pipeline", "usa agentes", or passes `-c`. The orchestrator NEVER self-activates. If the user said "hazlo directo", this skill does not apply — work directly without agents.

## Rule #1 — Human gates between agents (MANDATORY)

After each agent completes:
1. Show the user a concise summary of what the agent produced
2. Ask (in Spanish): "¿Paso al siguiente agente (X)?" or "¿Quieres revisar primero?"
3. Continue ONLY with explicit user confirmation ("sí", "dale", "siguiente")

The user may at any point:
- Skip an agent: "salta el tester"
- Stop the pipeline: "hasta aquí"
- Switch to direct mode: "el resto hazlo directo"

---

## Step -1 — In-flight snapshot (before triage)

Run `git status --short`. If non-empty, capture the list as **"Archivos ya modificados en esta sesión"** and pass it inline in every developer invocation. Skip only when `git status` is empty.

---

## Step 0 — Triage

### User-specified complexity (`-c` / `--complexity`)

**Accepted values:** `medium` (developer → tester) | `complex` (pm → architect → developer → tester → qa) | `max` (full pipeline).

When `-c` is present: use that classification, apply modifiers, confirm pipeline before launching.

### When user does NOT specify `-c`

Present a recommended pipeline to the user using this reference table. **Do NOT auto-execute — always ask for confirmation.**

| Signal | Recommended pipeline |
|--------|---------------------|
| 3-5 files, known pattern, no design decisions | developer → tester |
| New feature, new endpoint, design decisions needed | pm → architect → developer → tester → qa? |
| Cross-cutting, UI+backend, multi-service | scanner → pm → designer → architect → developer → tester → security → qa |
| Bug fix (clear repro) | developer → tester |
| Bug fix (unclear) | pm → developer → tester → qa? |
| DB migration | architect → dba → qa |
| Infra / CI | devops → security |
| Security audit | security |
| Docs only | tech-writer |
| Refactor | architect → developer → tester → qa |
| LinkedIn post / social content | mkt-content |
| Unclear scope | pm first — always |

**Modifiers:** UI → designer; DB schema → dba; infra/CI → devops; auth/sensitive data → security; context.md stale → scanner; marketing → mkt-content.

Ask the user (in Spanish): "Recomiendo este pipeline: [lista]. ¿Apruebas o quieres ajustar?"

### Scope-based routing (read from PRD)

| PRD Scope | Designer | Architect | design-to-code |
|---|---|---|---|
| `new` / `both` | yes | yes | yes (if design file exists) |
| `visual-improvement` | yes | skip | yes |
| `functional-improvement` | skip | yes | skip |

UI work + design file (.pen / Figma) → use `/design-to-code`, pass architecture-frontend.md + ui-spec.md. No file → developer.

---

## Boundary rule (orchestration mode only)

These boundaries apply when running agents. In direct mode (user's choice), the orchestrator reads and writes code freely.

| | MAY | MUST NOT |
|---|---|---|
| Read | Vault docs, task docs, handoffs, project-registry.md | Source: `.go .ts .tsx .jsx .vue .svelte .py .rs .dart .astro .kt .swift .sql .css .scss` · Design files: `.pen .fig .sketch` |
| Write | Vault docs (copy-paste), sprint-current.md | Technical plans, task.md w/ synthesis, code files |

`Glob` and `Bash` (build/test) always allowed. *Full detail + examples: `anti-patterns.md`.*

**Design file reads delegated:** the orchestrator NEVER calls `mcp__pencil__*`, Figma MCP, or similar. `.pen` / Figma / Sketch reads belong to the downstream agent that consumes them (designer → architect → developer). If the orchestrator needs a fact from a design file, it asks the first pipeline agent that can legitimately read it — never peeks itself.

---

## Agent skip rules

| Agent | Skip when |
|-------|-----------|
| scanner | context.md exists and was updated this session |
| pm | requirements are already clear and specific (bug with repro steps, user gave exact spec) |
| designer | no UI changes |
| architect | no design decisions (pattern already exists, just extending) |
| dba | no DB changes |
| devops | no infra changes |
| security | no auth, no sensitive data, no external APIs |
| qa | see QA gating rules below — NOT always run |
| reporter | **SKIP BY DEFAULT** — see reporter gating below |
| tester | no testable code (docs, config, infra) |
| mkt-content | no marketing content needed |

**Never skip without asking:** developer (code to write), tester (testable code). The user may override ("salta el tester").
**Always run (no user override):** lint + run-tests before ship.

### QA gating rules

**Run QA when ANY of:** complexity Large/Maximum (≥8 pts) — touches auth/permissions/sessions/tokens — touches DB schema/migrations — touches payment/billing — touches public DTOs/API contracts — touches crypto/secrets/input-sanitization/SQL-construction/file-paths — touches concurrency with shared state — user explicitly requests QA — refactor of critical subsystem (store, event bus, middleware, pipeline, migration runner) — previous task in series had QA score < 8.

**Skip QA when ALL:** Medium (3-5 pts) + none of the critical paths above + user did not request + no prior QA warning.

**Quality floor when skipped:** Self-QA checklist (developer.md) + lint + run-tests + enriched tester handoff.

**Include QA recommendation in triage proposal.** User decides. **When in doubt: recommend QA.**

### Reporter gating

**Run ONLY when:** cross-service run · incident/postmortem · release/tag · user explicitly asks. **Skip by default** for single tasks.

---

## Gates (hard stops — in addition to Rule #1 human gates)

- **PM gate:** user approves PRD before architect starts
- **Design execution gate:** after designer produces ui-spec.md → PAUSE for user to execute in Pencil/Figma
- **Architect gate:** veto → STOP, re-discuss with user
- **QA gate:** score < 7 → STOP, fix before continuing
- **Security gate:** CVE critical/high → STOP, fix before continuing
- **PM backlog gate:** after PRD → verify tasks in sprint-current.md
- **Cross-repo sync gate:** backend DTO/endpoint/auth changes → developer lists affected frontend files

---

## Post-developer verification (MANDATORY for Medium+ tasks)

Before launching tester, verify ALL. Any fail → re-invoke developer, do NOT proceed.

1. `.handoff/<TASK-ID>.md` exists
2. Claimed files exist (`Glob`)
3. Build passes (`go build -tags <tag>` / `npm run build`)
4. **Lint HARD GATE** — handoff lists `golangci-lint` / `npm run lint` / `ruff` / `cargo clippy` with **0 issues** on the touched scope. If missing, bounce back — developer owns the lint run, orchestrator does NOT silently accept
5. `## Handoff for tester` section complete (signatures + edge cases + validation log — see `plan-approval.md`)

---

## Orchestration rules

- Resolve `<docs>` from `~/.claude/project-registry.md` before any agent; pass docs path + TASK-ID to every agent
- Specify convention skill for Developer; specify stack for Tester
- **One writer at a time.** Max tasks per run: 2 (preferred: 1). Scope change → re-run PM.

---

## Language

All docs in Spanish. Titles, headers, Mermaid labels → Spanish. Code, YAML keys, file names, paths → English.
Questions to user → always in Spanish.

---

## Post-completion (MANDATORY)

After all agents finish: update `sprint-current.md` (Done row), `board.md` (Done column), task frontmatter (`status: done`).

---

## Context passing

- Content already in context → pass inline to agent. Content NOT in context → tell agent the file path.
- Agent output feeds next agent → pass inline. Never read source code just to relay it (anti-pattern #5).
- Each agent: MAX 1 document per invocation. Multiple docs → run agent twice.

### Architecture views → Developer routing

The architect produces `architecture.md` + domain-specific views. Pass ONLY the relevant views per developer stack:

| Developer stack | Pass these architecture files |
|---|---|
| Go (backend) | `architecture.md` + `architecture-backend.md` + `architecture-db.md` (if exists) |
| React / frontend | `architecture.md` + `architecture-frontend.md` |
| Flutter / mobile | `architecture.md` + `architecture-frontend.md` (mobile section) |
| DBA | `architecture.md` + `architecture-db.md` |
| DevOps | `architecture.md` + `architecture-infra.md` |
| Full-stack (single dev) | `architecture.md` + all generated views |

The `architecture.md` (overview) ALWAYS goes to every developer — it has the decisions and cross-cutting context. Domain views are additive.

---

## Sub-files — load on trigger

Path: `/Users/ernestodiaz/projects/anvil/skills/orchestrate/<sub-file>.md`.

| Trigger | Load |
|---|---|
| Invoking developer (Flow 0/A/B) | `plan-approval.md` |
| Re-invoking developer after QA/security | `qa-fix.md` |
| Session start — registry/context stale | `vault-setup.md` |
| Boundary doubt | `anti-patterns.md` |

---

## Size budget

SKILL.md ≤ 200 lines · sub-files ≤ 150. `wc -l skills/orchestrate/*.md`
