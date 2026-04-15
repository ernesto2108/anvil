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
  branch?: string
  startedAt: string
  endedAt: string
  durationMs: number
  filesCount: number
  agentsCount: number
  errorReason?: string
  toolUsesCount?: number
  compactionsCount?: number
}

export interface FileDTO {
  path: string
  action: string
  diff?: string
  agentId: string
}

export interface ActivityEventDTO {
  timestamp: string
  agentId: string
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
  startedAt: string
  endedAt: string
  durationMs: number | null
  output: string
}

export interface ToolUsageDTO {
  toolName: string
  count: number
}

export interface ToolUseDetailDTO {
  toolName: string
  command: string
  agentId: string
  timestamp: string
}

export interface TaskDTO {
  id: string
  title: string
  status: string
  createdAt: string
  completedAt: string
}

export interface SessionDetailDTO {
  run: RunDTO
  files: FileDTO[]
  agents: SessionAgentDTO[]
  activityEvents: ActivityEventDTO[]
  toolUsage: ToolUsageDTO[]
  toolDetails: ToolUseDetailDTO[]
  tasks: TaskDTO[]
}

export interface PromptDTO {
  sequence: number
  prompt: string
  timestamp: string
}

export interface TurnStatsDTO {
  turnNumber: number
  prompt: string
  timestamp: string
  endTimestamp: string
  filesCount: number
  toolUsesCount: number
  agentsCount: number
}

// --- Error types ---

export interface ErrorsQuery {
  status: string
  search: string
}

export interface ErrorGroupDTO {
  id: string
  fingerprint: string
  title: string
  normalizedMsg: string
  resolutionStatus: string
  firstSeenAt: string
  lastSeenAt: string
  occurrenceCount: number
  notes: string
  commitLink: string
}

export interface ErrorGroupRunDTO {
  runId: string
  agentName: string
  errorMsg: string
  exitCode: number | null
  occurredAt: string
}

export interface ErrorGroupHistoryDTO {
  oldStatus: string
  newStatus: string
  note: string
  commitLink: string
  createdAt: string
}

export interface TrendPointDTO {
  date: string
  count: number
}

export interface ErrorGroupDetailDTO {
  group: ErrorGroupDTO
  runs: ErrorGroupRunDTO[]
  history: ErrorGroupHistoryDTO[]
  trend: TrendPointDTO[]
}

export interface ErrorsCountDTO {
  new: number
  investigating: number
  weekTotal: number
}

export interface UpdateResolutionRequest {
  groupId: string
  status: string
  notes: string
  commitLink: string
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
          RevertFile?: (project: string, filePath: string, diff: string) => Promise<void>
          GetToolUsage?: (runId: string) => Promise<ToolUsageDTO[]>
          GetTasks?: (runId: string) => Promise<TaskDTO[]>
          GetPrompts?: (runId: string) => Promise<PromptDTO[]>
          GetTurnStats?: (runId: string) => Promise<TurnStatsDTO[]>
          DeleteRun?: (runId: string) => Promise<void>
          DeleteRuns?: (runIds: string[]) => Promise<void>
          GetErrorGroups?: (q: ErrorsQuery) => Promise<ErrorGroupDTO[]>
          GetErrorGroup?: (groupId: string) => Promise<ErrorGroupDetailDTO | null>
          UpdateErrorResolution?: (req: UpdateResolutionRequest) => Promise<void>
          GetErrorsCounts?: () => Promise<ErrorsCountDTO>
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

// revertFile aplica el parche inverso (git apply -R) para revertir un archivo.
export async function revertFile(project: string, filePath: string, diff: string): Promise<void> {
  const binding = window.go?.dashboard?.App?.RevertFile
  if (!binding) {
    throw new Error('Binding RevertFile no disponible (modo dev)')
  }
  return binding(project, filePath, diff)
}

// getToolUsage returns tool usage breakdown for a run.
export async function getToolUsage(runId: string): Promise<ToolUsageDTO[]> {
  const binding = window.go?.dashboard?.App?.GetToolUsage
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando [] (modo dev)')
    return []
  }
  return binding(runId)
}

// getTasks returns all tasks tracked within a run.
export async function getTasks(runId: string): Promise<TaskDTO[]> {
  const binding = window.go?.dashboard?.App?.GetTasks
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando [] (modo dev)')
    return []
  }
  return binding(runId)
}

// getPrompts returns all prompts for a run, ordered by sequence ASC.
export async function getPrompts(runId: string): Promise<PromptDTO[]> {
  const binding = window.go?.dashboard?.App?.GetPrompts
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando [] (modo dev)')
    return []
  }
  return binding(runId)
}

// deleteRun deletes a single run and all associated data.
export async function deleteRun(runId: string): Promise<void> {
  const binding = window.go?.dashboard?.App?.DeleteRun
  if (!binding) {
    throw new Error('Binding DeleteRun no disponible (modo dev)')
  }
  return binding(runId)
}

// deleteRuns deletes multiple runs in a single transaction.
export async function deleteRuns(runIds: string[]): Promise<void> {
  const binding = window.go?.dashboard?.App?.DeleteRuns
  if (!binding) {
    throw new Error('Binding DeleteRuns no disponible (modo dev)')
  }
  return binding(runIds)
}

// getTurnStats returns activity statistics per turn for a run.
export async function getTurnStats(runId: string): Promise<TurnStatsDTO[]> {
  const binding = window.go?.dashboard?.App?.GetTurnStats
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando [] (modo dev)')
    return []
  }
  return binding(runId)
}

// --- Error bindings ---

export async function getErrorGroups(q: ErrorsQuery): Promise<ErrorGroupDTO[]> {
  const binding = window.go?.dashboard?.App?.GetErrorGroups
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando [] (modo dev)')
    return []
  }
  return binding(q)
}

export async function getErrorGroup(groupId: string): Promise<ErrorGroupDetailDTO | null> {
  const binding = window.go?.dashboard?.App?.GetErrorGroup
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando null (modo dev)')
    return null
  }
  return binding(groupId)
}

export async function updateErrorResolution(req: UpdateResolutionRequest): Promise<void> {
  const binding = window.go?.dashboard?.App?.UpdateErrorResolution
  if (!binding) {
    throw new Error('Binding UpdateErrorResolution no disponible (modo dev)')
  }
  return binding(req)
}

export async function getErrorsCounts(): Promise<ErrorsCountDTO> {
  const binding = window.go?.dashboard?.App?.GetErrorsCounts
  if (!binding) {
    console.warn('[wails] window.go no disponible — retornando zeros (modo dev)')
    return { new: 0, investigating: 0, weekTotal: 0 }
  }
  return binding()
}

