---
name: wails-recipes
description: Patrones reutilizables para el desarrollo del Anvil Dashboard con Wails v2 + backend Go + frontend React + SQLite. Úsalo para features del dashboard (DASH-FEAT-*) para evitar re-derivar patrones. Cubre bindings Wails, DTOs, queries del store, vistas React, nodos personalizados de React Flow, y el layout de build tags.
---

# Wails Dashboard Recipes (Anvil)

Snippets y patrones probados en `DASH-FEAT-001..006`. Cargar esta skill cuando el orquestador te invoque para trabajo del dashboard — reemplaza la necesidad de leer toda la base de código para re-derivar patrones.

## Layout del repositorio (scope del dashboard)

```
internal/dashboard/
├── app.go                   // Wails App con bindings — build tag: dashboard
├── dtos.go                  // DTOs expuestos al frontend — build tag: dashboard
├── store.go                 // Interfaz Store (definida en package dashboard) — build tag: dashboard
└── store/
    ├── store.go             // Struct SQLiteStore + constructores New / NewFS — SIN build tag
    ├── read.go              // Queries de lectura (ListRuns, ListAgentsByRun, ...) — SIN build tag
    ├── migrate.go           // Runner de migración (iofs) — SIN build tag
    └── *_test.go            // Tests — SIN build tag

frontend/src/
├── App.tsx                  // Selector de vistas
├── lib/
│   ├── wails.ts             // Bindings + tipos 1:1 con dtos.go (camelCase)
│   ├── format.ts            // formatDuration, formatTokens, formatTokensFlow
│   └── utils.ts             // helper cn
├── components/
│   ├── ui/                  // Primitivos shadcn (button, card, table)
│   ├── layout/              // app-shell, sidebar, topbar
│   ├── status-badge.tsx     // Reutilizable: success|failed|running|pending
│   └── flow/                // Nodos personalizados de React Flow
└── views/                   // runs-view.tsx, flow-view.tsx, ...

migrations/                  // Compartido con CLI — SIN build tag
  000001_create_runs.up.sql
  000002_create_agents.up.sql
  000003_create_files.up.sql
  000004_create_events.up.sql
```

## Reglas de build tags (memorizar)

| Capa | Build tag | Razón |
|---|---|---|
| `internal/dashboard/app.go`, `dtos.go`, `store.go` (interfaz) | `//go:build dashboard` | La dependencia Wails solo se linkea cuando el tag está activado |
| `internal/dashboard/store/*.go` | NINGUNO | El store es reutilizable por el CLI para tooling; sin dependencias Wails aquí |
| `internal/dashboard/store/*_test.go` | NINGUNO | Los tests corren en el build por defecto |
| `internal/dashboard/*_test.go` (tests para app.go) | `//go:build dashboard` | Importan el package dashboard |
| `cmd/anvil/*dashboard*.go` | `//go:build dashboard` / stub de fallback con `//go:build !dashboard` | Subcomando del CLI solo activo con el tag |

**`go build -tags dashboard ./...` es suficiente para validar la compilación. NO necesitas archivos de test para la validación del build.** Lo mismo para `go vet -tags dashboard ./...`.

## Receta 1 — Agregar una nueva query de lectura en el store

**Cuándo:** un nuevo binding requiere leer datos que aún no están expuestos.

**Patrón** (archivo: `internal/dashboard/store/read.go`, sin build tag):

```go
// XRow es la proyección para el camino de lectura de X.
// Los campos con sql.Null* se preservan como punteros *T para distinguir "cero" de "no establecido".
type XRow struct {
    Field1     string
    Field2     *int64   // nullable en BD
    Field3     int      // not null
    Field4     *float64 // nullable
}

// ListX retorna filas ordenadas por <columna> <ASC|DESC>.
// Reglas de <param>: describir defaults y normalización.
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

**Reglas:**
- Siempre envolver errores con `fmt.Errorf("dashboard/store: <verbo> <tabla>: %w", err)` — prefijo consistente
- Siempre retornar un slice vacío no-nil en resultados vacíos (el frontend espera arrays, no `null`)
- Usar `sql.Null*` en variables de scan; convertir a `*T` en el struct final
- Parsear timestamps con `time.Parse(time.RFC3339Nano, raw)` — la BD los almacena como TEXT en ese formato
- `defer rows.Close() //nolint:errcheck` es el idioma

## Receta 2 — Agregar el método de interfaz Store (package dashboard)

**Patrón** (archivo: `internal/dashboard/store.go`, build tag `dashboard`):

```go
//go:build dashboard

package dashboard

import (
    "context"

    "github.com/ernesto2108/anvil/internal/dashboard/store"
)

// Store es la interfaz de solo lectura consumida por la Wails App.
type Store interface {
    ListRuns(ctx context.Context, limit, offset int) ([]store.RunSummary, error)
    ListAgentsByRun(ctx context.Context, runID string) ([]store.AgentRow, error)
    // ListX(ctx context.Context, param string) ([]store.XRow, error) // NUEVO
}
```

