import { useEffect, useState, useCallback, useMemo } from 'react'
import {
  ReactFlow,
  Background,
  BackgroundVariant,
  type Edge,
  type NodeTypes,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { ArrowLeft, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { AgentNode, type AgentNodeType } from '@/components/flow/agent-node'
import { layoutFlow } from '@/lib/dagre-layout'
import { getFlow } from '@/lib/wails'

interface FlowViewProps {
  runId: string
  onBack: () => void
  onAgentSelect?: (agentId: string) => void
}

type LoadState =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'empty' }
  | { status: 'data'; nodes: AgentNodeType[]; edges: Edge[] }

// nodeTypes definido fuera del componente para evitar re-creación en cada render.
const nodeTypes: NodeTypes = {
  agentNode: AgentNode,
}

export function FlowView({ runId, onBack, onAgentSelect }: FlowViewProps) {
  const [state, setState] = useState<LoadState>({ status: 'loading' })

  useEffect(() => {
    let cancelled = false
    setState({ status: 'loading' })

    getFlow(runId)
      .then((dto) => {
        if (cancelled) return

        if (dto.nodes.length === 0) {
          setState({ status: 'empty' })
          return
        }

        const rawNodes: AgentNodeType[] = dto.nodes.map((n) => ({
          id: n.id,
          type: 'agentNode' as const,
          data: n.data,
          position: { x: 0, y: 0 }, // dagre calculará la posición real
        }))

        const rawEdges: Edge[] = dto.edges.map((e) => ({
          id: e.id,
          source: e.source,
          target: e.target,
          type: 'smoothstep',
        }))

        // layoutFlow retorna Node<T>[] con type: string | undefined;
        // como los nodos son todos 'agentNode' const, el cast es seguro.
        const positioned = layoutFlow(rawNodes, rawEdges) as AgentNodeType[]
        setState({ status: 'data', nodes: positioned, edges: rawEdges })
      })
      .catch((err: unknown) => {
        if (cancelled) return
        const msg = err instanceof Error ? err.message : String(err)
        setState({ status: 'error', message: msg })
      })

    return () => {
      cancelled = true
    }
  }, [runId])

  const handleNodeClick = useCallback(
    (_: React.MouseEvent, node: AgentNodeType) => {
      onAgentSelect?.(node.id)
    },
    [onAgentSelect],
  )

  // Nodos y aristas memoizados para evitar re-renders del canvas en cada ciclo.
  const { nodes, edges } = useMemo(() => {
    if (state.status === 'data') return { nodes: state.nodes, edges: state.edges }
    return { nodes: [] as AgentNodeType[], edges: [] as Edge[] }
  }, [state])

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="flex shrink-0 items-center gap-3 border-b border-border px-4 py-3">
        <Button variant="ghost" size="sm" onClick={onBack} className="gap-1.5">
          <ArrowLeft size={14} />
          Volver
        </Button>
        <span className="text-sm text-muted-foreground">Run</span>
        <code className="font-mono text-xs text-brand-text">{runId}</code>
      </div>

      {/* Canvas area */}
      <div className="relative flex-1">
        {state.status === 'loading' && (
          <div className="absolute inset-0 flex items-center justify-center">
            <Loader2 size={20} className="animate-spin text-muted-foreground" />
          </div>
        )}

        {state.status === 'error' && (
          <div className="absolute inset-0 flex flex-col items-center justify-center gap-2 p-6">
            <p className="text-sm text-fail">Error al cargar el grafo</p>
            <p className="max-w-sm text-center font-mono text-xs text-muted-foreground">
              {state.message}
            </p>
          </div>
        )}

        {state.status === 'empty' && (
          <div className="absolute inset-0 flex items-center justify-center">
            <p className="text-sm text-muted-foreground">Sin agentes registrados para este run.</p>
          </div>
        )}

        {state.status === 'data' && (
          <ReactFlow
            nodes={nodes}
            edges={edges}
            nodeTypes={nodeTypes}
            onNodeClick={handleNodeClick}
            fitView
            fitViewOptions={{ padding: 0.2 }}
            proOptions={{ hideAttribution: true }}
            className="bg-background"
          >
            <Background variant={BackgroundVariant.Dots} gap={24} size={1} className="opacity-20" />
          </ReactFlow>
        )}
      </div>
    </div>
  )
}
