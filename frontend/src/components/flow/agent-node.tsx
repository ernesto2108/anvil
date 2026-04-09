import { Handle, Position } from '@xyflow/react'
import type { Node, NodeProps } from '@xyflow/react'
import { Bot } from 'lucide-react'
import { StatusBadge } from '@/components/status-badge'
import { formatDuration, formatTokensFlow } from '@/lib/format'
import type { FlowNodeData } from '@/lib/wails'
import { cn } from '@/lib/utils'

// AgentNodeType es el tipo de nodo para el grafo de agentes.
// Se define aquí para usarlo como tipo exportado en el registro de nodeTypes.
export type AgentNodeType = Node<FlowNodeData, 'agentNode'>

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

export function AgentNode({ data, selected }: NodeProps<AgentNodeType>) {
  const duration = data.durationMs != null ? formatDuration(data.durationMs) : '—'
  const tokens = data.tokens != null ? formatTokensFlow(data.tokens) : null

  const durationLabel = data.durationMs != null ? formatDuration(data.durationMs) : 'pendiente'

  return (
    <>
      {/* Handle invisible — izquierda (target) */}
      <Handle
        type="target"
        position={Position.Left}
        className="!opacity-0 !pointer-events-none"
      />

      <div
        role="button"
        tabIndex={0}
        aria-label={`Agente ${data.label}, estado ${data.status}, duración ${durationLabel}`}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            e.currentTarget.click()
          }
        }}
        className={cn(
          'flex flex-col gap-1.5 rounded-lg border bg-card px-3 py-2.5',
          'cursor-pointer hover:bg-secondary/80 transition-colors',
          'w-[220px] min-h-[72px]',
          borderColorForStatus(data.status),
          selected && 'ring-1 ring-ring',
        )}
      >
        {/* Cabecera: icono + role */}
        <div className="flex items-center gap-1.5">
          <Bot size={13} className="shrink-0 text-muted-foreground" aria-hidden="true" />
          <span className="truncate text-[13px] font-medium leading-none">{data.label}</span>
          <div className="ml-auto shrink-0">
            <StatusBadge status={data.status} />
          </div>
        </div>

        {/* Meta: duración y tokens */}
        <div className="flex items-center gap-2 font-mono text-[11px] text-muted-foreground">
          <span>{duration}</span>
          {tokens != null && <span>·</span>}
          {tokens != null && <span>{tokens}</span>}
        </div>
      </div>

      {/* Handle invisible — derecha (source) */}
      <Handle
        type="source"
        position={Position.Right}
        className="!opacity-0 !pointer-events-none"
      />
    </>
  )
}