**Regla:** la interfaz se define en el package `dashboard`, no en el package `store`. Esto mantiene la Wails app desacoplada del store SQLite concreto para pruebas (puedes sustituir un fake en los tests sin importar sqlite).

## Receta 3 — Agregar un DTO en `dtos.go`

**Patrón** (archivo: `internal/dashboard/dtos.go`, build tag `dashboard`):

```go
//go:build dashboard

package dashboard

// XDTO es el DTO expuesto al frontend via bindings Wails.
// Los JSON tags son camelCase (consistente con todos los DTOs del dashboard existentes y wails.ts).
type XDTO struct {
    ID        string   `json:"id"`
    Name      string   `json:"name"`
    Duration  int64    `json:"durationMs"` // 0 si no ha terminado
    Score     *float64 `json:"score"`      // nullable
    Items     []YDTO   `json:"items"`
}
```

**Reglas:**
- **Siempre camelCase** en JSON tags — no snake_case. El doc de arquitectura puede mostrar snake_case; ignorarlo, seguir la base de código
- Los campos nullable usan punteros `*T` — nunca usar strings vacíos o cero como centinelas
- Los slices nunca deben ser `nil` en la frontera de la API — convertir con `make([]T, 0)` en el conversor

## Receta 4 — Agregar un binding Wails y su conversor

**Patrón** (archivo: `internal/dashboard/app.go`, build tag `dashboard`):

```go
//go:build dashboard

// GetX retorna la proyección X para el param dado.
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

// toXDTOs convierte filas del store a DTOs. Mantiene la conversión separada para testeabilidad.
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

// derefInt64 retorna el valor o 0 si el puntero es nil. Usado para semántica de "no terminado".
func derefInt64(p *int64) int64 {
    if p == nil {
        return 0
    }
    return *p
}
```

**Reglas:**
- Los bindings siempre `return (T, error)` — Wails convierte el error a `{ error: string }` en TS
- Los bindings deben proteger `a.ctx == nil` con fallback `context.Background()` (necesario para tests que no llaman `Startup`)
- El conversor es una función privada nombrada `to<DTO>s` (pluralizado si es lista) o `to<DTO>` (si es único)
- Sin mutación en conversores — son funciones puras

## Receta 5 — Extender `wails.ts` con tipos y una nueva función

**Patrón** (archivo: `frontend/src/lib/wails.ts`):

```ts
// Los tipos corresponden 1:1 con internal/dashboard/dtos.go (camelCase, no snake_case)

export interface XDTO {
  id: string
  name: string
  durationMs: number
  score: number | null
  items: YDTO[]
}

// Extender la declaración window.go.dashboard.App
declare global {
  interface Window {
    go?: {
      dashboard?: {
        App?: {
          GetRuns?: (q: RunsQuery) => Promise<RunDTO[]>
          GetFlow?: (runId: string) => Promise<FlowDTO>
          GetX?: (param: string) => Promise<XDTO[]>  // NUEVO
        }
      }
    }
  }
}

// getX llama al binding Wails. Retorna [] en modo dev de vite (sin runtime Wails).
export async function getX(param: string): Promise<XDTO[]> {
  const binding = window.go?.dashboard?.App?.GetX
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando arreglo vacío (modo dev)')
    return []
  }
  return binding(param)
}
```

**Reglas:**
- Los campos `nullable` en Go (`*T`) se convierten en `T | null` en TS
- Siempre proveer un fallback (`[]`, `null`, o un objeto por defecto) para el modo dev donde `window.go` es undefined
- El mensaje de warn no es fatal — intencional para vite dev

## Receta 6 — Agregar una vista

**Patrón** (archivo: `frontend/src/views/x-view.tsx`):

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
        {/* renderizar datos */}
      </div>
    </div>
  )
}
```

**Reglas:**
- Usar el patrón con flag `cancelled` en `useEffect` para evitar actualizaciones de estado después de desmontar
- Cuatro estados: loading / error / empty / data — siempre los cuatro
- El estado vacío debe tener texto (no solo en blanco), usar `text-text-muted`
- La vista acepta props, no estado del router — `App.tsx` es el selector de vistas

## Receta 7 — Nodo personalizado de React Flow (para grafos)

**Patrón** (archivo: `frontend/src/components/flow/x-node.tsx`):

```tsx
import { Handle, Position, type NodeProps, type Node } from '@xyflow/react'
import { cn } from '@/lib/utils'
import { StatusBadge } from '@/components/status-badge'

