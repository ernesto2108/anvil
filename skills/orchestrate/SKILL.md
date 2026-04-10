---
name: orchestrate
description: Smart orchestration — triages task complexity and runs only the agents needed. Use when user says "orchestrate", "new feature", "full workflow", "run the pipeline", or for any non-trivial task. Also auto-triggered by hook on complex requests.
disable-model-invocation: true
---

# Orchestration Workflow

The system acts as **Orchestrator**. First triages, then runs only the agents the task needs.

---

## Step -1 — In-flight snapshot (before triage)

Run `git status --short`. If non-empty, capture the list as **"Archivos ya modificados en esta sesión"** and pass it inline in every Medium+ developer invocation. Without this, agents collide with pending work from earlier tasks (the `stringPtr` incident in DASH-FEAT-008 traces to this gap). Skip only when `git status` is empty.

---

## Step 0 — Triage (ALWAYS FIRST)

### User-specified complexity (`-c` / `--complexity`)

**Accepted values:** `trivial` (direct, no agents) | `medium` (developer → tester) | `complex` (pm → architect → developer → tester → qa) | `max` (full pipeline).

When `-c` is present: skip automatic triage, still apply modifiers, confirm pipeline before launching, do NOT second-guess the user's choice. When `-c` is absent: fall through to automatic triage below.

### Automatic triage table

| Signal | Level | Pipeline |
|--------|-------|----------|
| Typo, config change, 1-2 files, clear fix | **Trivial** | Direct — no agents |
| 3-5 files, known pattern, no design decisions | **Medium** | developer → tester |
| New feature, new endpoint, design decisions needed | **Complex** | pm → architect → developer → tester → qa? |
| Cross-cutting, UI+backend, multi-service | **Maximum** | scanner → pm → designer → architect → developer → tester → security → qa → reporter |
| Bug fix (clear repro) | **Medium** | developer → tester |
| Bug fix (unclear) | **Medium** | pm → developer → tester → qa? |
| DB migration | **Complex** | architect → dba → qa |
| Infra / CI | **Complex** | devops → security |
| Security audit | **Medium** | security |
| Docs only | **Trivial** | tech-writer |
| Architecture docs | **Medium** | → `/document-service` (dedicated skill) |
| Refactor | **Complex** | architect → developer → tester → qa |
| LinkedIn post / social content | **Medium** | mkt-content |
| Content campaign / series | **Complex** | pm → mkt-content |
| Unclear scope | — | pm first — always |

**Triage modifiers:** UI → designer; DB schema → dba; infra/CI → devops; auth/sensitive data → security; context.md stale → scanner; marketing → mkt-content; two stacks → `docs/parallel-dev-phase.md`. **Reporter is NOT a default modifier** — see reporter gating below.

### Scope-based routing (read from PRD)

| PRD Scope | Designer | Architect | design-to-code |
|---|---|---|---|
| `new` / `both` | yes | yes | yes (if design file exists) |
| `visual-improvement` | yes | skip | yes |
| `functional-improvement` | skip | yes | skip |

UI work + design file (.pen / Figma) → use `/design-to-code`, pass design.md. No file → orchestrator → developer.

**After triage:** tell user the level and agents. Proceed only after they confirm.

---

## Boundary rule (summary)

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

**Never skip:** developer (code to write), lint + run-tests (before ship), tester (testable code).

### QA gating rules

**Run QA when ANY of:** complexity Large/Maximum (≥8 pts) — touches auth/permissions/sessions/tokens — touches DB schema/migrations — touches payment/billing — touches public DTOs/API contracts — touches crypto/secrets/input-sanitization/SQL-construction/file-paths — touches concurrency with shared state — user explicitly requests QA — refactor of critical subsystem (store, event bus, middleware, pipeline, migration runner) — previous task in series had QA score < 8.

**Skip QA when ALL:** Medium (3-5 pts) + none of the critical paths above + user did not request + no prior QA warning.

**Quality floor when skipped:** Self-QA checklist (developer.md) + lint + run-tests + enriched tester handoff.

**Announce QA decision during triage.** User may override. **When in doubt: run QA.**

### Reporter gating rules (skip by default)

**Run reporter ONLY when:** cross-service run · incident/postmortem · release/tag · user explicitly asks · `/document-service` flow.

**Skip when:** single task + `.handoff/` complete + sprint-current.md Done row updated. In that case, the `## Post-completion` block below IS the report.

Rationale and full trigger list: `agents/reporter.md`. **Announce the decision during triage.**

---

## Gates (hard stops)

- **PM gate:** user must approve PRD before architect starts
- **Design execution gate:** after designer produces ui-spec.md → PAUSE. Tell user: "Las specs de diseño están listas. Ejecuta el diseño en Pencil/Figma. Cuando termines, dime 'ya acabé' para continuar." Resume ONLY when user confirms. Verification checklist: see `vault-setup.md`.
- **Architect gate:** veto → STOP, re-discuss with user
- **QA gate:** score < 7 → STOP, fix issues before continuing
- **Security gate:** CVE critical/high → STOP, fix before continuing
- **PM backlog gate:** after PM produces PRD → verify tasks exist in sprint-current.md. If not, invoke PM again: "Break the PRD into backlog tasks in sprint-current.md."
- **Cross-repo sync gate:** backend DTO/endpoint/auth changes → developer MUST list affected frontend files in completion notes. Orchestrator adds these as follow-up tasks.

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

All docs MUST be in Spanish. Titles, headers, Mermaid labels → Spanish. Code, YAML keys, file names, paths → English.

---

## Post-completion (MANDATORY)

After all agents finish:
1. `<docs>/02-backlog/sprint-current.md` — add/update: ID, title, status `done`, service, type, story points, links, sprint metrics
2. `<docs>/02-backlog/board.md` — move task to **Done** column
3. Task frontmatter: set `status: done`

---

## Agent scope limits

Each agent: MAX 1 document per invocation. Multiple docs → run agent twice. Never 3+ files in a single run.

---

## Context passing (summary)

- **Content in context from legitimate source** (user messages, prior subagent output, vault docs) → pass inline in the agent prompt
- **Content NOT in context, subagent needs it** → tell subagent the file path to read
- **Content needs synthesis across many files** → spawn scanner/Explore, do NOT read them yourself
- **Agent output feeds the next agent** → pass relevant output inline (you already have it)

Reading production source code just to relay it = anti-pattern #5. See `anti-patterns.md`.

---

## Sub-files — load on trigger

Absolute path: `/Users/ernestodiaz/projects/anvil/skills/orchestrate/<sub-file>.md`.

| Trigger | Load |
|---|---|
| Invoking developer for Medium+ task (Flow 0/A/B) | `plan-approval.md` |
| Re-invoking developer after QA/security findings | `qa-fix.md` |
| Session start — project not in registry OR context.md stale OR Design Execution GATE resume | `vault-setup.md` |
| Any Read/Write that might violate boundary | `anti-patterns.md` |

---

## Size budget (enforced)

- SKILL.md ≤ 200 líneas · sub-files ≤ 150 (vault-setup.md ≤ 120)
- New rule que no cabe → borrar una vieja o mover a `agents/<agent>.md`
- Measure: `wc -l skills/orchestrate/*.md`
