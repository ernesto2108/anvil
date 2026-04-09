---
name: wails-recipes
description: Reusable patterns for Anvil Dashboard development with Wails v2 + Go backend + React frontend + SQLite. Use for dashboard features (DASH-FEAT-*) to skip pattern re-derivation. Covers Wails bindings, DTOs, store queries, React views, custom React Flow nodes, and the build tag layout.
---

# Wails Dashboard Recipes (Anvil)

Snippets and patterns proven in `DASH-FEAT-001..006`. Load this skill when the orchestrator invokes you for dashboard work — it replaces reading the whole codebase to re-derive patterns.

## Repo layout (dashboard scope)

```
internal/dashboard/
├── app.go                   // Wails App with bindings — build tag: dashboard
├── dtos.go                  // DTOs exposed to frontend — build tag: dashboard
├── store.go                 // Store interface (defined in package dashboard) — build tag: dashboard
└── store/
    ├── store.go             // SQLiteStore struct + New / NewFS constructors — NO build tag
    ├── read.go              // Read queries (ListRuns, ListAgentsByRun, ...) — NO build tag
    ├── migrate.go           // Migration runner (iofs) — NO build tag
    └── *_test.go            // Tests — NO build tag

frontend/src/
├── App.tsx                  // View switcher
├── lib/
│   ├── wails.ts             // Bindings + types 1:1 with dtos.go (camelCase)
│   ├── format.ts            // formatDuration, formatTokens, formatTokensFlow
│   └── utils.ts             // cn helper
├── components/
│   ├── ui/                  // shadcn primitives (button, card, table)
│   ├── layout/              // app-shell, sidebar, topbar
│   ├── status-badge.tsx     // Reusable: success|failed|running|pending
│   └── flow/                // React Flow custom nodes
└── views/                   // runs-view.tsx, flow-view.tsx, ...

migrations/                  // Shared with CLI — NO build tag
  000001_create_runs.up.sql
  000002_create_agents.up.sql
  000003_create_files.up.sql
  000004_create_events.up.sql
```

## Build tag rules (memorize)

| Layer | Build tag | Reason |
|---|---|---|
| `internal/dashboard/app.go`, `dtos.go`, `store.go` (interface) | `//go:build dashboard` | Wails dependency only links when tag set |
| `internal/dashboard/store/*.go` | NONE | Store is reusable by CLI for tooling; no Wails deps here |
| `internal/dashboard/store/*_test.go` | NONE | Tests run in default build |
| `internal/dashboard/*_test.go` (tests for app.go) | `//go:build dashboard` | They import the dashboard package |
| `cmd/anvil/*dashboard*.go` | `//go:build dashboard` / fallback stub with `//go:build !dashboard` | CLI subcommand only active with tag |

**`go build -tags dashboard ./...` is enough to validate compilation. You do NOT need test files for build validation.** Same for `go vet -tags dashboard ./...`.

## Recipe 1 — Add a new read query in the store

**When:** new binding requires reading data not already exposed.

**Pattern** (file: `internal/dashboard/store/read.go`, no build tag):

```go
// XRow is the projection for the X read path.
// Fields with sql.Null* are preserved as *T pointers to distinguish "zero" from "not set".
type XRow struct {
    Field1     string
    Field2     *int64   // nullable in DB
    Field3     int      // not null
    Field4     *float64 // nullable
}

// ListX returns rows ordered by <column> <ASC|DESC>.
// <param> rules: describe defaults and normalization.
func (s *SQLiteStore) ListX(ctx context.Context, param string) ([]XRow, error) {
    const q = `
        SELECT field1, field2, field3, field4
        FROM x_table
        WHERE <filter> = ?
        ORDER BY <column> ASC`

    rows, err := s.db.QueryContext(ctx, q, param)
    if err != nil {
        return nil, fmt.Errorf("dashboard/store: consultar x_table: %w", err)
    }
    defer rows.Close() //nolint:errcheck

    var results []XRow
    for rows.Next() {
        var (
            field1 string
            field2 sql.NullInt64
            field3 int
            field4 sql.NullFloat64
        )
        if err := rows.Scan(&field1, &field2, &field3, &field4); err != nil {
            return nil, fmt.Errorf("dashboard/store: escanear fila de x_table: %w", err)
        }

        r := XRow{Field1: field1, Field3: field3}
        if field2.Valid {
            v := field2.Int64
            r.Field2 = &v
        }
        if field4.Valid {
            v := field4.Float64
            r.Field4 = &v
        }
        results = append(results, r)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("dashboard/store: iterar filas de x_table: %w", err)
    }

    if results == nil {
        results = []XRow{}
    }
    return results, nil
}
```

