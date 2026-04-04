---
name: cross-service-dev
description: Orchestrate agents across multiple microservice repos in a single session. Use when user says "implement across services", "this touches X and Y services", "cross-service feature", "work on multiple repos", "remove this endpoint from all services", "deprecate this across services", "refactor cross-service", or describes any change (create, update, delete, deprecate) that requires coordinated work in 2+ services. Extends the orchestrate workflow for multi-repo scenarios. Requires service-map.yaml to resolve repo paths.
disable-model-invocation: true
---

# Cross-Service Dev — Multi-Repo Orchestration

## Purpose

Extend the `orchestrate` workflow to coordinate agents across multiple microservice repos in one session. Same agents, same gates — The orchestrator resolves paths, discovers dependencies, and routes agents to the right repos.

```
orchestrate          = pipeline for 1 repo
cross-service-dev    = orchestrate × N repos (coordinated)
```

## Prerequisites

- `service-map.yaml` exists in `<vault>/04-architecture/`
- Affected service repos are on disk (local_path must resolve)
- If `service-map.local.yaml` exists, use it for local path overrides

---

## Workflow

### Phase 1 — PM Discovery (pm agent)

Invoke pm agent with the user request + vault path. PM handles discovery in Spanish.

Additionally, The orchestrator must:

1. **Classify operation type:** Create | Update | Delete | Deprecate

2. **Resolve service paths from service-map.yaml:**
   ```
   full_path = projects_root + "/" + service.local_path
   Verify each path exists on disk. If missing → warn user, ask for path.
   ```

3. **Discover transitive dependencies (MANDATORY):**
   For each service being changed, check service-map.yaml:
   - Who consumes this endpoint? (consumed_by)
   - Who subscribes to this event? (consumed_by in publishes)
   - Who reads this table? (readers in shared_database)
   - Who depends on this service? (depends_on)

   **If additional services found → STOP and report to user before proceeding.**
   Rules:
   - NEVER silently skip affected services
   - DELETE/DEPRECATE → transitive check is CRITICAL
   - UPDATE with contract changes → all consumers are affected

4. PM writes **one** `prd.md` in `<vault>/03-tasks/<TASK-ID>/prd.md`
   - Must list ALL services in scope under Dependencies
   - Must note skipped services as pending
   - Must specify operation type

### Phase 2 — Architecture (1 architect agent)

One architect receives:
- `<vault>/03-tasks/<TASK-ID>/prd.md`
- `<vault>/01-project/context.md` from **each** service in scope

Produces one `<vault>/03-tasks/<TASK-ID>/design.md` with:
- A section per service
- Contract definitions (shared)
- Execution order
- Migration ownership

**GATE: architect veto → STOP**

### Phase 3 — Implementation (N developer agents)

One developer agent per service. Each receives:
- `<vault>/03-tasks/<TASK-ID>/prd.md`
- `<vault>/03-tasks/<TASK-ID>/design.md`
- Convention skill to load
- Their specific service path as working directory

**Parallelism:** independent services → parallel. If B depends on A's output → sequential.
**DBA:** 0-1 agent, runs before developers if migration needed.
**DELETE operations:** reverse order — consumers first, producer last.

### Phase 4 — Testing (N tester agents, parallel)

One tester per modified service. All run in parallel.

### Phase 5 — QA (1 QA agent)

One QA agent sees combined diff from all services. Focus on:
- Contract consistency between producer and consumers
- Event payload matches
- API type alignment
- DB consistency across shared tables

**GATE: score < 7 → STOP**

### Phase 6 — Document + Report

**6a.** Append to `<vault>/06-reports/cross-service-changes.md`:
- Date, operation, scope, changes per service, contracts, pending work, deploy order

**6b.** Update `<vault>/04-architecture/service-map.yaml` to reflect new state

**6c.** Reporter agent → `<vault>/06-reports/last-run.md`

---

## Agent routing summary

| Phase | Agent | Count | Parallel? |
|-------|-------|-------|-----------|
| PM | pm | 1 | — |
| Architecture | architect | 1 | — |
| DB migration | dba | 0-1 | — |
| Implementation | developer | N | Yes (when independent) |
| Testing | tester | N | Yes |
| QA | qa | 1 | — |
| Security | security | 0-1 | — |
| Report | reporter | 1 | — |

## Key Rules

1. Same agents, same gates as orchestrate
2. Architect and QA see ALL services — full cross-service context
3. Developer and Tester are per-service — guided by consolidated design.md
4. NEVER silently skip affected services
5. Delete order is reverse of create — consumers first, producer last
6. All docs centralized in vault — no duplication across repos
