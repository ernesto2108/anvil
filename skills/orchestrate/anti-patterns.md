---
name: orchestrate/anti-patterns
description: The 7 orchestrator anti-patterns with full reasoning, examples, and the Token discipline section. Load when about to do any non-trivial Read/Write operation or when boundary doubt arises.
---

# Anti-Patterns

**Load when:** about to do any non-trivial Read/Write op, or when boundary doubt arises.

---

## 1. Manual handoff as triage bypass

**Wrong:** user asks for Medium+ feature, orchestrator writes the plan in `.handoff/<TASK>.md` directly and executes it themselves (rationalizes: "I already have context, agents are overhead").

**Why it is wrong:** The `/orchestrate` skill IS the triage. Bypassing it skips pipeline selection, skill loading, and agent boundaries — violating the "developer is the only code writer" rule.

**Right:** Invoke triage, select pipeline, spawn developer with the plan inline. Handoff is written BY the developer, not instead of it.

---

## 2. "This looks simple, I'll just do it"

**Wrong:** orchestrator decides a task is "basically boilerplate" or "just config files" and skips straight to editing.

**Why it is wrong:** "Boilerplate" ≠ trivial. A scaffold with 12 config files, new build tags, new Makefile targets is Medium+ — touches many files, affects every future build. Small per-file changes ≠ small task.

**Right:** use the triage table. If in doubt, round UP.

---

## 3. Writing code "just this once"

**Wrong:** orchestrator writes a small `.go`, `.ts`, `.tsx`, `.py`, `.rs`, `.dart` file directly because spawning the developer feels disproportionate.

**Why it is wrong:** "Just this once" happens every time. Each direct edit bypasses the convention skills (react-conventions, go-conventions, etc.) that the developer agent loads. The resulting code drifts from project standards and the user notices.

**Right:** if the change touches any application code file, delegate to the developer agent with the correct convention skill named explicitly in the prompt. If the overhead feels high, batch multiple changes into ONE developer call — do not bypass the boundary.

**Exceptions (direct edit OK):**
- `go.mod` via `go get` / `go mod tidy` (tooling, not hand-written code)
- Config files: `Makefile`, `wails.json`, `.gitignore`, `tsconfig.json`, `package.json` (but validate versions carefully)
- Shell commands for verification: `go build`, `go vet`, `npm run build`
- Markdown docs, handoff files, sprint-current.md

If the extension is `.go`, `.ts`, `.tsx`, `.jsx`, `.vue`, `.svelte`, `.py`, `.rs`, `.dart`, `.astro`, `.kt`, `.swift` — **always** delegate to developer.

---

## 4. Claiming a skill was loaded without confirmation

**Wrong:** orchestrator names a skill in the prompt and assumes the agent loaded it without verifying.

**Why it is wrong:** the agent may have silently skipped it. The user catches this later — forcing a re-run.

**Right:** after agent completes, verify its report names the skill or references its rules. If neither, follow up: "confirm you loaded <skill> and re-validate files against its rules." Cheaper than a redo.

---

## 5. Building context in Opus instead of delegating to a subagent

**Wrong:** orchestrator reads 10+ production source files (Go, React, DTOs, store internals, migrations, existing views) to "understand the feature" before invoking any agent. The content then gets pasted inline into the developer prompt — paying Opus rates for a read the developer would have done in Sonnet anyway.

**Why it happens:** feels like due diligence. The orchestrator rationalizes "I'm only gathering context, not writing code, so the code-writer boundary doesn't apply." It also feels faster than spawning a scanner.

**Why it is wrong:**
1. **Cost ratio.** Opus is roughly 4x Sonnet per token. Every file the orchestrator reads costs 4x the same read inside any subagent. 15 files × average 300 lines × 4x = tens of thousands of wasted Opus tokens per task.
2. **Double-counting.** Any content the orchestrator reads AND then pastes inline into a subagent prompt is paid for twice. The context-passing rule says "pass inline only when you already have it" — reading files SO THAT you can pass them inline is the failure mode, not the intent.
3. **Subagents read better.** `scanner`, `Explore`, and domain agents read with the right conventions loaded and the right framing. The orchestrator reads raw.

**Right:**
- For initial codebase understanding: spawn `scanner` (if `context.md` is stale) OR the `Explore` subagent with a precise question — "list existing stubs, DTOs, and affected files for <feature>; summarize the pattern used by <sibling feature>". Consume the summary, not the files.
- For feature-specific details the developer will need: let the developer subagent read them during its planning pass (Flow B) — that is exactly what Flow B is for.
- If the orchestrator already has content from PRIOR conversation turns (user pasted code, previous agent result, docs read for routing), that is legitimate inline-pass material. Freshly reading code in order to relay it is not.

**Hard rule — the Opus read allow-list:** the orchestrator MAY read these paths with the `Read` tool:
- `<docs>/01-project/context.md`
- `<docs>/02-backlog/sprint-current.md`, `board.md`, `dashboard.md`
- `<docs>/03-tasks/<TASK-ID>/*.md` (task.md, prd.md, design.md, ui-spec.md)
- `<docs>/04-architecture/*.md`
- `.handoff/<TASK-ID>.md` (own and sibling handoffs for pattern reference)
- `~/.claude/project-registry.md`
- Other markdown docs in the vault

The orchestrator MUST NOT `Read` source files with these extensions: `.go`, `.ts`, `.tsx`, `.jsx`, `.vue`, `.svelte`, `.py`, `.rs`, `.dart`, `.astro`, `.kt`, `.swift`, `.sql`, `.css`, `.scss`. Even to "just check the signature of one function." Delegate to a subagent with the exact question. `Glob` to verify file existence and `Bash` for `go build`/`npm run build` verification are allowed — they do not consume file contents.

---