**Rules:**
- Always wrap errors with `fmt.Errorf("dashboard/store: <verb> <table>: %w", err)` — consistent prefix
- Always return non-nil empty slice on empty results (frontend expects arrays, not `null`)
- Use `sql.Null*` in scan variables; convert to `*T` in the final struct
- Parse timestamps with `time.Parse(time.RFC3339Nano, raw)` — the DB stores them as TEXT in that format
- `defer rows.Close() //nolint:errcheck` is the idiom

## Recipe 2 — Add the Store interface method (package dashboard)

**Pattern** (file: `internal/dashboard/store.go`, build tag `dashboard`):

```go
//go:build dashboard

package dashboard

import (
    "context"

    "github.com/ernesto2108/anvil/internal/dashboard/store"
)

// Store is the read-only interface consumed by the Wails App.
type Store interface {
    ListRuns(ctx context.Context, limit, offset int) ([]store.RunSummary, error)
    ListAgentsByRun(ctx context.Context, runID string) ([]store.AgentRow, error)
    // ListX(ctx context.Context, param string) ([]store.XRow, error) // NEW
}
```

**Rule:** the interface is defined in package `dashboard`, not in package `store`. This keeps the Wails app decoupled from the concrete SQLite store for testing (you can substitute a fake in tests without importing sqlite).

## Recipe 3 — Add a DTO in `dtos.go`

**Pattern** (file: `internal/dashboard/dtos.go`, build tag `dashboard`):

```go
//go:build dashboard

package dashboard

// XDTO is the DTO exposed to the frontend via Wails bindings.
// JSON tags are camelCase (consistent with all existing dashboard DTOs and wails.ts).
type XDTO struct {
    ID        string   `json:"id"`
    Name      string   `json:"name"`
    Duration  int64    `json:"durationMs"` // 0 if not finished
    Score     *float64 `json:"score"`      // nullable
    Items     []YDTO   `json:"items"`
}
```

**Rules:**
- **Always camelCase** in JSON tags — not snake_case. The architecture doc may show snake_case; ignore it, follow the codebase
- Nullable fields use `*T` pointers — never use empty strings or zero as sentinels
- Slices must never be `nil` at the API boundary — convert with `make([]T, 0)` in the converter

## Recipe 4 — Add a Wails binding and its conversor

**Pattern** (file: `internal/dashboard/app.go`, build tag `dashboard`):

```go
//go:build dashboard

// GetX returns the X projection for the given param.
func (a *App) GetX(param string) ([]XDTO, error) {
    ctx := a.ctx
    if ctx == nil {
        ctx = context.Background()
    }
    rows, err := a.store.ListX(ctx, param)
    if err != nil {
        return nil, err
    }
    return toXDTOs(rows), nil
}

// toXDTOs converts store rows to DTOs. Keeps the conversion separate for testability.
func toXDTOs(rows []store.XRow) []XDTO {
    out := make([]XDTO, 0, len(rows))
    for _, r := range rows {
        var score *float64
        if r.Field4 != nil {
            v := *r.Field4
            score = &v
        }
        out = append(out, XDTO{
            ID:       r.Field1,
            Duration: derefInt64(r.Field2),
            Score:    score,
        })
    }
    return out
}

// derefInt64 returns the value or 0 if the pointer is nil. Used for "not finished" semantics.
func derefInt64(p *int64) int64 {
    if p == nil {
        return 0
    }
    return *p
}
```

