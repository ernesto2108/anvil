import { Activity, ChevronRight } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { StatusBadge } from '@/components/status-badge'

import {
  formatDate,
  formatDuration,
  formatTokens,
  formatQAScore,
} from '@/lib/format'
import { cn } from '@/lib/utils'
import { useRunsPolling } from '@/hooks/use-runs-polling'

interface RunsViewProps {
  onRowClick: (runId: string) => void
}

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

// Celda de QA Score con color según tono.
function QACell({ score }: { score: number | null }) {
  const { text, tone } = formatQAScore(score)
  const colorClass = {
    success: 'text-success',
    warn: 'text-running',
    fail: 'text-fail',
    muted: 'text-muted-foreground',
  }[tone]
  return <span className={cn('font-mono text-xs', colorClass)}>{text}</span>
}

// Determina si un run está actualmente en ejecución.
function isRunning(status: string): boolean {
  return status === 'running' || status === 'in_progress' || status === 'in-progress'
}

export function RunsView({ onRowClick }: RunsViewProps) {
  const { runs } = useRunsPolling()

  if (runs === null) return <LoadingSkeleton />
  if (runs.length === 0) return <EmptyState />

  return (
    <div className="p-6">
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
                // Fila activa recibe un tinte sutil del color "running".
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
                <QACell score={run.qaScore} />
              </TableCell>
              <TableCell className="w-[40px] text-right">
                <ChevronRight size={14} className="text-muted-foreground ml-auto" />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
