---
name: go-conventions
description: Go backend conventions and coding standards. Use when writing Go code, reviewing Go patterns, or user mentions "go conventions", "idiomatic Go", "error handling in Go", "Go best practices", "Go testing patterns", or working with .go files.
---

# Go Conventions

> **IMPORTANT:** This file is a lightweight dispatcher. Do NOT load all referenced files at once. Read the routing table below, identify which files are relevant to the current task, and load ONLY those using the Read tool. Each file is ~3-5KB. Loading unnecessary files wastes context tokens.

## Stack & Philosophy

- **Go stdlib first** — only add dependencies when stdlib is genuinely insufficient
- **Simplicity over cleverness** — if it needs a comment to explain, simplify it
- **Explicit over implicit** — no magic, no init() side effects, no global state
- **Errors are values** — handle them, don't hide them
- **Composition over inheritance** — embed, don't extend

## Red Flags (always stop work)

- `panic()` outside `main()` → error
- `init()` doing real work → error
- Ignored errors (`_ = f()`) → error
- Global mutable state → error
- Resource leaks (unclosed tickers, deferred in loops) → error

## Anti-Pattern Detection

**Passive detection:** When reviewing Go code, load `detection/anti-patterns.md` and scan for `error` and `warning` patterns. Report as `[file:line] [severity] [category] anti-pattern-name`.

**Active detection:** When user asks to "improve", "refactor", "optimize", or "clean" — also report `suggestion` level patterns and propose fixes referencing the relevant rule or guide.

## What to Load

Load **only** the files relevant to the current task:

### Rules (quick reference, ~2-3KB each)

| Working on... | Load |
|---|---|
| Error handling, naming, context, concurrency basics | `rules/coding.md` |
| Imports, DTOs, validation, DI | `rules/architecture.md` |
| SQL, transactions, repositories | `rules/database.md` |
| Kafka/RabbitMQ critical rules | `rules/messaging.md` |
| Functional options, constructors, guard clauses | `rules/patterns.md` |

### Guides (detailed patterns with code, ~3-5KB each)

| Working on... | Load |
|---|---|
| Which concurrency primitive to use | `guides/concurrency/decision-matrix.md` |
| Fan-out/fan-in pattern | `guides/concurrency/fan-out-fan-in.md` |
| Worker pools, bounded concurrency | `guides/concurrency/worker-pools.md` |
| Pipeline stages | `guides/concurrency/pipelines.md` |
| Timeouts, context cancellation | `guides/concurrency/timeout-cancellation.md` |
| Rate limiting | `guides/concurrency/rate-limiting.md` |
| Graceful shutdown (HTTP, workers) | `guides/concurrency/graceful-shutdown.md` |
| Concurrent map access | `guides/concurrency/concurrent-map.md` |
| Pub/sub event broadcasting | `guides/concurrency/pub-sub.md` |
| Concurrency anti-patterns + checklist | `guides/concurrency/anti-patterns.md` |
| HTTP client, DB, Redis, gRPC context | `guides/cleanup/context-propagation.md` |
| Multi-level timeout design | `guides/cleanup/timeout-architecture.md` |
| Rows, transactions, HTTP body, tickers | `guides/cleanup/resource-cleanup.md` |
| sql.DB pool config, monitoring | `guides/cleanup/connection-pools.md` |
| Resource checklist, linters, production detection | `guides/cleanup/detection-checklist.md` |
| Test structure, table-driven tests | `guides/testing/structure-tables.md` |
| Test helpers, mocking with interfaces | `guides/testing/helpers-mocking.md` |
| Testing HTTP handlers (Gin) | `guides/testing/http-handlers.md` |
| Testing repositories (mock rows) | `guides/testing/repositories.md` |
| Fixtures, testdata, integration tests | `guides/testing/fixtures-integration.md` |
| Coverage, benchmarks | `guides/testing/coverage-benchmarks.md` |
| Kafka overview, library selection | `guides/kafka/overview.md` |
| Kafka producer patterns | `guides/kafka/producer.md` |
| Kafka consumer, groups, ordering | `guides/kafka/consumer.md` |
| Kafka DLQ, retry, poison messages | `guides/kafka/dlq-retry.md` |
| Kafka circuit breaker, idempotency, backpressure | `guides/kafka/resilience.md` |
| Kafka shutdown, tracing, schema, anti-patterns | `guides/kafka/operations.md` |
| RabbitMQ overview, exchanges, durability | `guides/rabbitmq/overview.md` |
| RabbitMQ connection, auto-reconnect | `guides/rabbitmq/connection.md` |
| RabbitMQ producer, confirms | `guides/rabbitmq/producer.md` |
| RabbitMQ consumer, QoS, ack/nack | `guides/rabbitmq/consumer.md` |
| RabbitMQ DLX/DLQ, TTL retry chains | `guides/rabbitmq/dlq-retry.md` |
| RabbitMQ backpressure, shutdown, anti-patterns | `guides/rabbitmq/operations.md` |
| Structured logging (slog) | `guides/slog.md` |
| Health checks, Prometheus, OpenTelemetry | `guides/observability.md` |
| HTTP middleware composition | `guides/middleware.md` |
| SQL injection, crypto, input validation | `guides/security.md` |

### Detection & Checklists

| When... | Load |
|---|---|
| Code review | `detection/anti-patterns.md` |
| Before writing Go code | `checklists/pre.md` |
| After writing Go code | `checklists/post.md` |

### Examples (good + bad patterns by domain, ~2-3KB each)

| Working on... | Load |
|---|---|
| Error handling patterns | `examples/errors.md` |
| Architecture, interfaces, DI | `examples/architecture.md` |
| Testing patterns | `examples/testing.md` |
| Database, repositories, DTOs | `examples/database.md` |
| Entity validation | `examples/validation.md` |
| Concurrency, shutdown | `examples/concurrency.md` |
| Handler → Service → Repo full flow, error wrapping rules | `examples/service-contracts.md` |

## Post-Implementation Gate

After ANY code change to `.go` files, invoke the `/lint` skill before considering the task done.
