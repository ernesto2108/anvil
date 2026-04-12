import { useEffect, useMemo, useState } from 'react'
import { Activity, ChevronRight, FolderOpen, Folder, GitBranch, X } from 'lucide-react'
import { StatusBadge } from '@/components/status-badge'
import { formatDate, formatDuration } from '@/lib/format'
import { cn } from '@/lib/utils'
import { getChildRuns, getProjects, type RunDTO } from '@/lib/wails'
import { useRunsPolling, type RunsFilter } from '@/hooks/use-runs-polling'

interface RunsViewProps {
  onRowClick: (runId: string) => void
}

const STATUS_OPTIONS = [
  { value: '', label: 'Todos' },
  { value: 'running', label: 'Running' },
  { value: 'success', label: 'Success' },
  { value: 'failed', label: 'Failed' },
] as const

function LoadingSkeleton() {
  return (
    <div className="p-6 space-y-3">
      {[0, 1, 2].map((i) => (
        <div key={i} className="h-10 rounded-md bg-secondary/50 animate-pulse" />
      ))}
    </div>
  )
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center h-full min-h-[400px] gap-4 px-6">
      <Activity size={48} className="text-muted-foreground" />
      <div className="text-center space-y-1">
        <p className="text-sm font-medium text-foreground">Sin ejecuciones todavía</p>
        <p className="text-sm text-muted-foreground">
          Anvil está listo. Ejecuta tu primer workflow
          desde la terminal para ver los resultados aquí.
        </p>
      </div>
      <div className="bg-background border border-border rounded-md px-4 py-2.5 font-mono text-sm text-muted-foreground">
        $ anvil orchestrate
      </div>
    </div>
  )
}

function isRunning(status: string): boolean {
  return status === 'running' || status === 'in_progress' || status === 'in-progress'
}

// Group runs by project name
function groupByProject(runs: RunDTO[]): Map<string, RunDTO[]> {
  const groups = new Map<string, RunDTO[]>()
  for (const run of runs) {
    const key = run.project || '(sin proyecto)'
    const arr = groups.get(key) ?? []
    arr.push(run)
    groups.set(key, arr)
  }
  return groups
}

