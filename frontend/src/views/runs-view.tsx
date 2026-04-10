import { useState } from 'react'
import { Activity, ChevronRight, X } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { StatusBadge } from '@/components/status-badge'
import { QABadge } from '@/components/qa-badge'

import {
  formatDate,
  formatDuration,
  formatTokens,
} from '@/lib/format'
import { cn } from '@/lib/utils'
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

// Skeleton de carga: tres filas de barras animadas.
function LoadingSkeleton() {
  return (
    <div className="p-6 space-y-3">
      {[0, 1, 2].map((i) => (
        <div key={i} className="h-10 rounded-md bg-secondary/50 animate-pulse" />
      ))}
    </div>
  )
}

// Estado vacío cuando no hay runs registrados.
function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center h-full min-h-[400px] gap-4 px-6">
      <Activity size={48} className="text-muted-foreground" />
      <div className="text-center space-y-1">
        <p className="text-sm font-medium text-foreground">Sin ejecuciones todavía</p>
        <p className="text-sm text-muted-foreground">
          Anvil está listo. Ejecuta tu primer workflow
        </p>
        <p className="text-sm text-muted-foreground">
          desde la terminal para ver los resultados aquí.
        </p>
      </div>
      <div className="bg-background border border-border rounded-md px-4 py-2.5 font-mono text-sm text-muted-foreground">
        $ anvil orchestrate
      </div>
    </div>
  )
}

// Determina si un run está actualmente en ejecución.
function isRunning(status: string): boolean {
  return status === 'running' || status === 'in_progress' || status === 'in-progress'
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
    <div className="flex items-center gap-3 px-6 pt-6 pb-2">
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
          onClick={() => onChange({ status: '', startDate: '', endDate: '' })}
          className="inline-flex items-center gap-1 h-8 rounded-md px-2 text-xs text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors"
        >
          <X size={12} />
          Limpiar
        </button>
      )}
    </div>
  )
}

export function RunsView({ onRowClick }: RunsViewProps) {
  const [filter, setFilter] = useState<RunsFilter>({
    status: '',
    startDate: '',
    endDate: '',
  })

  const { runs } = useRunsPolling(filter)

  if (runs === null) return <LoadingSkeleton />

  return (
    <div>
      <FilterBar filter={filter} onChange={setFilter} />

      {runs.length === 0 ? (
        filter.status !== '' || filter.startDate !== '' || filter.endDate !== '' ? (
          <div className="flex flex-col items-center justify-center min-h-[300px] gap-3 px-6">
            <Activity size={36} className="text-muted-foreground" />
            <p className="text-sm text-muted-foreground">
              Sin resultados para los filtros seleccionados
            </p>
            <button
              type="button"
              onClick={() => setFilter({ status: '', startDate: '', endDate: '' })}
              className="text-xs text-muted-foreground hover:text-foreground underline"
            >
              Limpiar filtros
            </button>
          </div>
        ) : (
          <EmptyState />
        )
      ) : (
        <div className="px-6 pb-6">
          <Table>
            <TableHeader>
              <TableRow className="border-border hover:bg-transparent">
                <TableHead className="w-[140px]">ID</TableHead>
                <TableHead className="w-[160px]">Fecha</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
                <TableHead className="w-[100px]">Duración</TableHead>
                <TableHead className="w-[110px] text-right">Tokens</TableHead>
                <TableHead className="w-[80px] text-right">QA</TableHead>
                <TableHead className="w-[40px]" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {runs.map((run) => (
                <TableRow
                  key={run.id}
                  className={cn(
                    'border-border cursor-pointer hover:bg-secondary/50 transition-colors',
                    isRunning(run.status) && 'bg-running-bg/20',
                  )}
                  onClick={() => onRowClick(run.id)}
                >
                  <TableCell className="font-mono text-xs text-text-secondary w-[140px] truncate max-w-[140px]">
                    {run.id}
                  </TableCell>
                  <TableCell className="text-xs text-foreground w-[160px]">
                    {formatDate(run.startedAt)}
                  </TableCell>
                  <TableCell className="w-[110px]">
                    <StatusBadge status={run.status} />
                  </TableCell>
                  <TableCell className="font-mono text-xs text-foreground w-[100px]">
                    {isRunning(run.status) ? '—' : formatDuration(run.durationMs)}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-foreground w-[110px] text-right">
                    {formatTokens(run.totalTokens)}
                  </TableCell>
                  <TableCell className="w-[80px] text-right">
                    <QABadge score={run.qaScore} />
                  </TableCell>
                  <TableCell className="w-[40px] text-right">
                    <ChevronRight size={14} className="text-muted-foreground ml-auto" />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}
