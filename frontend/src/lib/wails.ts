// Tipos corresponden 1:1 con internal/dashboard/dtos.go

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
