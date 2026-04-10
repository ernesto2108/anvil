import { useEffect, useState, useCallback } from 'react'
import { BarChart2, AlertTriangle, TrendingUp, TrendingDown } from 'lucide-react'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts'
import { getMetrics, type MetricsDTO } from '@/lib/wails'
import { formatDuration } from '@/lib/format'
import { cn } from '@/lib/utils'

// --- Constantes de diseño -----------------------------------------------

const COLOR_BRAND = '#8B5CF6'
const COLOR_GRID = '#27272A'
const COLOR_SUCCESS = '#4ADE80'
const COLOR_FAIL = '#F87171'
const COLOR_TEXT_MUTED = '#A1A1AA'

const ENOUGH_DATA_THRESHOLD = 5

// --- Tipos internos -------------------------------------------------------

type ViewState =
  | { kind: 'loading' }
  | { kind: 'error'; err: string }
  | { kind: 'empty' }
  | { kind: 'insufficient'; data: MetricsDTO }
  | { kind: 'data'; data: MetricsDTO }

type DeltaDirection = 'up' | 'down'
type DeltaSentiment = 'positive' | 'negative'

interface DeltaChip {
  direction: DeltaDirection
  label: string
  sentiment: DeltaSentiment
}

// --- Helpers ---------------------------------------------------------------