function FilterBar({
  filter,
  onChange,
}: {
  filter: RunsFilter
  onChange: (f: RunsFilter) => void
}) {
  const hasFilters = filter.status !== '' || filter.startDate !== '' || filter.endDate !== ''

  return (
    <div className="flex items-center gap-3 px-6 pt-6 pb-2 flex-wrap">
      <select
        value={filter.status}
        onChange={(e) => onChange({ ...filter, status: e.target.value })}
        className="h-8 rounded-md border border-border bg-background px-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
      >
        {STATUS_OPTIONS.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>

      <div className="flex items-center gap-1.5">
        <label className="text-xs text-muted-foreground">Desde</label>
        <input
          type="date"
          value={filter.startDate}
          onChange={(e) => onChange({ ...filter, startDate: e.target.value })}
          className="h-8 rounded-md border border-border bg-background px-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
        />
      </div>

      <div className="flex items-center gap-1.5">
        <label className="text-xs text-muted-foreground">Hasta</label>
        <input
          type="date"
          value={filter.endDate}
          onChange={(e) => onChange({ ...filter, endDate: e.target.value })}
          className="h-8 rounded-md border border-border bg-background px-2 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring"
        />
      </div>

      {hasFilters && (
        <button
          type="button"
          onClick={() => onChange({ status: '', startDate: '', endDate: '', project: '' })}
          className="inline-flex items-center gap-1 h-8 rounded-md px-2 text-xs text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors"
        >
          <X size={12} />
          Limpiar
        </button>
      )}
    </div>
  )
}

function RunRow({ run, onRowClick }: { run: RunDTO; onRowClick: (id: string) => void }) {
  return (
    <button
      type="button"
      onClick={() => onRowClick(run.id)}
      className={cn(
        'flex w-full items-center gap-3 rounded-md px-3 py-2 text-left transition-colors',
        'hover:bg-secondary/50 cursor-pointer',
        isRunning(run.status) && 'bg-running-bg/20',
      )}
    >
      <span className="font-mono text-xs text-text-secondary w-[160px] truncate shrink-0">
        {run.id}
      </span>
      <span className="text-xs text-foreground w-[140px] shrink-0">
        {formatDate(run.startedAt)}
      </span>
      <span className="w-[100px] shrink-0">
        <StatusBadge status={run.status} />
      </span>
      <span className="font-mono text-xs text-foreground w-[80px] shrink-0">
        {isRunning(run.status) ? '—' : formatDuration(run.durationMs)}
      </span>
      <span className="flex-1 text-xs text-text-muted truncate">
        {run.taskDesc ? run.taskDesc.slice(0, 60) : ''}
      </span>
      <ChevronRight size={14} className="text-muted-foreground shrink-0" />
    </button>
  )
}

function ParentRunRow({
  run,
  onRowClick,
}: {
  run: RunDTO
  onRowClick: (id: string) => void
}) {
  const [expanded, setExpanded] = useState(false)
  const [children, setChildren] = useState<RunDTO[] | null>(null)
  const [loading, setLoading] = useState(false)

  const handleToggle = async () => {
    if (!expanded && children === null) {
      setLoading(true)
      const result = await getChildRuns(run.id)
      setChildren(result)
      setLoading(false)
    }
    setExpanded(!expanded)
  }

  // Group children by project
  const childGroups = useMemo(() => {
    if (!children) return null
    return groupByProject(children)
  }, [children])

  return (
    <div>
      {/* Header row — click to expand, NOT navigate */}
      <button
        type="button"
        onClick={handleToggle}
        className={cn(
          'flex w-full items-center gap-3 rounded-md px-3 py-2 text-left transition-colors',
          'hover:bg-secondary/50 cursor-pointer',
          expanded && 'bg-secondary/30',
        )}
      >
        <GitBranch size={14} className="text-brand shrink-0" />
        <span className="text-xs text-foreground flex-1 truncate">
          {run.taskDesc ? run.taskDesc.slice(0, 60) : 'Cross-service run'}
        </span>
        <span className="text-[11px] text-muted-foreground font-mono">
          {run.childrenCount ?? 0} {run.childrenCount === 1 ? 'servicio' : 'servicios'}
        </span>
        <StatusBadge status={run.status} />
        <span className="font-mono text-xs text-foreground w-[80px] shrink-0">
          {isRunning(run.status) ? '—' : formatDuration(run.durationMs)}
        </span>
        <ChevronRight
          size={14}
          className={cn(
            'text-muted-foreground transition-transform shrink-0',
            expanded && 'rotate-90',
          )}
        />
      </button>

      {/* Expanded children */}
      {expanded && (
        <div className="ml-6 mt-1 space-y-2 border-l-2 border-border pl-3 pb-2">
          {loading && (
            <div className="py-2">
              <div className="h-8 rounded-md bg-secondary/50 animate-pulse" />
            </div>
          )}
          {childGroups && Array.from(childGroups.entries()).map(([project, projectRuns]) => (
            <div key={project}>
              <div className="text-[11px] text-muted-foreground font-medium px-1 py-1">
                {project}
              </div>
              {projectRuns.map((child) => (
                <RunRow key={child.id} run={child} onRowClick={onRowClick} />
              ))}
            </div>
          ))}
          {children && children.length === 0 && (
            <span className="text-xs text-muted-foreground px-1">Sin sub-runs</span>
          )}
        </div>
      )}
    </div>
  )
}

function ProjectGroup({
  name,
  runs,
  expanded,
  onToggle,
  onRowClick,
}: {
  name: string
  runs: RunDTO[]
  expanded: boolean
  onToggle: () => void
  onRowClick: (id: string) => void
}) {
  const successCount = runs.filter((r) => r.status === 'success').length
  const failedCount = runs.filter((r) => r.status === 'failed').length

  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <button
        type="button"
        onClick={onToggle}
        className="flex w-full items-center gap-2.5 px-4 py-3 text-left bg-card hover:bg-secondary/30 transition-colors"
      >
        {expanded
          ? <FolderOpen size={16} className="text-brand shrink-0" />
          : <Folder size={16} className="text-muted-foreground shrink-0" />
        }
        <span className="text-sm font-medium text-foreground">{name}</span>
        <span className="text-xs text-muted-foreground">
          {runs.length} {runs.length === 1 ? 'sesión' : 'sesiones'}
        </span>
        <div className="flex items-center gap-2 ml-auto">
          {successCount > 0 && (
            <span className="text-[11px] text-success font-mono">{successCount} ok</span>
          )}
          {failedCount > 0 && (
            <span className="text-[11px] text-fail font-mono">{failedCount} fail</span>
          )}
          <ChevronRight
            size={14}
            className={cn(
              'text-muted-foreground transition-transform',
              expanded && 'rotate-90',
            )}
          />
        </div>
      </button>

      {expanded && (
        <div className="border-t border-border bg-background px-2 py-1 space-y-0.5">
          {runs.map((run) =>
            (run.childrenCount ?? 0) > 0 ? (
              <ParentRunRow key={run.id} run={run} onRowClick={onRowClick} />
            ) : (
              <RunRow key={run.id} run={run} onRowClick={onRowClick} />
            )
          )}
        </div>
      )}
    </div>
  )
}

