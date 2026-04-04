---
name: service-map
description: Cross-service dependency awareness for microservices. Use when modifying endpoints, DB schemas, shared contracts, or any code that other services consume. Triggers on "service map", "who uses this endpoint", "impact analysis", "cross-service", "dependency check", "what services depend on", "before refactoring endpoint", or when working in a project that has a service-map.yaml file.
---

# Service Map — Cross-Service Dependency Awareness

## Purpose

Prevent breaking changes across microservices by checking dependencies **before** modifying endpoints, DB schemas, shared contracts, or inter-service communication.

## Service Map Location

```
<vault>/04-architecture/service-map.yaml
```

If no service map exists, prompt the user to create one using the template in `service-map-template.yaml`.

## Pre-Change Flow

**Step 1 — Identify what's changing**
- HTTP endpoint (route, request/response schema, status codes)
- DB table/column (schema change, migration)
- Event/message (payload structure, topic)
- Shared library/package
- Environment variable or config

**Step 2 — Consult the service map**
- Look up the resource in service-map.yaml
- Identify all consumers and the owner
- Resolve local paths: `projects_root` + service's `local_path`

**Step 3 — Report impact**
```
## Impact Analysis
**Changing:** [what]
**Owner:** [service]
**Consumers:** [list]

### Breaking changes:
- [what could break and where]

### Safe changes:
- [additive/non-breaking]

### Recommended approach:
- [steps to change safely]
```

**Step 4 — Inspect affected services**
- Use resolved paths to locate consumer repos on disk
- Read consumer code to find WHERE the dependency exists
- Report exact file and line
- If repo not found on disk, warn the user

**Step 5 — If uncertain, ASK**
- Never assume a change is safe — verify or ask

## Change Safety Rules

**Always safe:** new optional response field, new endpoint, new event topic, new DB column with default

**Potentially breaking (requires consumer check):** removing/renaming response field, changing field type, changing URL/method, stricter validation, changing DB column type, changing event payload

**Always breaking (coordinated deploy):** removing endpoint, changing auth, renaming shared DB table, changing event topic name

## Recommended Patterns

**Endpoint versioning:** keep v1 unchanged, add v2

**Expand-and-contract for DB:** add new column → migrate consumers → backfill → remove old

**Deprecation flow:** mark in service-map.yaml with `status: deprecated`, `deprecated_since`, `replacement`, `consumed_by`

## Cross-Stack Awareness

| Stack | What to check |
|-------|--------------|
| Backend (Go) | HTTP endpoints, gRPC protos, DB schemas, event producers |
| Frontend (React) | API calls, shared types, environment configs |
| Mobile (Flutter) | API calls, push notification contracts, deep link schemas |
| Infrastructure | Environment variables, secrets, DNS, load balancer routes |

## When to Trigger

- User modifies endpoints, APIs, DB schemas, protos, shared types
- Always ask: "Does any other service consume what I'm about to change?"
- If yes → consult service map. If no map → ask the user.

## Schema Reference

See `service-map-template.yaml` for full schema. Key sections: `services`, `shared_databases`, `events`, `shared_contracts`.
