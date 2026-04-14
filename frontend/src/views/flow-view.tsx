import { useEffect, useRef, useState, useCallback, useMemo } from 'react'
import { ArrowLeft, ChevronRight, Loader2, FileText, Bot, Activity, Undo2, GitBranch, ListChecks, AlertTriangle, Minimize2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { StatusBadge } from '@/components/status-badge'
import { DiffViewer } from '@/components/diff-viewer'
import { PromptTimeline } from '@/components/prompt-timeline'
import { TurnGroup } from '@/components/turn-group'
import { formatDate, formatDuration } from '@/lib/format'
import { getSessionDetail, getPrompts, getTurnStats, revertFile, type SessionDetailDTO, type SessionAgentDTO, type ActivityEventDTO, type ToolUseDetailDTO, type TaskDTO, type PromptDTO, type TurnStatsDTO } from '@/lib/wails'

interface FlowViewProps {
  runId: string
  onBack: () => void
  onAgentSelect?: (agentId: string) => void
}

type LoadState =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'empty' }
  | { status: 'data'; detail: SessionDetailDTO; prompts: PromptDTO[]; turnStats: TurnStatsDTO[] }

type TabId = 'traza' | 'cambios' | 'comandos' | 'agentes'

const POLL_INTERVAL_MS = 2000

function isRunFinished(status: string): boolean {
  return status === 'success' || status === 'failed' || status === 'error' || status === 'abandoned'
}

