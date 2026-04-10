// Tipos corresponden 1:1 con internal/dashboard/dtos.go
// Campos con *T en Go se mapean a T | null en TypeScript.

export interface RunsQuery {
  limit: number
  offset: number
  status: string
}

export interface RunDTO {
  id: string
  taskId: string
  taskDesc: string
  status: string
  complexity: string
  provider: string
  startedAt: string
  endedAt: string
  durationMs: number
  totalTokens: number
  filesCount: number
  agentsCount: number
  qaScore: number | null
}

export interface FileDTO {
  path: string
  action: string
}

export interface AgentDTO {
  id: string
  name: string
  status: string
  durationMs: number | null
  tokensIn: number | null
  tokensOut: number | null
  tokensTotal: number | null
  startedAt: string
  endedAt: string
  errorMsg: string
}

export interface AgentDetailDTO {
  agent: AgentDTO
  files: FileDTO[]
  output: string // siempre "" hasta DASH-FEAT-013
}

// TokensPerRunPoint es una barra del chart "Tokens per run".
export interface TokensPerRunPoint {
  runId: string
  startedAt: string // RFC3339Nano
  tokens: number
}

// AgentTimePoint es una fila del chart "Avg time per agent".
export interface AgentTimePoint {
  role: string
  avgDurationMs: number
  runsIncluded: number
}

// MetricsDTO contiene los agregados sobre los últimos 30 runs terminados.
export interface MetricsDTO {
  runsCount: number
  hasEnoughData: boolean
  totalTokens: number
  totalTokensDeltaPct: number | null
  avgDurationMs: number
  avgDurationDeltaMs: number | null
  successRate: number
  successRateDelta: number | null
  tokensPerRun: TokensPerRunPoint[]
  avgTimePerAgent: AgentTimePoint[]
}

// FlowNodeData usa `type` (no `interface`) para satisfacer el constraint
// `Record<string, unknown>` que requiere @xyflow/react v12 en Node<T>.
export type FlowNodeData = {
  label: string
  status: string
  durationMs: number | null
  tokens: number | null
}

export interface FlowNode {
  id: string
  type: string
  data: FlowNodeData
}

export interface FlowEdge {
  id: string
  source: string
  target: string
}

export interface FlowDTO {
  nodes: FlowNode[]
  edges: FlowEdge[]
}

// Forma del objeto global window.go inyectado por Wails en producción.
declare global {
  interface Window {
    go?: {
      dashboard?: {
        App?: {
          GetRuns?: (q: RunsQuery) => Promise<RunDTO[]>
          GetFlow?: (runId: string) => Promise<FlowDTO>
          GetAgent?: (runId: string, agentId: string) => Promise<AgentDetailDTO | null>
          GetMetrics?: () => Promise<MetricsDTO>
          GetRunSummary?: (runId: string) => Promise<RunDTO | null>
        }
      }
    }
  }
}

// getRuns llama al binding Wails GetRuns. Cuando window.go no está disponible
// (modo vite dev sin Wails) retorna un arreglo vacío para no bloquear el desarrollo.
export async function getRuns(q: RunsQuery): Promise<RunDTO[]> {
  const binding = window.go?.dashboard?.App?.GetRuns
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando arreglo vacío (modo dev)')
    return []
  }
  return binding(q)
}

// getFlow llama al binding Wails GetFlow para obtener el grafo de un run.
// Cuando window.go no está disponible retorna un FlowDTO vacío para no bloquear el desarrollo.
export async function getFlow(runId: string): Promise<FlowDTO> {
  const binding = window.go?.dashboard?.App?.GetFlow
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando FlowDTO vacío (modo dev)')
    return { nodes: [], edges: [] }
  }
  return binding(runId)
}

// getAgent llama al binding Wails GetAgent para obtener el detalle de un agente.
// Cuando window.go no está disponible retorna null para no bloquear el desarrollo.
export async function getAgent(runId: string, agentId: string): Promise<AgentDetailDTO | null> {
  const binding = window.go?.dashboard?.App?.GetAgent
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando null (modo dev)')
    return null
  }
  return binding(runId, agentId)
}

// getRunSummary llama al binding Wails GetRunSummary para obtener el resumen de un run.
// Cuando window.go no está disponible retorna null para no bloquear el desarrollo.
export async function getRunSummary(runId: string): Promise<RunDTO | null> {
  const binding = window.go?.dashboard?.App?.GetRunSummary
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando null (modo dev)')
    return null
  }
  return binding(runId)
}

// MOCK_METRICS_DATA es el fallback en modo dev (sin Wails).
// Simula 10 runs con datos plausibles — reemplazar con binding real en producción.
// TODO(integration): replace with real API
const MOCK_METRICS_DATA: MetricsDTO = {
  runsCount: 10,
  hasEnoughData: true,
  totalTokens: 1_234_500,
  totalTokensDeltaPct: 0.12,
  avgDurationMs: 134_000,
  avgDurationDeltaMs: -8_500,
  successRate: 0.8,
  successRateDelta: 0.05,
  tokensPerRun: [
    { runId: 'run-001', startedAt: '2026-03-31T10:00:00Z', tokens: 98_000 },
    { runId: 'run-002', startedAt: '2026-04-01T11:00:00Z', tokens: 134_200 },
    { runId: 'run-003', startedAt: '2026-04-01T15:00:00Z', tokens: 87_500 },
    { runId: 'run-004', startedAt: '2026-04-02T09:30:00Z', tokens: 210_000 },
    { runId: 'run-005', startedAt: '2026-04-02T14:00:00Z', tokens: 155_800 },
    { runId: 'run-006', startedAt: '2026-04-03T10:15:00Z', tokens: 92_300 },
    { runId: 'run-007', startedAt: '2026-04-04T09:00:00Z', tokens: 178_400 },
    { runId: 'run-008', startedAt: '2026-04-05T11:30:00Z', tokens: 64_200 },
    { runId: 'run-009', startedAt: '2026-04-07T10:00:00Z', tokens: 145_600 },
    { runId: 'run-010', startedAt: '2026-04-08T14:45:00Z', tokens: 68_500 },
  ],
  avgTimePerAgent: [
    { role: 'developer', avgDurationMs: 82_000, runsIncluded: 10 },
    { role: 'qa', avgDurationMs: 45_000, runsIncluded: 9 },
    { role: 'architect', avgDurationMs: 38_000, runsIncluded: 8 },
    { role: 'tester', avgDurationMs: 25_000, runsIncluded: 7 },
    { role: 'pm', avgDurationMs: 12_000, runsIncluded: 10 },
  ],
}

// getMetrics llama al binding Wails GetMetrics para obtener los agregados de métricas.
// Cuando window.go no está disponible retorna datos mock fijos para desarrollo.
export async function getMetrics(): Promise<MetricsDTO> {
  const binding = window.go?.dashboard?.App?.GetMetrics
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando mock de métricas (modo dev)')
    return Promise.resolve(MOCK_METRICS_DATA)
  }
  return binding()
}