## 6. Writing the technical plan in Opus instead of using Flow B

**Wrong:** orchestrator drafts a 200-400 line developer prompt containing exact Go function signatures, SQL queries, React component structure, file-by-file execution order, and marks it `plan_preapproved=true`. The "plan" is effectively architect + developer design work, performed in Opus.

**Why it happens:** the orchestrator thinks a complete spec saves the developer from rediscovery. It does — at 4x the cost per token. And the orchestrator is NOT faster at designing than the developer agent with fresh code context loaded through its convention skills.

**Why it is wrong:** Flow A (`plan_preapproved=true`) exists for the case where the USER typed the file-level plan in chat and the orchestrator is just relaying it. It is NOT a license for the orchestrator to write the plan itself and then "pre-approve" it on the user's behalf. If the orchestrator wrote the plan, that is Flow B territory — the developer should have written it.

**Right:** default to Flow B for anything bigger than Trivial. Invoke developer with "plan and STOP". Developer returns a ~50 line plan summary (Sonnet). Orchestrator surfaces it to the user verbatim. After explicit user approval, re-invoke developer with `plan_preapproved=true` and the now-approved short plan inline. Two Sonnet invocations beat one Opus-authored mega-prompt.

**Flow A legitimacy test:** before claiming a plan is "pre-approved in the main conversation", check — did the USER type the file list and approach in chat? Or did the ORCHESTRATOR synthesize it and the user only approved the high-level strategy (e.g., "go with option B")? If the latter, that is Flow B, not Flow A. "Option B" = strategic approval, not file-level approval. The developer still needs to produce the file-level plan. See `plan-approval.md` for the full Flow A / Flow B decision tree.

---

## 7. Writing docs the PM, developer, or reporter should write

**Wrong:** orchestrator writes `03-tasks/<NEW-ID>/task.md` frontmatter + context + AC for follow-up tasks discovered mid-execution. Orchestrator also writes "Resultado de ejecución" sections on completed task files, and enriched `sprint-current.md` Done rows synthesized from agent outputs.

**Why it is wrong:** task creation with Obsidian Dataview frontmatter + backlog formatting is PM work — the PM agent has `backlog-management` and `prd-template` skills loaded and produces the correct format first time. Final task reports and sprint enrichment are `reporter` work (or can be inlined into the developer handoff). All Sonnet. The orchestrator is producing Opus-priced docs that a Sonnet agent produces better.

**Right:**
- **Follow-up tasks discovered mid-run:** spawn `pm` with "create follow-up <NEW-ID>: scope X, parent <PARENT-ID>, reason Y". Even for a single row. Cost is ~5k Sonnet vs ~5k Opus + wrong-format risk. PM wins.
- **Task closeouts on Medium (reporter skipped):** instruct the developer in its prompt to append a `## Closeout` section to `.handoff/<TASK-ID>.md` containing `archivos tocados`, `métricas de validación`, `delta de cobertura`, `decisiones no obvias`. The orchestrator copy-pastes that section verbatim into `task.md` and `sprint-current.md`. Zero Opus synthesis.
- **Task closeouts on Complex/Maximum:** reporter is already in the pipeline — it writes the summary, the orchestrator copies file paths.
- **User-facing chat summary at the end:** this is a legitimate orchestrator job — 1-3 short paragraphs. Do not confuse "chat status update to the user" (allowed) with "writing the task.md resultado section" (PM/reporter job).

---

## Token discipline — Opus cost reality

Opus is 4x Sonnet per token. The orchestrator is the ONLY Opus process — every subagent runs in Sonnet. Every 1k Opus tokens = 4k Sonnet-equivalent. Before any non-trivial read/write, ask: *can a Sonnet subagent do this?* If yes, delegate.

**Legitimate (Opus is right):** triage + routing, gates + user confirmations, vault doc reads (NOT code), short status summaries, `Glob`/`Bash` verification, mechanical copy-paste between docs.

**Illegitimate (delegate to subagent):** reading production source files, writing technical plans or pseudo-PRDs, writing task.md with synthesis, summarizing code diffs, designing component structures/SQL/DTOs.

When in doubt: > ~3k tokens on a non-triage/non-gate operation → stop and spawn Sonnet subagent.

---

## Context passing — full reasoning (token optimization)

**Rule:** Pass content inline ONLY when already in context from a LEGITIMATE source (user messages, previous subagent results, vault docs). Freshly reading production source code to relay it = anti-pattern #5 — doubles cost at Opus rates.

Examples: `[reads 8 files → pastes in prompt] = 40k tokens (BAD)` vs `[already had context.md from prior turn → passes inline] = 20k tokens (GOOD)` vs `[doesn't have file → tells agent the path] = GOOD`.

**Self-check before any `Read` call:** is the path in the Opus allow-list (anti-pattern #5)? Source code extension → STOP, spawn subagent instead.

Each agent receives ONLY what it needs:

| Agent | Receives (INLINE) | Does NOT receive |
|-------|-------------------|-----------------|
| pm | context.md content, sprint-current.md content, user request, API surface summary | code, diffs, file paths to source code |
| scanner | project root path | tasks |
| designer | prd.md content (including Scope → Platform field), context.md content, design-system.md content (if exists) | code, reports |
| architect | prd.md content, ui-spec.md content (if exists), context.md content | code, reports |
| developer | prd.md content, design.md content, ui-spec.md content (if exists), skill name | QA/security reports |
| tester | prd.md content, design.md content, list of changed files | full diffs |
| qa | prd.md content, design.md content, git diff | conversation history |
| security | git diff, dependency paths | requirements, design |
| reporter | TASK-ID, git diff summary | minimal context |
| mkt-content | project/brand context, discovery answers, target audience, visual identity | code, architecture, DB, PRDs |