export function FlowView({ runId, onBack }: FlowViewProps) {
  const [state, setState] = useState<LoadState>({ status: 'loading' })
  const [activeTab, setActiveTab] = useState<TabId>('traza')
  const [expandedFiles, setExpandedFiles] = useState<Set<number>>(new Set())
  const [expandedAgents, setExpandedAgents] = useState<Set<number>>(new Set())
  const mountedRef = useRef(true)
  const initialFetchDoneRef = useRef(false)

  useEffect(() => {
    mountedRef.current = true
    initialFetchDoneRef.current = false
    setState({ status: 'loading' })
    setActiveTab('traza')
    setExpandedFiles(new Set())
    setExpandedAgents(new Set())

    let intervalId: ReturnType<typeof setInterval> | null = null

    async function fetchDetail() {
      try {
        const [detail, prompts, turnStats] = await Promise.all([
          getSessionDetail(runId),
          getPrompts(runId),
          getTurnStats(runId),
        ])
        if (!mountedRef.current) return
        if (!detail) { setState({ status: 'empty' }); return }
        setState({ status: 'data', detail, prompts: prompts ?? [], turnStats: turnStats ?? [] })
        initialFetchDoneRef.current = true
        if (isRunFinished(detail.run.status) && intervalId) {
          clearInterval(intervalId)
          intervalId = null
        }
      } catch (err: unknown) {
        if (!mountedRef.current) return
        if (!initialFetchDoneRef.current) {
          setState({ status: 'error', message: err instanceof Error ? err.message : String(err) })
          initialFetchDoneRef.current = true
        }
      }
    }

    fetchDetail()
    intervalId = setInterval(fetchDetail, POLL_INTERVAL_MS)
    return () => { mountedRef.current = false; if (intervalId) clearInterval(intervalId) }
  }, [runId])

  const toggleFile = useCallback((idx: number) => {
    setExpandedFiles((prev) => { const n = new Set(prev); n.has(idx) ? n.delete(idx) : n.add(idx); return n })
  }, [])

  const toggleAgent = useCallback((idx: number) => {
    setExpandedAgents((prev) => { const n = new Set(prev); n.has(idx) ? n.delete(idx) : n.add(idx); return n })
  }, [])

  return (
    <div className="flex flex-1 min-h-0 flex-col">
      {/* ── Header bar ── */}
      <div className="flex shrink-0 items-center gap-3 border-b border-border px-4 py-3">
        <Button variant="ghost" size="sm" onClick={onBack} className="gap-1.5">
          <ArrowLeft size={14} />
          Volver
        </Button>
        <span className="text-sm text-muted-foreground">Run</span>
        <code className="font-mono text-xs text-brand-text">{runId}</code>
        {state.status === 'data' && state.detail.run.branch && (
          <span className="flex items-center gap-1 rounded-full border border-border bg-secondary/50 px-2 py-0.5 text-[11px] font-mono text-muted-foreground">
            <GitBranch size={12} />
            {state.detail.run.branch}
          </span>
        )}
      </div>

      {/* ── Loading / Error / Empty ── */}
      {state.status === 'loading' && (
        <div className="flex flex-1 items-center justify-center">
          <Loader2 size={20} className="animate-spin text-muted-foreground" />
        </div>
      )}
      {state.status === 'error' && (
        <div className="flex flex-1 flex-col items-center justify-center gap-2 p-6">
          <p className="text-sm text-fail">Error al cargar la sesión</p>
          <p className="max-w-sm text-center font-mono text-xs text-muted-foreground">{state.message}</p>
        </div>
      )}
      {state.status === 'empty' && (
        <div className="flex flex-1 items-center justify-center">
          <p className="text-sm text-muted-foreground">Sesión no encontrada.</p>
        </div>
      )}

      {state.status === 'data' && (() => {
        const d = state.detail
        const { prompts, turnStats } = state
        // Defensive: backend may return null arrays when binary is outdated
        const safeFiles = d.files ?? []
        const safeAgents = d.agents ?? []
        const safeToolUsage = d.toolUsage ?? []
        const safeToolDetails = d.toolDetails ?? []
        const safeTasks = d.tasks ?? []
        const safeActivityEvents = d.activityEvents ?? []
        const multiPrompt = prompts.length >= 2
        const hasTasks = safeTasks.length > 0
        const hasAgents = safeAgents.length > 0
        const hasCommands = safeToolDetails.length > 0

        return (
          <div className="flex-1 min-h-0 overflow-y-auto">
            {/* ════════════════════════════════════════════════════════
                SUMMARY ZONE
               ════════════════════════════════════════════════════════ */}
            <div className="border-b border-border">
              <div className="max-w-4xl px-6 py-4 space-y-3">

                {/* Error banner */}
                {d.run.errorReason && (
                  <div className="rounded-lg border border-fail-border bg-fail-bg px-4 py-2.5 flex items-center gap-2">
                    <AlertTriangle size={14} className="text-fail shrink-0" />
                    <p className="text-sm text-fail">
                      Error: <span className="font-mono font-medium">{d.run.errorReason}</span>
                    </p>
                  </div>
                )}

                {/* Status row — single inline row, no cards */}
                <div className="flex items-center gap-4 flex-wrap">
                  <StatusBadge status={d.run.status} />
                  <span className="font-mono text-sm text-foreground">
                    {d.run.durationMs > 0 ? formatDuration(d.run.durationMs) : '—'}
                  </span>
                  <span className="text-text-muted">|</span>
                  <span className="text-sm text-muted-foreground">{formatDate(d.run.startedAt)}</span>
                  <span className="text-text-muted">|</span>
                  <span className="font-mono text-sm text-muted-foreground">{safeFiles.length} archivos</span>
                </div>

                {/* Prompt / PromptTimeline */}
                {prompts.length > 0 ? (
                  <PromptTimeline prompts={prompts} turnStats={turnStats} />
                ) : d.run.taskDesc ? (
                  <div className="rounded-lg border border-border bg-card px-4 py-3 max-h-[120px] overflow-y-auto">
                    <p className="text-[11px] font-medium uppercase tracking-wide text-text-muted">Prompt</p>
                    <p className="mt-1 text-sm text-foreground whitespace-pre-wrap">{d.run.taskDesc}</p>
                  </div>
                ) : null}

                {/* Tasks progress bar (summary only) */}
                {hasTasks && <TasksBar tasks={safeTasks} />}
              </div>
            </div>

            {/* ════════════════════════════════════════════════════════
                TAB BAR — sticky so it stays visible while scrolling
               ════════════════════════════════════════════════════════ */}
            <div className="sticky top-0 z-10 border-b border-border bg-background" role="tablist">
              <div className="max-w-4xl px-6 flex">
                <TabButton id="traza" label="Traza" active={activeTab === 'traza'} onClick={() => setActiveTab('traza')} />
                <TabButton
                  id="cambios"
                  label={`Archivos${safeFiles.length > 0 ? ` (${safeFiles.length})` : ''}`}
                  active={activeTab === 'cambios'}
                  onClick={() => setActiveTab('cambios')}
                />
                {hasCommands && (
                  <TabButton
                    id="comandos"
                    label={`Comandos (${safeToolDetails.length})`}
                    active={activeTab === 'comandos'}
                    onClick={() => setActiveTab('comandos')}
                  />
                )}
                {hasAgents && (
                  <TabButton
                    id="agentes"
                    label={`Agentes (${safeAgents.length})`}
                    active={activeTab === 'agentes'}
                    onClick={() => setActiveTab('agentes')}
                  />
                )}
              </div>
            </div>

            {/* ════════════════════════════════════════════════════════
                TAB CONTENT
               ════════════════════════════════════════════════════════ */}
            <div>
              <div className="max-w-4xl px-6 py-5 space-y-6">

                {/* ── Tab: Traza ── */}
                {activeTab === 'traza' && (
                  <>
                    {/* Tool pills */}
                    {safeToolUsage.length > 0 && (
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="text-[11px] font-medium uppercase tracking-wide text-text-muted">Tools</span>
                        {safeToolUsage.map((t) => (
                          <span key={t.toolName} className="inline-flex items-center gap-1.5 rounded-full border border-border bg-card px-2.5 py-1 text-[11px] font-mono text-foreground">
                            {t.toolName}
                            <span className="text-muted-foreground">{t.count}</span>
                          </span>
                        ))}
                        {(d.run.compactionsCount ?? 0) > 0 && (
                          <span className="inline-flex items-center gap-1 rounded-full border border-running-border bg-running-bg px-2.5 py-1 text-[11px] font-mono text-running">
                            <Minimize2 size={10} />
                            {d.run.compactionsCount} compact
                          </span>
                        )}
                      </div>
                    )}

                    {/* Agent timeline */}
                    <AgentTimeline
                      agents={safeAgents}
                      runStartedAt={d.run.startedAt}
                      runDurationMs={d.run.durationMs}
                      runStatus={d.run.status}
                      activityEvents={safeActivityEvents}
                    />

                    {!safeAgents.length && !safeToolUsage.length && (
                      <p className="text-sm text-muted-foreground text-center py-8">Sin datos de ejecución disponibles.</p>
                    )}
                  </>
                )}

                {/* ── Tab: Cambios (archivos) ── */}
                {activeTab === 'cambios' && (
                  <>
                    {safeFiles.length === 0 ? (
                      <p className="text-sm text-muted-foreground text-center py-8">Sin archivos modificados en esta sesión.</p>
                    ) : multiPrompt ? (
                      <GroupedFilesList
                        files={safeFiles}
                        project={d.run.project}
                        agents={safeAgents}
                        prompts={prompts}
                        turnStats={turnStats}
                        expandedFiles={expandedFiles}
                        toggleFile={toggleFile}
                      />
                    ) : (
                      <FilesList
                        files={safeFiles}
                        project={d.run.project}
                        agents={safeAgents}
                        expandedFiles={expandedFiles}
                        toggleFile={toggleFile}
                      />
                    )}
                  </>
                )}

                {/* ── Tab: Comandos ── */}
                {activeTab === 'comandos' && (
                  multiPrompt ? (
                    <GroupedCommandsList commands={safeToolDetails} agents={safeAgents} prompts={prompts} turnStats={turnStats} />
                  ) : (
                    <CommandsTerminal commands={safeToolDetails} agents={safeAgents} />
                  )
                )}

                {/* ── Tab: Agentes ── */}
                {activeTab === 'agentes' && (
                  <>
                    {/* Full task checklist */}
                    {hasTasks && <TasksSection tasks={safeTasks} />}

                    {/* Agent list */}
                    {multiPrompt ? (
                      <GroupedAgentsList agents={safeAgents} prompts={prompts} turnStats={turnStats} expandedAgents={expandedAgents} toggleAgent={toggleAgent} />
                    ) : (
                      <section>
                        <h3 className="flex items-center gap-2 mb-3 text-sm font-medium text-foreground">
                          <Bot size={14} className="text-muted-foreground" />
                          Agentes ({safeAgents.length})
                        </h3>
                        <ul className="space-y-1">
                          {safeAgents.map((agent, idx) => {
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
                                    <ChevronRight size={14} className={`shrink-0 text-text-muted transition-transform ${isExpanded ? 'rotate-90' : ''}`} />
                                  )}
                                  <span className="font-mono text-xs text-foreground">{agent.role}</span>
                                  <StatusBadge status={agent.status} />
                                  {agent.durationMs != null && (
                                    <span className="font-mono text-[11px] text-muted-foreground">{formatDuration(agent.durationMs)}</span>
                                  )}
                                </button>
                                {hasOutput && isExpanded && (
                                  <div className="mt-1 ml-5 rounded-md border border-border bg-card p-4 max-h-[500px] overflow-auto">
                                    <pre className="whitespace-pre-wrap break-all font-mono text-xs text-foreground">{agent.output}</pre>
                                  </div>
                                )}
                              </li>
                            )
                          })}
                        </ul>
                      </section>
                    )}
                  </>
                )}
              </div>
            </div>
          </div>
        )
      })()}
    </div>
  )
}

