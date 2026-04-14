import { useEffect, useState, useCallback } from 'react'
import { ArrowLeft, ChevronRight, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { StatusBadge } from '@/components/status-badge'
import { getAgent, type AgentDetailDTO } from '@/lib/wails'
import { formatDate, formatDuration } from '@/lib/format'

interface AgentViewProps {
  runId: string
  agentId: string
  onBack: () => void
}

type LoadState =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'not-found' }
  | { status: 'data'; detail: AgentDetailDTO }

// MetaCard renderiza una tarjeta de metadato con label y valor.
function MetaCard({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card px-4 py-3">
      <p className="text-[11px] font-medium uppercase tracking-wide text-text-muted">{label}</p>
      <div className="mt-1">{children}</div>
    </div>
  )
}

export function AgentView({ runId, agentId, onBack }: AgentViewProps) {
  const [state, setState] = useState<LoadState>({ status: 'loading' })
  const [outputExpanded, setOutputExpanded] = useState(false)
  const [expandedFiles, setExpandedFiles] = useState<Set<number>>(new Set())

  useEffect(() => {
    let cancelled = false
    setState({ status: 'loading' })
    setOutputExpanded(false)
    setExpandedFiles(new Set())

    getAgent(runId, agentId)
      .then((detail) => {
        if (cancelled) return
        if (detail === null) {
          setState({ status: 'not-found' })
          return
        }
        setState({ status: 'data', detail })
      })
      .catch((err: unknown) => {
        if (cancelled) return
        const msg = err instanceof Error ? err.message : String(err)
        setState({ status: 'error', message: msg })
      })

    return () => {
      cancelled = true
    }
  }, [runId, agentId])

  // Handler de teclado para items de archivo (accesibilidad).
  // TODO(DASH-FEAT-futura): implementar acción de "abrir archivo" cuando esté disponible.
  const handleFileKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      // Acción futura — ver TODO arriba.
    }
  }, [])

  const agentName = state.status === 'data' ? state.detail.agent.name : agentId

  return (
    <div className="flex flex-1 min-h-0 flex-col">
      {/* Header */}
      <div className="flex shrink-0 items-center gap-3 border-b border-border px-4 py-3">
        <Button
          variant="ghost"
          size="sm"
          onClick={onBack}
          className="gap-1.5"
          aria-label="Volver al flujo del run"
        >
          <ArrowLeft size={14} />
          Volver
        </Button>
        <span className="text-sm text-muted-foreground">Run</span>
        <code className="font-mono text-xs text-brand-text">{runId}</code>
        <span className="text-sm text-muted-foreground">/</span>
        <span className="text-sm text-foreground">Agente</span>
        <code className="font-mono text-xs text-brand-text">{agentName}</code>
      </div>

      {/* Estados de carga / error / not-found */}
      {state.status === 'loading' && (
        <div className="flex flex-1 items-center justify-center">
          <Loader2 size={20} className="animate-spin text-muted-foreground" />
        </div>
      )}

      {state.status === 'error' && (
        <div className="flex flex-1 flex-col items-center justify-center gap-2 p-6">
          <p className="text-sm text-fail">Error al cargar el agente</p>
          <p className="max-w-sm text-center font-mono text-xs text-muted-foreground">
            {state.message}
          </p>
        </div>
      )}

      {state.status === 'not-found' && (
        <div className="flex flex-1 items-center justify-center p-6">
          <p className="text-sm text-text-muted">Agente no encontrado.</p>
        </div>
      )}

      {state.status === 'data' && (
        <div className="flex-1 overflow-auto p-6 space-y-6">
          {/* Metadatos */}
          <section aria-label="Metadatos del agente">
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
              <MetaCard label="Estado">
                <StatusBadge status={state.detail.agent.status} />
              </MetaCard>

              <MetaCard label="Duración">
                <span className="font-mono text-sm text-foreground">
                  {state.detail.agent.durationMs !== null
                    ? formatDuration(state.detail.agent.durationMs)
                    : '—'}
                </span>
              </MetaCard>

              <MetaCard label="Inicio">
                <span className="text-sm text-foreground">
                  {formatDate(state.detail.agent.startedAt)}
                </span>
              </MetaCard>

              <MetaCard label="Fin">
                <span className="text-sm text-foreground">
                  {formatDate(state.detail.agent.endedAt)}
                </span>
              </MetaCard>
            </div>
          </section>

          {/* Error del agente — solo visible si falló */}
          {state.detail.agent.errorMsg && (
            <section aria-label="Error del agente">
              <h3 className="mb-2 text-sm font-medium text-foreground">Error</h3>
              <div className="rounded-lg border border-fail-border bg-fail-bg p-4">
                <p className="font-mono text-xs text-fail whitespace-pre-wrap break-all">
                  {state.detail.agent.errorMsg}
                </p>
              </div>
            </section>
          )}

          {/* Archivos modificados */}
          <section aria-label="Archivos modificados">
            <h3 className="mb-2 text-sm font-medium text-foreground">Archivos modificados</h3>
            {state.detail.files.length === 0 ? (
              <p className="text-sm text-text-muted">Sin archivos modificados.</p>
            ) : (
              <ul role="list" className="space-y-1">
                {state.detail.files.map((file, idx) => {
                  const hasDiff = !!file.diff
                  const isExpanded = expandedFiles.has(idx)
                  return (
                    <li key={idx}>
                      <button
                        type="button"
                        tabIndex={0}
                        onClick={() => {
                          if (!hasDiff) return
                          setExpandedFiles((prev) => {
                            const next = new Set(prev)
                            if (next.has(idx)) next.delete(idx)
                            else next.add(idx)
                            return next
                          })
                        }}
                        onKeyDown={handleFileKeyDown}
                        className={`flex w-full items-center gap-3 rounded-md border border-border bg-card px-3 py-2 text-left transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring ${hasDiff ? 'cursor-pointer hover:bg-secondary/50' : 'cursor-default'}`}
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

          {/* Output */}
          <section aria-label="Output del agente">
            <h3 className="mb-2 text-sm font-medium text-foreground">Output</h3>
            {state.detail.output === '' ? (
              <p className="text-sm text-text-muted">Sin output registrado.</p>
            ) : (() => {
              const outputLines = state.detail.output.split('\n')
              const isTruncated = outputLines.length > 20
              const displayText = outputExpanded
                ? state.detail.output
                : outputLines.slice(0, 20).join('\n')
              return (
                <>
                  <div className={`rounded-lg border border-border bg-card p-4${outputExpanded ? ' max-h-[600px] overflow-auto' : ' overflow-hidden'}`}>
                    <pre className="whitespace-pre-wrap break-all font-mono text-xs text-foreground">
                      {displayText}
                    </pre>
                  </div>
                  {isTruncated && (
                    <button
                      type="button"
                      onClick={() => setOutputExpanded((prev) => !prev)}
                      className="mt-2 cursor-pointer text-xs text-brand-text hover:underline"
                    >
                      {outputExpanded ? 'Colapsar' : 'Expandir'}
                    </button>
                  )}
                </>
              )
            })()}
          </section>
        </div>
      )}
    </div>
  )
}
