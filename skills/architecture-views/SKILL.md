---
name: architecture-views
description: Templates and format guides for architecture document views. Loaded by the architect agent to produce domain-specific architecture files with executable specs (SDD). Use when the architect needs view templates.
disable-model-invocation: true
---

# Architecture Views — Templates & Format Guide

Reference skill for the architect agent. Provides templates for each architecture view with Spec Driven Development (SDD) formats.

## Philosophy

Architecture documents serve two audiences:
1. **Humans** — developers, reviewers, future maintainers need context, rationale, trade-offs
2. **Machines** — agents, CI, code generators need executable specs they can consume and validate

Every view balances both. Narrative sections explain "why"; spec sections define "what" in machine-readable format.

## When to use each format

| Complexity | Output |
|---|---|
| Small (2 pts), single-stack, no contracts | `architecture.md` only — narrative with diagrams |
| Medium (5 pts), single-stack with DB or API | `architecture.md` + relevant domain view |
| Medium+ with cross-stack contracts | `architecture.md` + all applicable views with executable specs |
| Large (8+ pts), multi-service | All applicable views, full SDD specs, contract bridge |

## Guides — load per view

Each guide contains the template + format rules for one view. Load ONLY the guides relevant to the task.

| View | Guide | When to load |
|---|---|---|
| Overview | `guides/overview.md` | Always |
| Backend | `guides/backend.md` | Backend work |
| Frontend | `guides/frontend.md` | Frontend work |
| Database | `guides/database.md` | DB changes |
| Infrastructure | `guides/infrastructure.md` | Infra changes |

## Cross-view contract consistency

When the architect generates multiple views, contracts MUST be consistent:

1. **Backend OpenAPI schema ↔ Frontend TypeScript interface** — same field names, same types, same required/optional
2. **Backend persistence types ↔ DB schema intent** — same columns, same types, same constraints
3. **Infra env vars ↔ Backend config references** — same variable names

**Rule:** Define the contract ONCE in the primary view (usually backend), then reference or derive in secondary views. Never duplicate with different shapes.

## Validation checklist (architect self-check before closing)

- [ ] Every decision in `architecture.md` has a rationale ("why")
- [ ] Cross-view contracts are consistent (same shapes)
- [ ] All referenced file paths verified with Glob/Grep
- [ ] New files/paths marked as `NEW`
- [ ] OpenAPI spec is valid YAML (if present)
- [ ] DBML/DDL is syntactically correct (if present)
- [ ] Convention rules don't contradict the architecture
- [ ] Diagrams are readable (no more than 15 nodes per diagram)