// ── Tab Button ──────────────────────────────────────────────────────────────

function TabButton({ id, label, active, onClick }: { id: string; label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      role="tab"
      id={`tab-${id}`}
      aria-selected={active}
      aria-controls={`panel-${id}`}
      tabIndex={active ? 0 : -1}
      onClick={onClick}
      className={`px-4 h-9 text-[13px] font-medium border-b-2 transition-colors ${
        active
          ? 'border-brand-text text-foreground'
          : 'border-transparent text-muted-foreground hover:text-foreground'
      }`}
    >
      {label}
    </button>
  )
}

// ── Tasks Bar (summary — progress only) ─────────────────────────────────────

function TasksBar({ tasks }: { tasks: TaskDTO[] }) {
  const completed = tasks.filter((t) => t.status === 'completed').length
  const total = tasks.length
  const pct = total > 0 ? Math.round((completed / total) * 100) : 0

  return (
    <div className="flex items-center gap-3">
      <ListChecks size={14} className="text-muted-foreground shrink-0" />
      <div className="flex-1 h-1.5 rounded-full bg-secondary/30 overflow-hidden">
        <div className="h-full rounded-full bg-success transition-all duration-300" style={{ width: `${pct}%` }} />
      </div>
      <span className="font-mono text-[11px] text-muted-foreground shrink-0">{completed}/{total}</span>
    </div>
  )
}

// ── Tasks Section (full checklist — Agentes tab) ────────────────────────────

function TasksSection({ tasks }: { tasks: TaskDTO[] }) {
  const completed = tasks.filter((t) => t.status === 'completed').length
  const total = tasks.length
  const pct = total > 0 ? Math.round((completed / total) * 100) : 0

  return (
    <section>
      <h3 className="flex items-center gap-2 mb-3 text-sm font-medium text-foreground">
        <ListChecks size={14} className="text-muted-foreground" />
        Tasks ({completed}/{total})
      </h3>
      <div className="rounded-lg border border-border bg-card px-4 py-3 space-y-3">
        <div className="flex items-center gap-3">
          <div className="flex-1 h-2 rounded-full bg-secondary/30 overflow-hidden">
            <div className="h-full rounded-full bg-success transition-all duration-300" style={{ width: `${pct}%` }} />
          </div>
          <span className="font-mono text-[11px] text-muted-foreground shrink-0">{pct}%</span>
        </div>
        <ul className="space-y-1">
          {tasks.map((task) => (
            <li key={task.id} className="flex items-center gap-2 text-xs">
              <span className={`w-3.5 h-3.5 rounded-full border-2 shrink-0 flex items-center justify-center ${
                task.status === 'completed' ? 'border-success bg-success/20' : 'border-muted-foreground'
              }`}>
                {task.status === 'completed' && <span className="block w-1.5 h-1.5 rounded-full bg-success" />}
              </span>
              <span className={`flex-1 truncate ${task.status === 'completed' ? 'text-muted-foreground line-through' : 'text-foreground'}`}>
                {task.title || task.id}
              </span>
            </li>
          ))}
        </ul>
      </div>
    </section>
  )
}

