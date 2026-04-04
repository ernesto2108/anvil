# Scanner Recipe: Go + Gin

Extraction recipe for Go microservices using Gin framework. Follow this order exactly.
Use Grep over Read whenever possible. Only Read full files when grep context is insufficient.

## Phase 1 — Skeleton (5 tool calls max)

```
1. Glob "**/*.go" → directory structure (0 reads)
2. Read pkg/server/gin/gin-router.go → ALL endpoints in one file (<50 lines)
3. Read pkg/container/container.go limit:80 → ALL injected dependencies
4. Read go.mod limit:40 → only the require block
5. Read pkg/config/*.go limit:30 → config structure (interface methods = keys)
```

After Phase 1 you have: endpoints list, dependency map, tech stack, config keys.

## Phase 2 — Endpoint flows (grep-first)

For EACH endpoint found in gin-router.go:

```
Step A — Handler inputs (1 grep per handler file):
  grep "BindJSON\|BindQuery\|BindUri\|Param\|GetHeader\|ParseUint" handlers/*.go
  → extracts ALL bindings from ALL handlers in ONE call

Step B — Service calls (1 grep per service file):
  grep "func.*Service\)\|Repo\.\|Service\.\|Redis\|HTTP\|gRPC\|Kafka\|SNS\|Send" services/*.go
  → extracts ALL method signatures and external calls in ONE call

Step C — SQL queries (1 grep per queries dir):
  grep "SELECT\|INSERT\|UPDATE\|DELETE\|CALL" repositories/psql/queries/*.go
  → extracts ALL SQL statements in ONE call

Step D — External clients (1 grep per client dir):
  grep "func\|http\.\|url\|endpoint\|Timeout\|grpc\." repositories/client/*.go
  → extracts ALL external call signatures in ONE call
```

**Total Phase 2: ~4-6 grep calls** instead of reading every file.

### When to Read full functions

Only Read a specific function (with offset+limit) when:
- The grep context (default 2 lines) doesn't show what the function DOES (business logic)
- The function has complex branching (if/else chains, state machines, goroutines)
- The function calls multiple services in sequence (payment flow, notification fan-out)

Use `grep -C 5 "functionName" file.go` first. If that's enough context, don't Read.

## Phase 3 — Risks (2-3 targeted greps)

```
grep "_ :=\|_ =\|goto \|panic(\|log\.Print" internal/ → errors ignored, gotos, panics
grep "defer.*Close\|rows\.Close\|resp\.Body\.Close" internal/ → check for missing defers
grep "AllowOrigins\|AllowCredentials\|Recovery" pkg/server/ → CORS and middleware issues
```

## Phase 4 — Detect and skip in pkg/

`pkg/` varies between services. Some have Kafka, gRPC, SNS — others don't.

**Step 1: Detect what's there (1 glob call):**
```
Glob "pkg/**/*.go" → list all pkg subdirs
```

**Step 2: Classify each subdir:**

| Dir | Action | Reason |
|-----|--------|--------|
| `pkg/error/` | SKIP | Standard error types — shared across services |
| `pkg/log/` | SKIP | Logger wrapper — shared across services |
| `pkg/monitor/` | SKIP | APM wrappers — shared across services |
| `pkg/cloud/aws/` | SKIP | SSM/S3 session — shared across services |
| `pkg/config/` | READ limit:30 | Config interface → keys only |
| `pkg/container/` | READ full | DI wiring — unique per service, tells you ALL deps |
| `pkg/server/gin/` | READ router + grep middleware | Routes + CORS + middleware chain |
| `pkg/output/http/` | SKIP unless container references a unique HTTP client | Generic client — shared |
| `pkg/output/broker/kafka/` | GREP for topic names | Only if present — not all services have Kafka |
| `pkg/output/broker/sns/` | GREP for topic ARN patterns | Only if present |
| `pkg/data/postgresql/` | SKIP | Standard pool wrapper — shared across services |
| `pkg/data/redis/` | SKIP | Standard Redis client — shared across services |
| `pkg/output/notify/slack/` | SKIP | Standard webhook — shared across services |

**Step 3: For service-specific infra in pkg/**, grep for config and connections:
```
# Only for dirs that exist and aren't in the SKIP list:
grep "Topic\|Queue\|Broker\|Producer\|Consumer" pkg/output/broker/ → Kafka/SNS topics
grep "grpc\.\|Dial\|NewClient" pkg/ → gRPC connections
```

## Token budget

| Phase | Tool calls | Tokens (est.) |
|-------|-----------|---------------|
| Skeleton | 5 Read | ~5k |
| Endpoint flows | 4-6 Grep + 3-5 targeted Read | ~10k |
| Risks | 3 Grep | ~2k |
| Writes (3 files) | 3 Write | ~3k |
| **Total** | **~15-20 calls** | **~20-25k** |

Compare: without recipe, scanner uses ~38 calls / ~69k tokens.