const formatTokensAbbrev = (n: number): string => {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${Math.round(n / 1_000)}K`
  return String(n)
}

const formatSuccessRate = (rate: number): string =>
  `${Math.round(rate * 100)}%`

const shortDate = (iso: string): string => {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return '?'
  return `${d.getDate()}/${d.getMonth() + 1}`
}

// buildTokensDelta: positivo = ROJO (más tokens = peor).
const buildTokensDelta = (pct: number | null): DeltaChip | null => {
  if (pct === null) return null
  const label = `${pct >= 0 ? '+' : ''}${Math.round(pct * 100)}%`
  if (pct > 0) return { direction: 'up', label, sentiment: 'negative' }
  if (pct < 0) return { direction: 'down', label, sentiment: 'positive' }
  return { direction: 'up', label: '0%', sentiment: 'positive' }
}

// buildDurationDelta: positivo = ROJO (más tiempo = peor).
const buildDurationDelta = (ms: number | null): DeltaChip | null => {
  if (ms === null) return null
  const label = `${ms >= 0 ? '+' : ''}${formatDuration(Math.abs(ms))}`
  if (ms > 0) return { direction: 'up', label, sentiment: 'negative' }
  if (ms < 0) return { direction: 'down', label, sentiment: 'positive' }
  return { direction: 'up', label: '0ms', sentiment: 'positive' }
}

// buildSuccessRateDelta: positivo = VERDE (más éxito = mejor).
const buildSuccessRateDelta = (delta: number | null): DeltaChip | null => {
  if (delta === null) return null
  const pct = Math.round(delta * 100)
  const label = `${pct >= 0 ? '+' : ''}${pct}pp`
  if (delta > 0) return { direction: 'up', label, sentiment: 'positive' }
  if (delta < 0) return { direction: 'down', label, sentiment: 'negative' }
  return { direction: 'up', label: '0pp', sentiment: 'positive' }
}

// --- Subcomponentes -------------------------------------------------------

interface MetricCardProps {
  label: string
  value: string
  delta: DeltaChip | null
  'data-testid'?: string
}

function MetricCard({ label, value, delta, 'data-testid': testId }: MetricCardProps) {
  const sentimentColor =
    delta?.sentiment === 'positive' ? COLOR_SUCCESS : COLOR_FAIL

  return (
    <div
      data-testid={testId}
      className="rounded-lg border border-border bg-card px-5 py-4"
    >
      <p className="text-[11px] font-medium uppercase tracking-wide text-text-muted">
        {label}
      </p>
      <p className="mt-2 font-mono text-2xl font-semibold text-foreground">{value}</p>
      {delta !== null ? (
        <div
          className="mt-2 flex items-center gap-1"
          style={{ color: sentimentColor }}
        >
          {delta.direction === 'up' ? (
            <TrendingUp size={13} />
          ) : (
            <TrendingDown size={13} />
          )}
          <span className="text-xs font-medium">{delta.label}</span>
        </div>
      ) : (
        <p className="mt-2 text-xs text-text-muted">—</p>
      )}
    </div>
  )
}

// Skeleton de una card.
function CardSkeleton({ height }: { height: number }) {
  return (
    <div
      className="rounded-lg border border-border bg-card animate-pulse"
      style={{ height }}
    />
  )
}

// Tooltip personalizado para el chart de tokens.
const TokensTooltip = ({
  active,
  payload,
}: {
  active?: boolean
  payload?: Array<{ payload: { runId: string; tokens: number } }>
}) => {
  if (!active || !payload?.length) return null
  const { runId, tokens } = payload[0].payload
  return (
    <div className="rounded-md border border-border bg-card px-3 py-2 text-xs shadow-md">
      <p className="font-mono text-text-muted">{runId}</p>
      <p className="mt-0.5 font-semibold text-foreground">
        {formatTokensAbbrev(tokens)} tokens
      </p>
    </div>
  )
}

// Tooltip personalizado para el chart de tiempo por agente.
const AgentTimeTooltip = ({
  active,
  payload,
}: {
  active?: boolean
  payload?: Array<{ payload: { role: string; ms: number; runsIncluded: number } }>
}) => {
  if (!active || !payload?.length) return null
  const { role, ms, runsIncluded } = payload[0].payload
  return (
    <div className="rounded-md border border-border bg-card px-3 py-2 text-xs shadow-md">
      <p className="font-semibold text-foreground">{role}</p>
      <p className="mt-0.5 text-text-muted">{formatDuration(ms)}</p>
      <p className="text-text-muted">{runsIncluded} runs</p>
    </div>
  )
}

// --- Secciones de estado --------------------------------------------------

function LoadingState() {
  return (
    <div data-testid="metrics-loading" className="p-6 space-y-6">
      {/* Fila 1: 3 metric cards */}
      <div className="grid grid-cols-3 gap-4">
        <CardSkeleton height={112} />
        <CardSkeleton height={112} />
        <CardSkeleton height={112} />
      </div>
      {/* Fila 2: 2 chart cards */}
      <div className="grid grid-cols-2 gap-4">
        <CardSkeleton height={280} />
        <CardSkeleton height={280} />
      </div>
    </div>
  )
}

function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div
      data-testid="metrics-error"
      className="flex flex-1 flex-col items-center justify-center gap-4 p-6"
    >
      <div className="rounded-lg border border-border bg-card px-8 py-8 flex flex-col items-center gap-3 max-w-sm w-full text-center">
        <AlertTriangle size={32} className="text-fail" />
        <p className="text-sm font-medium text-foreground">
          No se pudieron cargar las métricas
        </p>
        <p className="font-mono text-xs text-text-muted break-all">{message}</p>
        <button
          onClick={onRetry}
          aria-label="Reintentar carga de métricas"
          className="mt-2 rounded-md bg-brand px-4 py-2 text-sm font-medium text-white hover:bg-brand/80 transition-colors"
        >
          Reintentar
        </button>
      </div>
    </div>
  )
}

function EmptyState() {
  return (
    <div
      data-testid="metrics-empty"
      className="flex flex-1 flex-col items-center justify-center gap-4 px-6 min-h-[400px]"
    >
      <BarChart2 size={48} className="text-text-muted" />
      <div className="text-center space-y-1">
        <p className="text-sm font-medium text-foreground">Aún no hay métricas</p>
        <p className="text-sm text-text-muted">
          Ejecuta tu primera orquestación para comenzar a ver datos
        </p>
      </div>
      <div className="bg-card border border-border rounded-md px-4 py-2.5 font-mono text-sm text-text-muted">
        $ anvil orchestrate &quot;tu tarea&quot;
      </div>
    </div>
  )
}

function InsufficientState({ data }: { data: MetricsDTO }) {
  return (
    <div data-testid="metrics-insufficient" className="p-6 space-y-6">
      {/* Fila 1: 3 MetricCards con valores reales, deltas "—" */}
      <div className="grid grid-cols-3 gap-4">
        <MetricCard
          data-testid="metric-card-tokens"
          label="TOTAL TOKENS"
          value={formatTokensAbbrev(data.totalTokens)}
          delta={null}
        />
        <MetricCard
          data-testid="metric-card-duration"
          label="AVG DURATION"
          value={formatDuration(data.avgDurationMs)}
          delta={null}
        />
        <MetricCard
          data-testid="metric-card-success-rate"
          label="SUCCESS RATE"
          value={formatSuccessRate(data.successRate)}
          delta={null}
        />
      </div>
      {/* Fila 2: placeholder unificado */}
      <div className="grid grid-cols-2 gap-4">
        <div className="col-span-2 rounded-lg border border-border bg-card flex flex-col items-center justify-center gap-3 py-12">
          <BarChart2 size={32} className="text-text-muted" />
          <p className="text-sm text-text-muted">
            Se necesitan al menos 5 ejecuciones para ver tendencias
          </p>
          <span className="rounded-full border border-border px-3 py-1 text-xs text-text-muted font-mono">
            {data.runsCount} de {ENOUGH_DATA_THRESHOLD}
          </span>
        </div>
      </div>
    </div>
  )
}

function DataState({ data }: { data: MetricsDTO }) {
  const tokensDelta = buildTokensDelta(data.totalTokensDeltaPct)
  const durationDelta = buildDurationDelta(data.avgDurationDeltaMs)
  const successRateDelta = buildSuccessRateDelta(data.successRateDelta)

  const tokensChartData = data.tokensPerRun.map((p) => ({
    name: shortDate(p.startedAt),
    tokens: p.tokens,
    runId: p.runId,
  }))

  const agentTimeChartData = data.avgTimePerAgent.map((a) => ({
    role: a.role,
    ms: a.avgDurationMs,
    runsIncluded: a.runsIncluded,
  }))

  return (
    <div data-testid="metrics-data" className="p-6 space-y-6">
      {/* Fila 1: 3 MetricCards */}
      <div className="grid grid-cols-3 gap-4">
        <MetricCard
          data-testid="metric-card-tokens"
          label="TOTAL TOKENS"
          value={formatTokensAbbrev(data.totalTokens)}
          delta={tokensDelta}
        />
        <MetricCard
          data-testid="metric-card-duration"
          label="AVG DURATION"
          value={formatDuration(data.avgDurationMs)}
          delta={durationDelta}
        />
        <MetricCard
          data-testid="metric-card-success-rate"
          label="SUCCESS RATE"
          value={formatSuccessRate(data.successRate)}
          delta={successRateDelta}
        />
      </div>

      {/* Fila 2: 2 charts */}
      <div className="grid grid-cols-2 gap-4">
        {/* Chart 1: Tokens per run (barras verticales) */}
        <div
          data-testid="chart-tokens-per-run"
          className="rounded-lg border border-border bg-card px-5 py-4"
        >
          <p className="mb-4 text-[11px] font-medium uppercase tracking-wide text-text-muted">
            Tokens per run
          </p>
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={tokensChartData} margin={{ top: 0, right: 4, bottom: 0, left: 4 }}>
              <CartesianGrid vertical={false} stroke={COLOR_GRID} strokeDasharray="0" />
              <XAxis
                dataKey="name"
                tick={{ fontSize: 10, fill: COLOR_TEXT_MUTED }}
                axisLine={false}
                tickLine={false}
                interval={Math.max(0, Math.floor(tokensChartData.length / 6) - 1)}
              />
              <YAxis
                tick={{ fontSize: 10, fill: COLOR_TEXT_MUTED }}
                axisLine={false}
                tickLine={false}
                tickFormatter={formatTokensAbbrev}
                width={40}
              />
              <Tooltip content={<TokensTooltip />} cursor={{ fill: COLOR_GRID }} />
              <Bar dataKey="tokens" fill={COLOR_BRAND} radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Chart 2: Avg time per agent (barras horizontales) */}
        <div
          data-testid="chart-avg-time-per-agent"
          className="rounded-lg border border-border bg-card px-5 py-4"
        >
          <p className="mb-4 text-[11px] font-medium uppercase tracking-wide text-text-muted">
            Avg time per agent
          </p>
          <ResponsiveContainer width="100%" height={220}>
            <BarChart
              layout="vertical"
              data={agentTimeChartData}
              margin={{ top: 0, right: 4, bottom: 0, left: 4 }}
            >
              <CartesianGrid horizontal={false} stroke={COLOR_GRID} strokeDasharray="0" />
              <XAxis type="number" hide />
              <YAxis
                dataKey="role"
                type="category"
                width={80}
                tick={{ fontSize: 10, fill: COLOR_TEXT_MUTED }}
                axisLine={false}
                tickLine={false}
              />
              <Tooltip content={<AgentTimeTooltip />} cursor={{ fill: COLOR_GRID }} />
              <Bar dataKey="ms" fill={COLOR_BRAND} radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  )
}

// --- Componente principal -------------------------------------------------

export function MetricsView() {
  const [state, setState] = useState<ViewState>({ kind: 'loading' })

  const fetchMetrics = useCallback(() => {
    setState({ kind: 'loading' })
    let cancelled = false

    getMetrics()
      .then((data) => {
        if (cancelled) return
        if (data.runsCount === 0) {
          setState({ kind: 'empty' })
        } else if (!data.hasEnoughData) {
          setState({ kind: 'insufficient', data })
        } else {
          setState({ kind: 'data', data })
        }
      })
      .catch((err: unknown) => {
        if (cancelled) return
        const msg = err instanceof Error ? err.message : String(err)
        setState({ kind: 'error', err: msg })
      })

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    const cleanup = fetchMetrics()
    return cleanup
  }, [fetchMetrics])

  return (
    <div className="flex h-full flex-col">
      {/* Topbar interno de la vista */}
      <div className="flex shrink-0 items-center justify-between border-b border-border px-6 py-3">
        <h2 className="text-sm font-semibold text-foreground">Metrics</h2>
        <button
          data-testid="metrics-range-selector"
          aria-disabled="true"
          title="Configurable en próxima versión"
          className={cn(
            'flex items-center gap-1.5 rounded-md border border-border bg-card px-3 py-1.5',
            'text-xs text-text-muted cursor-default select-none',
          )}
        >
          Last 30 runs
          <span className="text-[10px]">▼</span>
        </button>
      </div>

      {/* Contenido según estado */}
      <div className="flex-1 overflow-auto">
        {state.kind === 'loading' && <LoadingState />}
        {state.kind === 'error' && (
          <ErrorState message={state.err} onRetry={fetchMetrics} />
        )}
        {state.kind === 'empty' && <EmptyState />}
        {state.kind === 'insufficient' && <InsufficientState data={state.data} />}
        {state.kind === 'data' && <DataState data={state.data} />}
      </div>
    </div>
  )
}