// ── Commands Terminal ─────────────────────────────────────────────────────────

function toolPrefix(toolName: string): { prefix: string; cls: string } {
  switch (toolName) {
    case 'Bash':  return { prefix: '$', cls: 'text-running' }
    case 'Grep':  return { prefix: 'grep', cls: 'text-success' }
    case 'Glob':  return { prefix: 'find', cls: 'text-success' }
    case 'Read':  return { prefix: 'read', cls: 'text-brand-text' }
    case 'Agent': return { prefix: 'agent', cls: 'text-running' }
    default:      return { prefix: toolName.toLowerCase(), cls: 'text-muted-foreground' }
  }
}

function CommandsTerminal({ commands, agents }: { commands: ToolUseDetailDTO[]; agents: SessionAgentDTO[] }) {
  const agentRoleMap = useMemo(() => {
    const map = new Map<string, string>()
    for (const a of agents) map.set(a.id, a.role)
    return map
  }, [agents])

  return (
    <section>
      <div className="rounded-lg border border-border bg-[#0c0c14] overflow-hidden">
        {/* Terminal header */}
        <div className="flex items-center gap-2 px-4 py-2 bg-[#12121e] border-b border-border">
          <div className="flex gap-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-fail/60" />
            <span className="w-2.5 h-2.5 rounded-full bg-running/60" />
            <span className="w-2.5 h-2.5 rounded-full bg-success/60" />
          </div>
          <span className="text-[11px] font-mono text-muted-foreground ml-2">{commands.length} comandos</span>
        </div>

        {/* Command lines */}
        <div className="max-h-[500px] overflow-y-auto">
          {commands.map((cmd, idx) => {
            const { prefix, cls } = toolPrefix(cmd.toolName)
            const attribution = cmd.agentId ? (agentRoleMap.get(cmd.agentId) ?? 'agent') : ''
            return (
              <div
                key={idx}
                className="group flex items-start gap-0 px-4 py-1.5 hover:bg-white/[0.02] transition-colors border-b border-white/[0.04] last:border-0"
              >
                {/* Line number */}
                <span className="w-8 shrink-0 text-right pr-3 font-mono text-[10px] text-white/20 pt-px select-none">
                  {idx + 1}
                </span>
                {/* Prefix */}
                <span className={`shrink-0 font-mono text-xs font-bold ${cls} w-10 pt-px`}>
                  {prefix}
                </span>
                {/* Command */}
                <code className="flex-1 font-mono text-xs text-foreground whitespace-pre-wrap break-all">
                  {cmd.command}
                </code>
                {/* Agent attribution (on hover) */}
                {attribution && (
                  <span className="shrink-0 ml-2 opacity-0 group-hover:opacity-100 transition-opacity text-[10px] font-mono text-muted-foreground pt-px">
                    {attribution}
                  </span>
                )}
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}

// ── Agent Timeline (trace/waterfall) ─────────────────────────────────────────

interface SpanData { role: string; status: string; offsetMs: number; durationMs: number; id?: string }

function statusBarColor(status: string): string {
  switch (status) {
    case 'success': case 'completed': return 'bg-success'
    case 'failed': case 'fail': case 'error': return 'bg-fail'
    case 'running': case 'in_progress': case 'in-progress': return 'bg-running'
    default: return 'bg-pending'
  }
}

function AgentTimeline({ agents, runStartedAt, runDurationMs, runStatus, activityEvents }: {
  agents: SessionAgentDTO[]; runStartedAt: string; runDurationMs: number; runStatus: string; activityEvents: ActivityEventDTO[]
}) {
  const runStart = useMemo(() => new Date(runStartedAt).getTime(), [runStartedAt])

  const spans = useMemo<SpanData[]>(() => {
    if (isNaN(runStart)) return []
    return agents.filter((a) => a.startedAt).map((a) => {
      const start = new Date(a.startedAt).getTime()
      const offsetMs = Math.max(0, start - runStart)
      let dur = a.durationMs ?? 0
      if (dur === 0 && a.endedAt) dur = new Date(a.endedAt).getTime() - start
      if (dur <= 0 && !a.endedAt) dur = Date.now() - start
      return { role: a.role, status: a.status, offsetMs, durationMs: Math.max(dur, 0), id: a.id }
    })
  }, [agents, runStart])

  const { claudeDots, agentDotsMap } = useMemo(() => {
    const claude: number[] = []
    const byAgent = new Map<string, number[]>()
    if (isNaN(runStart)) return { claudeDots: claude, agentDotsMap: byAgent }
    for (const ev of activityEvents) {
      const ms = Math.max(0, new Date(ev.timestamp).getTime() - runStart)
      if (!ev.agentId) claude.push(ms)
      else { const arr = byAgent.get(ev.agentId) ?? []; arr.push(ms); byAgent.set(ev.agentId, arr) }
    }
    return { claudeDots: claude, agentDotsMap: byAgent }
  }, [activityEvents, runStart])

  const hasData = spans.length > 0 || claudeDots.length > 0
  if (!hasData) return null

  const isRunning = runStatus === 'running' || runStatus === 'in_progress'
  const nowOffset = isRunning ? Date.now() - runStart : 0
  const maxSpanEnd = spans.reduce((mx, s) => Math.max(mx, s.offsetMs + s.durationMs), 0)
  const allDots = [...claudeDots, ...Array.from(agentDotsMap.values()).flat()]
  const maxActivity = allDots.length > 0 ? Math.max(...allDots) : 0
  const totalMs = Math.max(runDurationMs, maxSpanEnd, maxActivity, nowOffset, 1)

  const markerCount = 5
  const markers = Array.from({ length: markerCount + 1 }, (_, i) => Math.round((totalMs / markerCount) * i))

  return (
    <section>
      <h3 className="flex items-center gap-2 mb-3 text-sm font-medium text-foreground">
        <Activity size={14} className="text-muted-foreground" />
        Traza de ejecución
      </h3>
      <div className="rounded-lg border border-border bg-card px-4 py-4 overflow-x-auto">
        <div className="relative h-4 ml-[120px] mb-1">
          {markers.map((ms, i) => (
            <span key={i} className="absolute text-[10px] font-mono text-text-muted -translate-x-1/2" style={{ left: `${(ms / totalMs) * 100}%` }}>
              {formatDuration(ms) === '—' ? '0s' : formatDuration(ms)}
            </span>
          ))}
        </div>
        <div className="space-y-1.5">
          {claudeDots.length > 0 && (
            <TimelineRow label="Claude" totalMs={totalMs} dots={claudeDots} dotColor="bg-brand"
              barClassName="bg-brand/15 border border-brand/25"
              barWidth={isRunning ? 100 : Math.min((runDurationMs / totalMs) * 100, 100)}
              suffix={`${claudeDots.length} edits`}
            />
          )}
          {spans.map((span, i) => {
            const leftPct = (span.offsetMs / totalMs) * 100
            const widthPct = Math.max((span.durationMs / totalMs) * 100, 0.5)
            const agentDots = span.id ? (agentDotsMap.get(span.id) ?? []) : []
            return (
              <div key={i} className="flex items-center gap-0 h-7">
                <div className="w-[120px] shrink-0 pr-2 text-right">
                  <span className="font-mono text-[11px] text-foreground truncate block">{span.role}</span>
                </div>
                <div className="relative flex-1 h-full rounded bg-secondary/30">
                  <div
                    className={`absolute top-0.5 bottom-0.5 rounded ${statusBarColor(span.status)} opacity-85 hover:opacity-100 transition-opacity cursor-default`}
                    style={{ left: `${leftPct}%`, width: `${widthPct}%` }}
                    title={`${span.role}: ${formatDuration(span.durationMs)} (offset: +${formatDuration(span.offsetMs)})`}
                  >
                    {widthPct > 8 && (
                      <span className="absolute inset-0 flex items-center justify-center text-[10px] font-mono text-white font-medium drop-shadow-sm">
                        {formatDuration(span.durationMs)}
                      </span>
                    )}
                  </div>
                  {agentDots.map((ms, j) => (
                    <div key={j} className="absolute top-0 bottom-0 w-1 rounded-full bg-success opacity-60"
                      style={{ left: `${(ms / totalMs) * 100}%` }}
                      title={`${span.role} editó archivo: +${formatDuration(ms)}`}
                    />
                  ))}
                </div>
                {widthPct <= 8 && <span className="ml-1 shrink-0 font-mono text-[10px] text-text-muted">{formatDuration(span.durationMs)}</span>}
                {agentDots.length > 0 && widthPct > 8 && <span className="ml-1 shrink-0 font-mono text-[10px] text-text-muted">{agentDots.length} edits</span>}
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}

function TimelineRow({ label, totalMs, dots, dotColor, barClassName, barWidth, suffix }: {
  label: string; totalMs: number; dots: number[]; dotColor: string; barClassName: string; barWidth: number; suffix: string
}) {
  return (
    <div className="flex items-center gap-0 h-7">
      <div className="w-[120px] shrink-0 pr-2 text-right">
        <span className="font-mono text-[11px] text-foreground truncate block">{label}</span>
      </div>
      <div className="relative flex-1 h-full rounded bg-secondary/30">
        <div className={`absolute top-0.5 bottom-0.5 rounded ${barClassName}`} style={{ left: '0%', width: `${barWidth}%` }} />
        {dots.map((ms, i) => (
          <div key={i} className={`absolute top-1 bottom-1 w-1 rounded-full ${dotColor} opacity-70`}
            style={{ left: `${(ms / totalMs) * 100}%` }}
            title={`${label} editó archivo: +${formatDuration(ms)}`}
          />
        ))}
      </div>
      <span className="ml-1 shrink-0 font-mono text-[10px] text-text-muted">{suffix}</span>
    </div>
  )
}

// ── Files List with diff viewer & revert ─────────────────────────────────────

function actionBadge(action: string) {
  const map: Record<string, { label: string; cls: string }> = {
    create: { label: 'A', cls: 'bg-success/20 text-success border-success/30' },
    write: { label: 'M', cls: 'bg-brand-subtle text-brand-text border-brand/30' },
    modify: { label: 'M', cls: 'bg-brand-subtle text-brand-text border-brand/30' },
    delete: { label: 'D', cls: 'bg-fail-bg text-fail border-fail-border' },
    rename: { label: 'R', cls: 'bg-running-bg text-running border-running-border' },
  }
  const cfg = map[action] ?? { label: action.charAt(0).toUpperCase(), cls: 'bg-secondary text-text-muted border-border' }
  return <span className={`inline-flex items-center justify-center w-5 h-5 rounded text-[10px] font-bold border ${cfg.cls}`}>{cfg.label}</span>
}

function FilesList({ files, project, agents, expandedFiles, toggleFile }: {
  files: { path: string; action: string; diff?: string; agentId: string }[]
  project: string; agents: SessionAgentDTO[]; expandedFiles: Set<number>; toggleFile: (idx: number) => void
}) {
  const agentRoleMap = useMemo(() => {
    const map = new Map<string, string>()
    for (const a of agents) map.set(a.id, a.role)
    return map
  }, [agents])

  const [reverting, setReverting] = useState<number | null>(null)
  const [revertResult, setRevertResult] = useState<{ idx: number; ok: boolean; msg: string } | null>(null)

  const handleRevert = useCallback(async (idx: number, filePath: string, diff: string) => {
    if (!window.confirm(`¿Revertir los cambios en ${filePath}?\n\nEsto aplicará el parche inverso con git apply -R.`)) return
    setReverting(idx); setRevertResult(null)
    try {
      await revertFile(project, filePath, diff)
      setRevertResult({ idx, ok: true, msg: 'Cambios revertidos' })
    } catch (err: unknown) {
      setRevertResult({ idx, ok: false, msg: err instanceof Error ? err.message : String(err) })
    } finally { setReverting(null) }
  }, [project])

  return (
    <section>
      <h3 className="flex items-center gap-2 mb-3 text-sm font-medium text-foreground">
        <FileText size={14} className="text-muted-foreground" />
        Archivos modificados ({files.length})
      </h3>
      {files.length === 0 ? (
        <p className="text-sm text-text-muted">Sin archivos modificados en esta sesión.</p>
      ) : (
        <div className="space-y-2">
          {files.map((file, idx) => {
            const hasDiff = !!file.diff
            const isExpanded = expandedFiles.has(idx)
            const isReverting = reverting === idx
            const result = revertResult?.idx === idx ? revertResult : null
            return (
              <div key={idx} className="rounded-lg border border-border overflow-hidden">
                <div
                  className={`flex items-center gap-2 px-3 py-2 bg-card ${hasDiff ? 'cursor-pointer hover:bg-secondary/40' : 'cursor-default'} transition-colors`}
                  onClick={() => hasDiff && toggleFile(idx)}
                  role={hasDiff ? 'button' : undefined}
                  tabIndex={hasDiff ? 0 : undefined}
                  onKeyDown={(e) => { if (hasDiff && (e.key === 'Enter' || e.key === ' ')) { e.preventDefault(); toggleFile(idx) } }}
                >
                  {hasDiff && <ChevronRight size={14} className={`shrink-0 text-text-muted transition-transform duration-150 ${isExpanded ? 'rotate-90' : ''}`} />}
                  {actionBadge(file.action)}
                  <span className="flex-1 truncate font-mono text-xs text-foreground">{file.path}</span>
                  <span className={`shrink-0 rounded-full px-2 py-0.5 text-[10px] font-medium border ${
                    file.agentId ? 'bg-success-bg text-success border-success-border' : 'bg-brand-subtle text-brand-text border-brand/30'
                  }`}>
                    {file.agentId ? (agentRoleMap.get(file.agentId) ?? 'agent') : 'Claude'}
                  </span>
                  {hasDiff && (
                    <Button variant="ghost" size="sm" className="h-6 px-2 text-[11px] gap-1 text-text-muted hover:text-fail shrink-0"
                      onClick={(e) => { e.stopPropagation(); handleRevert(idx, file.path, file.diff!) }}
                      disabled={isReverting}
                    >
                      {isReverting ? <Loader2 size={12} className="animate-spin" /> : <Undo2 size={12} />}
                      Revertir
                    </Button>
                  )}
                </div>
                {result && (
                  <div className={`px-3 py-1.5 text-xs border-t border-border ${result.ok ? 'bg-success-bg text-success' : 'bg-fail-bg text-fail'}`}>
                    {result.msg}
                  </div>
                )}
                {hasDiff && isExpanded && (
                  <div className="border-t border-border max-h-[500px] overflow-auto">
                    <DiffViewer diff={file.diff!} />
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}
    </section>
  )
}

// ── Turn-assignment helper ─────────────────────────────────────────────────────
// Returns the turnNumber for an item based on its timestamp range.
// Items with no timestamp or before first turn default to turn 1.
function assignTurn(itemTs: string, turnStats: TurnStatsDTO[]): number {
  if (!itemTs || turnStats.length === 0) return 1
  const t = new Date(itemTs).getTime()
  if (isNaN(t)) return 1
  for (const s of turnStats) {
    const start = new Date(s.timestamp).getTime()
    const end = s.endTimestamp ? new Date(s.endTimestamp).getTime() : Infinity
    if (t >= start && t < end) return s.turnNumber
  }
  // fallback: last turn
  return turnStats[turnStats.length - 1].turnNumber
}

// ── Grouped Files List ─────────────────────────────────────────────────────────

function GroupedFilesList({ files, project, agents, prompts, turnStats, expandedFiles, toggleFile }: {
  files: { path: string; action: string; diff?: string; agentId: string; timestamp?: string }[]
  project: string
  agents: SessionAgentDTO[]
  prompts: PromptDTO[]
  turnStats: TurnStatsDTO[]
  expandedFiles: Set<number>
  toggleFile: (idx: number) => void
}) {
  const agentRoleMap = useMemo(() => {
    const map = new Map<string, string>()
    for (const a of agents) map.set(a.id, a.role)
    return map
  }, [agents])

  const [reverting, setReverting] = useState<number | null>(null)
  const [revertResult, setRevertResult] = useState<{ idx: number; ok: boolean; msg: string } | null>(null)

  const handleRevert = useCallback(async (idx: number, filePath: string, diff: string) => {
    if (!window.confirm(`¿Revertir los cambios en ${filePath}?\n\nEsto aplicará el parche inverso con git apply -R.`)) return
    setReverting(idx); setRevertResult(null)
    try {
      await revertFile(project, filePath, diff)
      setRevertResult({ idx, ok: true, msg: 'Cambios revertidos' })
    } catch (err: unknown) {
      setRevertResult({ idx, ok: false, msg: err instanceof Error ? err.message : String(err) })
    } finally { setReverting(null) }
  }, [project])

  // Group file indices by turn
  const groups = useMemo(() => {
    const map = new Map<number, number[]>()
    files.forEach((f, idx) => {
      const turn = assignTurn((f as { timestamp?: string }).timestamp ?? '', turnStats)
      const arr = map.get(turn) ?? []
      arr.push(idx)
      map.set(turn, arr)
    })
    return map
  }, [files, turnStats])

  const promptMap = useMemo(() => {
    const m = new Map<number, string>()
    for (const p of prompts) m.set(p.sequence, p.prompt)
    return m
  }, [prompts])

  return (
    <section className="space-y-3">
      <h3 className="flex items-center gap-2 mb-3 text-sm font-medium text-foreground">
        <FileText size={14} className="text-muted-foreground" />
        Archivos modificados ({files.length})
      </h3>
      {prompts.map((p, gi) => {
        const indices = groups.get(p.sequence) ?? []
        if (indices.length === 0) return null
        return (
          <TurnGroup
            key={p.sequence}
            turnNumber={p.sequence}
            promptText={promptMap.get(p.sequence) ?? ''}
            count={indices.length}
            countLabel="archivos"
            defaultOpen={gi === 0}
          >
            <div className="divide-y divide-border">
              {indices.map((idx) => {
                const file = files[idx]
                const hasDiff = !!file.diff
                const isExpanded = expandedFiles.has(idx)
                const isReverting = reverting === idx
                const result = revertResult?.idx === idx ? revertResult : null
                return (
                  <div key={idx}>
                    <div
                      className={`flex items-center gap-2 px-3 py-2 bg-card ${hasDiff ? 'cursor-pointer hover:bg-secondary/40' : 'cursor-default'} transition-colors`}
                      onClick={() => hasDiff && toggleFile(idx)}
                      role={hasDiff ? 'button' : undefined}
                      tabIndex={hasDiff ? 0 : undefined}
                      onKeyDown={(e) => { if (hasDiff && (e.key === 'Enter' || e.key === ' ')) { e.preventDefault(); toggleFile(idx) } }}
                    >
                      {hasDiff && <ChevronRight size={14} className={`shrink-0 text-text-muted transition-transform duration-150 ${isExpanded ? 'rotate-90' : ''}`} />}
                      {actionBadge(file.action)}
                      <span className="flex-1 truncate font-mono text-xs text-foreground">{file.path}</span>
                      <span className={`shrink-0 rounded-full px-2 py-0.5 text-[10px] font-medium border ${
                        file.agentId ? 'bg-success-bg text-success border-success-border' : 'bg-brand-subtle text-brand-text border-brand/30'
                      }`}>
                        {file.agentId ? (agentRoleMap.get(file.agentId) ?? 'agent') : 'Claude'}
                      </span>
                      {hasDiff && (
                        <Button variant="ghost" size="sm" className="h-6 px-2 text-[11px] gap-1 text-text-muted hover:text-fail shrink-0"
                          onClick={(e) => { e.stopPropagation(); handleRevert(idx, file.path, file.diff!) }}
                          disabled={isReverting}
                        >
                          {isReverting ? <Loader2 size={12} className="animate-spin" /> : <Undo2 size={12} />}
                          Revertir
                        </Button>
                      )}
                    </div>
                    {result && (
                      <div className={`px-3 py-1.5 text-xs border-t border-border ${result.ok ? 'bg-success-bg text-success' : 'bg-fail-bg text-fail'}`}>
                        {result.msg}
                      </div>
                    )}
                    {hasDiff && isExpanded && (
                      <div className="border-t border-border max-h-[500px] overflow-auto">
                        <DiffViewer diff={file.diff!} />
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          </TurnGroup>
        )
      })}
    </section>
  )
}

// ── Grouped Commands List ──────────────────────────────────────────────────────

function GroupedCommandsList({ commands, agents, prompts, turnStats }: {
  commands: ToolUseDetailDTO[]
  agents: SessionAgentDTO[]
  prompts: PromptDTO[]
  turnStats: TurnStatsDTO[]
}) {
  const agentRoleMap = useMemo(() => {
    const map = new Map<string, string>()
    for (const a of agents) map.set(a.id, a.role)
    return map
  }, [agents])

  const groups = useMemo(() => {
    const map = new Map<number, number[]>()
    commands.forEach((c, idx) => {
      const turn = assignTurn(c.timestamp, turnStats)
      const arr = map.get(turn) ?? []
      arr.push(idx)
      map.set(turn, arr)
    })
    return map
  }, [commands, turnStats])

  const promptMap = useMemo(() => {
    const m = new Map<number, string>()
    for (const p of prompts) m.set(p.sequence, p.prompt)
    return m
  }, [prompts])

  return (
    <section className="space-y-3">
      {prompts.map((p, gi) => {
        const indices = groups.get(p.sequence) ?? []
        if (indices.length === 0) return null
        return (
          <TurnGroup
            key={p.sequence}
            turnNumber={p.sequence}
            promptText={promptMap.get(p.sequence) ?? ''}
            count={indices.length}
            countLabel="comandos"
            defaultOpen={gi === 0}
          >
            <div className="rounded-b-lg border-border bg-[#0c0c14] overflow-hidden">
              {indices.map((idx) => {
                const cmd = commands[idx]
                const { prefix, cls } = toolPrefix(cmd.toolName)
                const attribution = cmd.agentId ? (agentRoleMap.get(cmd.agentId) ?? 'agent') : ''
                return (
                  <div
                    key={idx}
                    className="group flex items-start gap-0 px-4 py-1.5 hover:bg-white/[0.02] transition-colors border-b border-white/[0.04] last:border-0"
                  >
                    <span className="w-8 shrink-0 text-right pr-3 font-mono text-[10px] text-white/20 pt-px select-none">
                      {idx + 1}
                    </span>
                    <span className={`shrink-0 font-mono text-xs font-bold ${cls} w-10 pt-px`}>
                      {prefix}
                    </span>
                    <code className="flex-1 font-mono text-xs text-foreground whitespace-pre-wrap break-all">
                      {cmd.command}
                    </code>
                    {attribution && (
                      <span className="shrink-0 ml-2 opacity-0 group-hover:opacity-100 transition-opacity text-[10px] font-mono text-muted-foreground pt-px">
                        {attribution}
                      </span>
                    )}
                  </div>
                )
              })}
            </div>
          </TurnGroup>
        )
      })}
    </section>
  )
}

// ── Grouped Agents List ────────────────────────────────────────────────────────

function GroupedAgentsList({ agents, prompts, turnStats, expandedAgents, toggleAgent }: {
  agents: SessionAgentDTO[]
  prompts: PromptDTO[]
  turnStats: TurnStatsDTO[]
  expandedAgents: Set<number>
  toggleAgent: (idx: number) => void
}) {
  const groups = useMemo(() => {
    const map = new Map<number, number[]>()
    agents.forEach((a, idx) => {
      const turn = assignTurn(a.startedAt, turnStats)
      const arr = map.get(turn) ?? []
      arr.push(idx)
      map.set(turn, arr)
    })
    return map
  }, [agents, turnStats])

  const promptMap = useMemo(() => {
    const m = new Map<number, string>()
    for (const p of prompts) m.set(p.sequence, p.prompt)
    return m
  }, [prompts])

  return (
    <section className="space-y-3">
      <h3 className="flex items-center gap-2 mb-3 text-sm font-medium text-foreground">
        <Bot size={14} className="text-muted-foreground" />
        Agentes ({agents.length})
      </h3>
      {prompts.map((p, gi) => {
        const indices = groups.get(p.sequence) ?? []
        if (indices.length === 0) return null
        return (
          <TurnGroup
            key={p.sequence}
            turnNumber={p.sequence}
            promptText={promptMap.get(p.sequence) ?? ''}
            count={indices.length}
            countLabel="agentes"
            defaultOpen={gi === 0}
          >
            <ul className="divide-y divide-border">
              {indices.map((idx) => {
                const agent = agents[idx]
                const hasOutput = !!agent.output
                const isExpanded = expandedAgents.has(idx)
                return (
                  <li key={idx}>
                    <button
                      type="button"
                      onClick={() => hasOutput && toggleAgent(idx)}
                      className={`flex w-full items-center gap-3 px-3 py-2 bg-card text-left transition-colors ${hasOutput ? 'cursor-pointer hover:bg-secondary/50' : 'cursor-default'}`}
                    >
                      {hasOutput && (
                        <ChevronRight size={14} className={`shrink-0 text-text-muted transition-transform ${isExpanded ? 'rotate-90' : ''}`} />
                      )}
                      <span className="font-mono text-xs text-foreground">{agent.role}</span>
                      <StatusBadge status={agent.status} />
                      {agent.durationMs != null && (
                        <span className="font-mono text-[11px] text-muted-foreground">{formatDuration(agent.durationMs)}</span>
                      )}
                    </button>
                    {hasOutput && isExpanded && (
                      <div className="mx-3 mb-2 rounded-md border border-border bg-card p-4 max-h-[500px] overflow-auto">
                        <pre className="whitespace-pre-wrap break-all font-mono text-xs text-foreground">{agent.output}</pre>
                      </div>
                    )}
                  </li>
                )
              })}
            </ul>
          </TurnGroup>
        )
      })}
    </section>
  )
}
