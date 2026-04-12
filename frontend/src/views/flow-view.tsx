import { useEffect, useState, useCallback } from 'react'
import { ArrowLeft, ChevronRight, Loader2, FileText, Bot } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { StatusBadge } from '@/components/status-badge'
import { formatDate, formatDuration } from '@/lib/format'
import { getSessionDetail, type SessionDetailDTO } from '@/lib/wails'

interface FlowViewProps {
  runId: string
  onBack: () => void
  onAgentSelect?: (agentId: string) => void
}

type LoadState =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'empty' }
  | { status: 'data'; detail: SessionDetailDTO }

export function FlowView({ runId, onBack }: FlowViewProps) {
  const [state, setState] = useState<LoadState>({ status: 'loading' })
  const [expandedFiles, setExpandedFiles] = useState<Set<number>>(new Set())
  const [expandedAgents, setExpandedAgents] = useState<Set<number>>(new Set())

  useEffect(() => {
    let cancelled = false
    setState({ status: 'loading' })
    setExpandedFiles(new Set())
    setExpandedAgents(new Set())

    getSessionDetail(runId)
      .then((detail) => {
        if (cancelled) return
        if (!detail) {
          setState({ status: 'empty' })
          return
        }
        setState({ status: 'data', detail })
      })
      .catch((err: unknown) => {
        if (cancelled) return
        const msg = err instanceof Error ? err.message : String(err)
        setState({ status: 'error', message: msg })
      })

    return () => { cancelled = true }
  }, [runId])

  const toggleFile = useCallback((idx: number) => {
    setExpandedFiles((prev) => {
      const next = new Set(prev)
      if (next.has(idx)) next.delete(idx)
      else next.add(idx)
      return next
    })
  }, [])

  const toggleAgent = useCallback((idx: number) => {
    setExpandedAgents((prev) => {
      const next = new Set(prev)
      if (next.has(idx)) next.delete(idx)
      else next.add(idx)
      return next
    })
  }, [])

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

      {/* Content */}
      <div className="flex-1 overflow-auto">
        {state.status === 'loading' && (
          <div className="flex items-center justify-center h-full">
            <Loader2 size={20} className="animate-spin text-muted-foreground" />
          </div>
        )}

        {state.status === 'error' && (
          <div className="flex flex-col items-center justify-center h-full gap-2 p-6">
            <p className="text-sm text-fail">Error al cargar la sesión</p>
            <p className="max-w-sm text-center font-mono text-xs text-muted-foreground">
              {state.message}
            </p>
          </div>
        )}

        {state.status === 'empty' && (
          <div className="flex items-center justify-center h-full">
            <p className="text-sm text-muted-foreground">Sesión no encontrada.</p>
          </div>
        )}

        {state.status === 'data' && (
          <div className="p-6 space-y-6 max-w-4xl">
            {/* Run summary */}
            <section>
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                <div className="rounded-lg border border-border bg-card px-4 py-3">
                  <p className="text-[11px] font-medium uppercase tracking-wide text-text-muted">Estado</p>
                  <div className="mt-1"><StatusBadge status={state.detail.run.status} /></div>
                </div>
                <div className="rounded-lg border border-border bg-card px-4 py-3">
                  <p className="text-[11px] font-medium uppercase tracking-wide text-text-muted">Duración</p>
                  <p className="mt-1 font-mono text-sm text-foreground">
                    {state.detail.run.durationMs > 0 ? formatDuration(state.detail.run.durationMs) : '—'}
                  </p>
                </div>
                <div className="rounded-lg border border-border bg-card px-4 py-3">
                  <p className="text-[11px] font-medium uppercase tracking-wide text-text-muted">Inicio</p>
                  <p className="mt-1 text-sm text-foreground">{formatDate(state.detail.run.startedAt)}</p>
                </div>
                <div className="rounded-lg border border-border bg-card px-4 py-3">
                  <p className="text-[11px] font-medium uppercase tracking-wide text-text-muted">Archivos</p>
                  <p className="mt-1 font-mono text-sm text-foreground">{state.detail.files.length}</p>
                </div>
              </div>

              {/* Task description */}
              {state.detail.run.taskDesc && (
                <div className="mt-3 rounded-lg border border-border bg-card px-4 py-3">
                  <p className="text-[11px] font-medium uppercase tracking-wide text-text-muted">Prompt</p>
                  <p className="mt-1 text-sm text-foreground whitespace-pre-wrap">{state.detail.run.taskDesc}</p>
                </div>
              )}
            </section>

            {/* Files changed */}
            <section>
              <h3 className="flex items-center gap-2 mb-3 text-sm font-medium text-foreground">
                <FileText size={14} className="text-muted-foreground" />
                Archivos modificados ({state.detail.files.length})
              </h3>
              {state.detail.files.length === 0 ? (
                <p className="text-sm text-text-muted">Sin archivos modificados en esta sesión.</p>
              ) : (
                <ul className="space-y-1">
                  {state.detail.files.map((file, idx) => {
                    const hasDiff = !!file.diff
                    const isExpanded = expandedFiles.has(idx)
                    return (
                      <li key={idx}>
                        <button
                          type="button"
                          onClick={() => hasDiff && toggleFile(idx)}
                          className={`flex w-full items-center gap-3 rounded-md border border-border bg-card px-3 py-2 text-left transition-colors ${hasDiff ? 'cursor-pointer hover:bg-secondary/50' : 'cursor-default'}`}
                        >
                          {hasDiff && (
                            <ChevronRight
                              size={14}
                              className={`shrink-0 text-text-muted transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                            />
                          )}
                          <span className="flex-1 truncate font-mono text-xs text-text-secondary">
                            {file.path}
                          </span>
                          <span className="shrink-0 rounded-full border border-border px-2 py-0.5 text-[11px] font-medium text-text-muted">
                            {file.action}
                          </span>
                        </button>
                        {hasDiff && isExpanded && (
                          <div className="mt-1 ml-5 rounded-md border border-border bg-[#1a1a2e] p-3 max-h-[400px] overflow-auto">
                            <pre className="whitespace-pre-wrap break-all font-mono text-xs leading-relaxed">
                              {file.diff!.split('\n').map((line, i) => {
                                let cls = 'text-text-muted'
                                if (line.startsWith('+')) cls = 'text-success'
                                else if (line.startsWith('-')) cls = 'text-fail'
                                else if (line.startsWith('@@')) cls = 'text-brand-text'
                                return (
                                  <span key={i} className={cls}>
                                    {line}
                                    {'\n'}
                                  </span>
                                )
                              })}
                            </pre>
                          </div>
                        )}
                      </li>
                    )
                  })}
                </ul>
              )}
            </section>

            {/* Agents */}
            {state.detail.agents.length > 0 && (
              <section>
                <h3 className="flex items-center gap-2 mb-3 text-sm font-medium text-foreground">
                  <Bot size={14} className="text-muted-foreground" />
                  Agentes ({state.detail.agents.length})
                </h3>
                <ul className="space-y-1">
                  {state.detail.agents.map((agent, idx) => {
                    const hasOutput = !!agent.output
                    const isExpanded = expandedAgents.has(idx)
                    return (
                      <li key={idx}>
                        <button
                          type="button"
                          onClick={() => hasOutput && toggleAgent(idx)}
                          className={`flex w-full items-center gap-3 rounded-md border border-border bg-card px-3 py-2 text-left transition-colors ${hasOutput ? 'cursor-pointer hover:bg-secondary/50' : 'cursor-default'}`}
                        >
                          {hasOutput && (
                            <ChevronRight
                              size={14}
                              className={`shrink-0 text-text-muted transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                            />
                          )}
                          <span className="font-mono text-xs text-foreground">{agent.role}</span>
                          <StatusBadge status={agent.status} />
                          {agent.durationMs != null && (
                            <span className="font-mono text-[11px] text-muted-foreground">
                              {formatDuration(agent.durationMs)}
                            </span>
                          )}
                        </button>
                        {hasOutput && isExpanded && (
                          <div className="mt-1 ml-5 rounded-md border border-border bg-card p-4 max-h-[500px] overflow-auto">
                            <pre className="whitespace-pre-wrap break-all font-mono text-xs text-foreground">
                              {agent.output}
                            </pre>
                          </div>
                        )}
                      </li>
                    )
                  })}
                </ul>
              </section>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
