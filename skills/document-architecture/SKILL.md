---
name: document-architecture
description: Document the architecture of a frontend or backend service. Auto-detects project type and runs the appropriate pipeline. Use when user says "documenta arquitectura", "document service", "documenta frontend", "document architecture", or when orchestrate routes here.
disable-model-invocation: true
---

# Document Architecture

Unified entry point for documenting any project. Detects type, loads the matching guide, and runs the pipeline.

## Step 0 — Input & Detection

If invoked without arguments, ask the user (in Spanish):
1. **Que proyecto?** — nombre del repo
2. **Que tarea del backlog?** — ID. Si no hay, preguntar si crear una.

Resolve `<docs>` from `~/.claude/project-registry.md`.
Resolve `<repo>` from `~/projects/<project-name>`.

### Auto-detect

Run `ls <repo>/` and determine:

| Root contains | Type |
|---|---|
| `package.json` + `src/` with `.tsx`/`.jsx` | **frontend** |
| `go.mod` + `internal/` (NO `domain/`) | **service** (MVC) |
| `go.mod` + `domain/` + `usecase/` | **service** (Clean) |
| `go.mod` + `internal/` with `business/` | **service** (Hex) |
| Ambiguous | Ask: "Frontend o backend?" |

**Load the matching guide** from `guides/frontend-pipeline.md` or `guides/service-pipeline.md`.

## Step 1 — Verify output pattern

Check `<docs>/04-architecture/` for the expected file structure. The guide specifies which files to generate.

## Step 2 — Decide security

Ask yourself (NOT the user): does this project handle auth, payments, PII, or sensitive data?
- **Yes** → include security in pipeline
- **No** → skip security

## Step 3 — Scanner (deep, skeleton-aware)

Read the skeleton file specified in the guide. Inject INLINE into the scanner prompt. Launch `scanner` agent with `mode: deep` following the guide's scanner instructions.

- Model: **sonnet**
- **Target: <25 tool calls**
- After completion: **Read the context files** with Read tool.

## Step 4 — Architect (2 agents in parallel)

Launch TWO architect agents following the guide:
- **4a — Overview:** inject context-summary.md INLINE
- **4b — Detail:** inject template + context-detail.md INLINE

- Model: **sonnet**
- **Target: 0 Read calls** for detail agent (everything inline)

**Wait for both to finish before Step 5.** Read overview.md for security summary.

## Step 5 — Security [conditional]

Skip if Step 2 said no. Read `known-systemic-issues.md`, inject INLINE with context-risks.md + overview summary. Launch `security` agent following the guide's security instructions.

- Model: **sonnet**
- **Target: <10 tool calls**

## Step 6 — Close task

1. Update task status to `done`
2. Update board.md
3. Remove backlog duplicate if exists
4. Update sprint metrics

## Rules

- **All output in Spanish** — titles, descriptions, Mermaid labels. Code/paths in English.
- **Context injection MANDATORY** — each agent gets ONLY its segment INLINE.
- **Model: sonnet** for all agents.
- **Mermaid: NEVER use `|` inside labels or messages.** Use `/` instead.
- **Token budget: <50 tool calls total.**