**Rules:**
- Bindings always `return (T, error)` — Wails converts the error to `{ error: string }` in TS
- Bindings must guard `a.ctx == nil` with `context.Background()` fallback (needed for tests that don't call `Startup`)
- Converter is a private function named `to<DTO>s` (pluralized if list) or `to<DTO>` (if single)
- No mutation in converters — they are pure functions

## Recipe 5 — Extend `wails.ts` with types and a new function

**Pattern** (file: `frontend/src/lib/wails.ts`):

```ts
// Types correspond 1:1 with internal/dashboard/dtos.go (camelCase, not snake_case)

export interface XDTO {
  id: string
  name: string
  durationMs: number
  score: number | null
  items: YDTO[]
}

// Extend the window.go.dashboard.App declaration
declare global {
  interface Window {
    go?: {
      dashboard?: {
        App?: {
          GetRuns?: (q: RunsQuery) => Promise<RunDTO[]>
          GetFlow?: (runId: string) => Promise<FlowDTO>
          GetX?: (param: string) => Promise<XDTO[]>  // NEW
        }
      }
    }
  }
}

// getX calls the Wails binding. Falls back to [] in vite dev mode (no Wails runtime).
export async function getX(param: string): Promise<XDTO[]> {
  const binding = window.go?.dashboard?.App?.GetX
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando arreglo vacío (modo dev)')
    return []
  }
  return binding(param)
}
```

**Rules:**
- `nullable` fields in Go (`*T`) become `T | null` in TS
- Always provide a fallback (`[]`, `null`, or a default object) for the dev mode where `window.go` is undefined
- The warn message is non-fatal — intentional for vite dev

## Recipe 6 — Add a view

**Pattern** (file: `frontend/src/views/x-view.tsx`):

```tsx
import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { getX, type XDTO } from '@/lib/wails'

interface XViewProps {
  param: string
  onBack: () => void
}

export function XView({ param, onBack }: XViewProps) {
  const [data, setData] = useState<XDTO[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    getX(param)
      .then((d) => { if (!cancelled) { setData(d); setError(null) } })
      .catch((e) => { if (!cancelled) setError(String(e)) })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [param])

  if (loading) return <div className="p-6">Cargando…</div>
  if (error) return <div className="p-6 border border-fail-border rounded">Error: {error}</div>
  if (!data || data.length === 0) {
    return <div className="p-6 text-text-muted">Sin datos.</div>
  }

  return (
    <div className="h-full w-full flex flex-col">
      <header className="flex items-center gap-3 p-4 border-b border-border">
        <Button variant="outline" size="sm" onClick={onBack}>Volver</Button>
        <h2 className="text-base">Vista X</h2>
      </header>
      <div className="flex-1 p-6 overflow-auto">
        {/* render data */}
      </div>
    </div>
  )
}
```

**Rules:**
- Use the `cancelled` flag pattern in `useEffect` to avoid state updates after unmount
- Four states: loading / error / empty / data — always all four
- Empty state must have text (not just blank), use `text-text-muted`
- The view accepts props, not router state — `App.tsx` is the view switcher

## Recipe 7 — Custom React Flow node (for grafos)

**Pattern** (file: `frontend/src/components/flow/x-node.tsx`):

```tsx
import { Handle, Position, type NodeProps, type Node } from '@xyflow/react'
import { cn } from '@/lib/utils'
import { StatusBadge } from '@/components/status-badge'

// Data MUST be a type alias (not interface) to satisfy Record<string, unknown>
export type XNodeData = {
  label: string
  status: string
  // ... other fields
}

export type XNodeType = Node<XNodeData, 'xNode'>

export function XNode({ data }: NodeProps<XNodeType>) {
  const border = borderColorForStatus(data.status)
  return (
    <div
      role="button"
      tabIndex={0}
      aria-label={`X ${data.label}, estado ${data.status}`}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault()
          e.currentTarget.click()
        }
      }}
      className={cn(
        'flex items-center gap-2.5 px-4 py-3 rounded-lg bg-secondary border-2 w-[220px]',
        'cursor-pointer hover:bg-secondary/80',
        border,
      )}
    >
      <Handle type="target" position={Position.Left} className="!bg-transparent !border-0 !w-0 !h-0" />
      <div className="flex-1 min-w-0">
        <div className="text-[13px] font-semibold text-foreground truncate">{data.label}</div>
      </div>
      <StatusBadge status={data.status} />
      <Handle type="source" position={Position.Right} className="!bg-transparent !border-0 !w-0 !h-0" />
    </div>
  )
}

function borderColorForStatus(status: string): string {
  switch (status) {
    case 'success':
    case 'completed':
      return 'border-success-border'
    case 'failed':
    case 'fail':
    case 'error':
      return 'border-fail-border'
    case 'running':
    case 'in_progress':
    case 'in-progress':
      return 'border-running-border'
    default:
      return 'border-pending-border'
  }
}
```

**Rules:**
- `XNodeData` MUST be `type`, not `interface` (React Flow v12 needs `Record<string, unknown>` satisfaction, only type aliases satisfy it)
- Handles are required even if visually hidden — React Flow uses them as connection anchors
- A11y: `role="button"`, `tabIndex={0}`, `aria-label`, `onKeyDown` for Enter/Space
- Reuse `StatusBadge` for the status indicator — do NOT duplicate color/icon logic

## Recipe 8 — Dagre layout for React Flow

See `frontend/src/lib/dagre-layout.ts` (from DASH-FEAT-006). The helper `layoutFlow<T>` takes `Node<T>[] + Edge[]` and returns positioned nodes with `rankdir: 'LR'`. Reuse it directly — no need to reimplement per feature.

## Recipe 9 — CSS tokens (dark-only, already defined)

**Color tokens available in Tailwind classes** (see `frontend/src/index.css`):

| Purpose | Class |
|---|---|
| Background canvas | `bg-background` |
| Surface (card) | `bg-card` |
| Elevated surface | `bg-secondary` |
| Foreground (text primary) | `text-foreground` |
| Muted text | `text-text-muted` |
| Secondary text | `text-text-secondary` |
| Subtle border | `border-border` |
| Strong border | `border-border-strong` |
| Brand primary | `bg-brand` / `text-brand` / `border-ring` |
| Brand subtle (active bg) | `bg-brand-subtle` |
| Brand text (active text) | `text-brand-text` |
| Success | `text-success` / `bg-success-bg` / `border-success-border` |
| Fail | `text-fail` / `bg-fail-bg` / `border-fail-border` |
| Running (in-progress) | `text-running` / `bg-running-bg` / `border-running-border` |
| Pending | `text-pending` / `bg-pending-bg` / `border-pending-border` |

**Do NOT add new tokens without checking first.** All 4 statuses (success/fail/running/pending) have full bg+border+text tokens. Use `running` in code (the design uses `in-progress` but the codebase has standardized on `running` — they are semantic aliases).

## Recipe 10 — Security & supply chain

- `npm install` MUST always use `--ignore-scripts` (per DASH-SEC-002)
- Version pinning: look at existing `package.json` — pin exact versions (no `^`, no `~`) unless the package requires semver flexibility
- Never add deps that phone home, bundle analytics, or require network at runtime (this is a local desktop dashboard)

## What this skill replaces

When you load this skill, you can skip:
- Reading existing `app.go`, `dtos.go`, `store/read.go` to understand patterns
- Deriving the build tag layout by searching the repo
- Deriving the DTO naming convention (camelCase) by comparison
- Looking up the CSS token names in `index.css`
- Re-reading `frontend/src/lib/wails.ts` to understand the binding guard pattern
- Re-reading existing views to understand loading/error/empty state patterns

If you need the exact content of an existing file that this skill does not cover, read ONLY that file — do not do a broad exploration.