// Data DEBE ser un type alias (no interface) para satisfacer Record<string, unknown>
export type XNodeData = {
  label: string
  status: string
  // ... otros campos
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

**Reglas:**
- `XNodeData` DEBE ser `type`, no `interface` (React Flow v12 necesita satisfacer `Record<string, unknown>`, solo los type aliases lo satisfacen)
- Los Handles son obligatorios aunque estén visualmente ocultos — React Flow los usa como anclas de conexión
- A11y: `role="button"`, `tabIndex={0}`, `aria-label`, `onKeyDown` para Enter/Space
- Reutilizar `StatusBadge` para el indicador de estado — NO duplicar la lógica de color/icono

## Receta 8 — Layout Dagre para React Flow

Ver `frontend/src/lib/dagre-layout.ts` (de DASH-FEAT-006). El helper `layoutFlow<T>` toma `Node<T>[] + Edge[]` y retorna nodos posicionados con `rankdir: 'LR'`. Reutilizarlo directamente — no es necesario reimplementar por feature.

## Receta 9 — Tokens CSS (solo dark, ya definidos)

**Tokens de color disponibles en clases Tailwind** (ver `frontend/src/index.css`):

| Propósito | Clase |
|---|---|
| Canvas de fondo | `bg-background` |
| Superficie (card) | `bg-card` |
| Superficie elevada | `bg-secondary` |
| Foreground (texto primario) | `text-foreground` |
| Texto atenuado | `text-text-muted` |
| Texto secundario | `text-text-secondary` |
| Borde sutil | `border-border` |
| Borde fuerte | `border-border-strong` |
| Marca primaria | `bg-brand` / `text-brand` / `border-ring` |
| Marca sutil (fondo activo) | `bg-brand-subtle` |
| Texto de marca (texto activo) | `text-brand-text` |
| Éxito | `text-success` / `bg-success-bg` / `border-success-border` |
| Fallo | `text-fail` / `bg-fail-bg` / `border-fail-border` |
| En ejecución (in-progress) | `text-running` / `bg-running-bg` / `border-running-border` |
| Pendiente | `text-pending` / `bg-pending-bg` / `border-pending-border` |

**NO agregar nuevos tokens sin verificar primero.** Los 4 estados (success/fail/running/pending) tienen tokens completos de bg+border+text. Usar `running` en el código (el diseño usa `in-progress` pero la base de código ha estandarizado en `running` — son alias semánticos).

## Receta 10 — Seguridad y cadena de suministro

- `npm install` DEBE usar siempre `--ignore-scripts` (según DASH-SEC-002)
- Fijado de versiones: revisar el `package.json` existente — fijar versiones exactas (sin `^`, sin `~`) a menos que el paquete requiera flexibilidad semver
- Nunca agregar dependencias que hagan llamadas a casa, empaqueten analytics, o requieran red en tiempo de ejecución (este es un dashboard de escritorio local)
- **Auditar inmediatamente después de instalar:** después de cualquier `npm install --ignore-scripts`, ejecutar `npm audit`. Si reporta vulnerabilidades moderadas o superiores, corregirlas ANTES de hacer commit. No entregar un scaffold con hallazgos de auditoría sin resolver.
- **Lista roja de Vite:** `vite < 6.4.2` tiene CVEs en el servidor de desarrollo (bypass CORS, bypass `server.fs.deny`, path traversal, lectura arbitraria de archivos via WebSocket). `esbuild <= 0.24.2` (fijado por vite 5.x) tiene un bypass CORS relacionado en el servidor de desarrollo. Vite 5.x NO puede parchearse completamente — actualizar a **vite 6.4.2+** que usa `esbuild ^0.25.0`.
- **Combo React + Vite conocido como seguro** (actualizar cuando existan versiones parcheadas más recientes):
  ```json
  "devDependencies": {
    "@vitejs/plugin-react": "4.3.4",
    "vite": "6.4.2"
  }
  ```
  `@vitejs/plugin-react@4.x` soporta `vite ^4 || ^5 || ^6` — NO saltar a vite 7+ sin actualizar plugin-react a 5.x+, y NO saltar a vite 8+ sin moverse a plugin-react 6.x (que agrega dependencias de rolldown + react-compiler).
- Todos los CVEs conocidos del servidor de desarrollo de vite requieren que el servidor de desarrollo esté corriendo Y que el usuario visite un sitio malicioso. Los builds de producción (`vite build`) y los bundles Wails embebidos NO se ven afectados. Pero siempre corregir de todos modos — el servidor de desarrollo se ejecuta durante el desarrollo.

## Qué reemplaza esta skill

Cuando cargas esta skill, puedes omitir:
- Leer `app.go`, `dtos.go`, `store/read.go` existentes para entender los patrones
- Derivar el layout de build tags buscando en el repositorio
- Derivar la convención de nombres de DTO (camelCase) por comparación
- Buscar los nombres de tokens CSS en `index.css`
- Re-leer `frontend/src/lib/wails.ts` para entender el patrón de guard del binding
- Re-leer vistas existentes para entender los patrones de estado loading/error/empty

Si necesitas el contenido exacto de un archivo existente que esta skill no cubre, leer SOLO ese archivo — no hacer una exploración amplia.