export function RunsView({ onRowClick }: RunsViewProps) {
  const [filter, setFilter] = useState<RunsFilter>({
    status: '',
    startDate: '',
    endDate: '',
    project: '',
  })
  const [expandedProjects, setExpandedProjects] = useState<Set<string>>(new Set())
  const [didInitialExpand, setDidInitialExpand] = useState(false)

  // Load projects list for auto-expanding the first time
  const [, setProjects] = useState<string[]>([])
  useEffect(() => {
    getProjects().then(setProjects).catch(() => {})
  }, [])

  const { runs } = useRunsPolling(filter)

  const grouped = useMemo(() => {
    if (!runs) return null
    return groupByProject(runs)
  }, [runs])

  // Auto-expand all projects only on first load
  useEffect(() => {
    if (grouped && !didInitialExpand) {
      setExpandedProjects(new Set(grouped.keys()))
      setDidInitialExpand(true)
    }
  }, [grouped, didInitialExpand])

  if (runs === null) return <LoadingSkeleton />

  const hasActiveFilters = filter.status !== '' || filter.startDate !== '' || filter.endDate !== ''

  const toggleProject = (name: string) => {
    setExpandedProjects((prev) => {
      const next = new Set(prev)
      if (next.has(name)) next.delete(name)
      else next.add(name)
      return next
    })
  }

  return (
    <div>
      <FilterBar filter={filter} onChange={setFilter} />

      {runs.length === 0 ? (
        hasActiveFilters ? (
          <div className="flex flex-col items-center justify-center min-h-[300px] gap-3 px-6">
            <Activity size={36} className="text-muted-foreground" />
            <p className="text-sm text-muted-foreground">
              Sin resultados para los filtros seleccionados
            </p>
            <button
              type="button"
              onClick={() => setFilter({ status: '', startDate: '', endDate: '', project: '' })}
              className="text-xs text-muted-foreground hover:text-foreground underline"
            >
              Limpiar filtros
            </button>
          </div>
        ) : (
          <EmptyState />
        )
      ) : (
        <div className="px-6 pb-6 space-y-3">
          {grouped && Array.from(grouped.entries()).map(([project, projectRuns]) => (
            <ProjectGroup
              key={project}
              name={project}
              runs={projectRuns}
              expanded={expandedProjects.has(project)}
              onToggle={() => toggleProject(project)}
              onRowClick={onRowClick}
            />
          ))}
        </div>
      )}
    </div>
  )
}
