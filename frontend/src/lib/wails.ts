// Tipos corresponden 1:1 con internal/dashboard/dtos.go
// Campos con *T en Go se mapean a T | null en TypeScript.

export interface RunsQuery {
  limit: number
  offset: number
  status: string
  startDate: string
  endDate: string
  project: string
}

export interface RunDTO {
  id: string
  taskId: string
  taskDesc: string
  status: string
  complexity: string
  provider: string
  project: string
  parentRunId?: string
  childrenCount?: number
  startedAt: string
  endedAt: string
  durationMs: number
  filesCount: number
  agentsCount: number
}

export interface FileDTO {
  path: string
  action: string
  diff?: string
}

export interface AgentDTO {
  id: string
  name: string
  status: string
  durationMs: number | null
  startedAt: string
  endedAt: string
  errorMsg: string
}

export interface AgentDetailDTO {
  agent: AgentDTO
  files: FileDTO[]
  output: string
}

export interface SessionAgentDTO {
  id: string
  role: string
  status: string
  durationMs: number | null
  output: string
}

export interface SessionDetailDTO {
  run: RunDTO
  files: FileDTO[]
  agents: SessionAgentDTO[]
}

// Forma del objeto global window.go inyectado por Wails en producción.
declare global {
  interface Window {
    go?: {
      dashboard?: {
        App?: {
          GetRuns?: (q: RunsQuery) => Promise<RunDTO[]>
          GetProjects?: () => Promise<string[]>
          GetAgent?: (runId: string, agentId: string) => Promise<AgentDetailDTO | null>
          GetRunSummary?: (runId: string) => Promise<RunDTO | null>
          GetSessionDetail?: (runId: string) => Promise<SessionDetailDTO | null>
          GetChildRuns?: (parentRunId: string) => Promise<RunDTO[]>
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

// getProjects returns distinct project names for the filter dropdown.
export async function getProjects(): Promise<string[]> {
  const binding = window.go?.dashboard?.App?.GetProjects
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando [] (modo dev)')
    return []
  }
  return binding()
}

// getSessionDetail returns run + files + agents for the session detail view.
export async function getSessionDetail(runId: string): Promise<SessionDetailDTO | null> {
  const binding = window.go?.dashboard?.App?.GetSessionDetail
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando null (modo dev)')
    return null
  }
  return binding(runId)
}

// getChildRuns llama al binding Wails GetChildRuns para obtener los runs hijos de un parent run.
// Cuando window.go no está disponible retorna [] para no bloquear el desarrollo.
export async function getChildRuns(parentRunId: string): Promise<RunDTO[]> {
  const binding = window.go?.dashboard?.App?.GetChildRuns
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando [] (modo dev)')
    return []
  }
  return binding(parentRunId)
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

