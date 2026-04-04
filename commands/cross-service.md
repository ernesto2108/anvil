---
name: cross-service
description: Orchestrate a feature/change across multiple microservice repos using the full agent pipeline
allowed-tools: Agent, Read, Glob, Grep, Bash, Edit, Write
---

# Cross-Service Development — Multi-Repo Pipeline

Load the `cross-service-dev` skill and follow its workflow.

## Context to gather first:

1. Read `<vault>/04-architecture/service-map.yaml` from the project vault
2. If it doesn't exist, ask the user to create one (reference the template in the service-map skill)

## What the user said: $ARGUMENTS

If no arguments provided, ask the user (in Spanish):
- ¿Qué cambio necesitás? (nuevo endpoint, modificar existente, eliminar, deprecar)
- ¿Qué servicios están involucrados?
- ¿Alguna restricción? (backwards compatibility, shared DB, deadline)

## Then follow the cross-service-dev skill phases:

1. **PM** — pm agent: classify operation, discovery, write prd.md
2. **Architect** — 1 agent, consolidated design.md
3. **Developer** — N agents (1 per service, parallel when possible)
4. **DBA** — 0-1 agent (only if migration needed)
5. **Tester** — N agents (1 per service, parallel)
6. **QA** — 1 agent (combined diff from all services)
7. **Document** — changelog, diagram, update service-map.yaml
8. **Reporter** — final summary

Gates apply: architect veto → STOP, QA < 7 → STOP.
