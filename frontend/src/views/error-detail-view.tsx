import { useEffect, useState } from 'react'
import { ArrowLeft, ChevronRight, GitCommit } from 'lucide-react'
import { ResolutionBadge } from '@/components/resolution-badge'
import { TrendMiniChart } from '@/components/trend-chart'
import { formatDate } from '@/lib/format'
import { cn } from '@/lib/utils'
import {
  getErrorGroup,
  updateErrorResolution,
  type ErrorGroupDetailDTO,
} from '@/lib/wails'

interface ErrorDetailViewProps {
  groupId: string
  onBack: () => void
  onRunClick: (runId: string) => void
}

const STATUSES = ['new', 'investigating', 'resolved', 'ignored'] as const

const STATUS_STYLES: Record<string, string> = {
  new: 'text-[hsl(var(--status-new))] border-[hsl(var(--status-new-border))] bg-[hsl(var(--status-new-bg))]',
  investigating: 'text-[hsl(var(--status-investigating))] border-[hsl(var(--status-investigating-border))] bg-[hsl(var(--status-investigating-bg))]',
  resolved: 'text-[hsl(var(--status-resolved))] border-[hsl(var(--status-resolved-border))] bg-[hsl(var(--status-resolved-bg))]',
  ignored: 'text-[hsl(var(--status-ignored))] border-[hsl(var(--status-ignored-border))] bg-[hsl(var(--status-ignored-bg))]',
}

export function ErrorDetailView({ groupId, onBack, onRunClick }: ErrorDetailViewProps) {
  const [detail, setDetail] = useState<ErrorGroupDetailDTO | null>(null)
  const [loading, setLoading] = useState(true)
  const [status, setStatus] = useState('')
  const [notes, setNotes] = useState('')
  const [commitLink, setCommitLink] = useState('')
  const [saving, setSaving] = useState(false)
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    setLoading(true)
    getErrorGroup(groupId).then((data) => {
      setDetail(data)
      if (data) {
        setStatus(data.group.resolutionStatus)
        setNotes(data.group.notes)
        setCommitLink(data.group.commitLink)
      }
      setLoading(false)
    })
  }, [groupId])

  const handleSave = async () => {
    setSaving(true)
    try {
      await updateErrorResolution({ groupId, status, notes, commitLink })
      // Refresh detail
      const data = await getErrorGroup(groupId)
      if (data) {
        setDetail(data)
        setStatus(data.group.resolutionStatus)
        setNotes(data.group.notes)
        setCommitLink(data.group.commitLink)
      }
      setDirty(false)
    } catch (err) {
      console.error('Error saving resolution:', err)
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="p-6 space-y-3">
        {[0, 1, 2].map((i) => (
          <div key={i} className="h-16 rounded-md bg-secondary/50 animate-pulse" />
        ))}
      </div>
    )
  }

  if (!detail) {
    return (
      <div className="p-6">
        <button onClick={onBack} className="inline-flex items-center gap-1.5 px-2 py-1 rounded-md text-xs text-muted-foreground hover:text-foreground hover:bg-secondary/50">
          <ArrowLeft size={14} /> Errors
        </button>
        <p className="mt-4 text-sm text-muted-foreground">Error group no encontrado.</p>
      </div>
    )
  }

  const { group, runs, history, trend } = detail

  return (
    <div className="h-full overflow-y-auto">
      {/* Back button */}
      <div className="px-6 pt-4">
        <button
          onClick={onBack}
          className="inline-flex items-center gap-1.5 px-2 py-1 rounded-md text-xs text-muted-foreground hover:text-foreground hover:bg-secondary/50"
        >
          <ArrowLeft size={14} /> Errors
        </button>
      </div>

      {/* Header */}
      <div className="px-6 py-4">
        <h2 className="text-base font-medium text-foreground truncate">{group.title}</h2>
        <p className="text-[11px] font-mono text-[hsl(var(--text-muted))] mt-0.5">{group.fingerprint}</p>
        <div className="flex items-center gap-2 mt-2">
          <ResolutionBadge status={group.resolutionStatus} />
          <span className="text-xs font-mono text-foreground">{group.occurrenceCount} ocurrencias</span>
        </div>
      </div>

      {/* Two columns */}
      <div className="flex gap-6 px-6 pb-6">
        {/* Left column */}
        <div className="flex-1 min-w-0 space-y-4">
          {/* Error context */}
          <div className="bg-card border border-border rounded-lg p-4">
            <p className="text-sm text-foreground font-mono whitespace-pre-wrap break-all">{group.title}</p>
            {group.normalizedMsg !== group.title && (
              <div className="mt-3 p-3 rounded-md bg-background text-[11px] font-mono text-[hsl(var(--text-secondary))] overflow-x-auto max-h-[300px] overflow-y-auto">
                {group.normalizedMsg}
              </div>
            )}

            {/* Agents involved */}
            {runs.length > 0 && (
              <div className="mt-4 pt-3 border-t border-border">
                <p className="text-[11px] text-muted-foreground mb-2">Agentes involucrados</p>
                <div className="flex flex-wrap gap-1">
                  {[...new Set(runs.filter((r) => r.agentName).map((r) => r.agentName))].map((agent) => (
                    <span key={agent} className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-[hsl(var(--brand-subtle))] text-[hsl(var(--brand-text))] text-[10px] font-mono">
                      {agent}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Trend card */}
          {trend.length > 0 && (
            <div className="bg-card border border-border rounded-lg p-4">
              <p className="text-xs font-medium text-foreground mb-3">Tendencia (14 dias)</p>
              <TrendMiniChart data={trend} width={500} height={120} />
              <div className="flex justify-between mt-2 text-[10px] text-[hsl(var(--text-muted))] font-mono">
                <span>{trend[0]?.date}</span>
                <span>{trend[trend.length - 1]?.date}</span>
              </div>
              <div className="flex gap-6 mt-3">
                <span className="text-xs text-[hsl(var(--text-secondary))]">Primera vez: {formatDate(group.firstSeenAt)}</span>
                <span className="text-xs text-[hsl(var(--text-secondary))]">Ultima vez: {formatDate(group.lastSeenAt)}</span>
              </div>
            </div>
          )}

          {/* Affected runs */}
          <div className="border border-border rounded-lg overflow-hidden">
            <div className="px-4 py-2 bg-card text-xs font-medium text-foreground">
              Runs afectados ({runs.length})
            </div>
            {runs.slice(0, 10).map((r) => (
              <button
                key={`${r.runId}-${r.agentName}`}
                onClick={() => onRunClick(r.runId)}
                className="w-full flex items-center gap-3 px-4 py-2 hover:bg-secondary/50 transition-colors text-left border-t border-border"
              >
                <span className="text-xs font-mono text-foreground truncate flex-1">{r.runId}</span>
                {r.agentName && (
                  <span className="text-[10px] font-mono text-[hsl(var(--text-muted))]">{r.agentName}</span>
                )}
                <span className="text-xs text-[hsl(var(--text-secondary))]">{formatDate(r.occurredAt)}</span>
                <ChevronRight size={14} className="text-muted-foreground" />
              </button>
            ))}
            {runs.length > 10 && (
              <div className="px-4 py-2 border-t border-border">
                <span className="text-xs text-[hsl(var(--brand-text))]">
                  Ver todos ({runs.length})
                </span>
              </div>
            )}
          </div>
        </div>

        {/* Right column - Resolution panel */}
        <div className="w-[320px] shrink-0">
          <div className="sticky top-4 bg-card border border-border rounded-lg p-4">
            <p className="text-sm font-medium text-foreground mb-3">Resolucion</p>

            {/* Status selector */}
            <div className="flex gap-1" role="radiogroup" aria-label="Estado de resolucion">
              {STATUSES.map((s) => (
                <button
                  key={s}
                  role="radio"
                  aria-checked={status === s}
                  onClick={() => { setStatus(s); setDirty(true) }}
                  className={cn(
                    'px-3 py-1.5 text-xs rounded-md border transition-colors capitalize',
                    status === s
                      ? STATUS_STYLES[s]
                      : 'bg-transparent text-muted-foreground border-border hover:text-foreground',
                  )}
                >
                  {s}
                </button>
              ))}
            </div>

            {/* Notes */}
            <textarea
              value={notes}
              onChange={(e) => { setNotes(e.target.value); setDirty(true) }}
              placeholder="Notas de diagnostico..."
              className="w-full min-h-[80px] mt-3 rounded-md border border-border bg-background p-2 text-xs text-foreground resize-y focus:ring-1 focus:ring-ring focus:outline-none"
            />

            {/* Commit link */}
            <div className="relative mt-2">
              <GitCommit size={12} className="absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                value={commitLink}
                onChange={(e) => { setCommitLink(e.target.value); setDirty(true) }}
                placeholder="Link al commit (opcional)"
                className="h-8 w-full rounded-md border border-border bg-background px-2 pl-7 text-xs font-mono text-foreground focus:ring-1 focus:ring-ring focus:outline-none"
              />
            </div>

            {/* Save button */}
            <button
              onClick={handleSave}
              disabled={!dirty || saving}
              className={cn(
                'mt-3 w-full px-4 py-2 rounded-md text-xs font-medium transition-colors',
                dirty && !saving
                  ? 'bg-foreground text-background hover:bg-foreground/90'
                  : 'bg-secondary text-muted-foreground cursor-not-allowed',
              )}
            >
              {saving ? 'Guardando...' : 'Guardar'}
            </button>

            {/* History */}
            {history.length > 0 && (
              <div className="mt-4 pt-3 border-t border-border space-y-2">
                {history.map((h, i) => (
                  <div key={i} className="text-[11px] text-[hsl(var(--text-muted))]">
                    <span className="font-mono">{formatDate(h.createdAt)}</span>
                    {' '}cambio status a{' '}
                    <ResolutionBadge status={h.newStatus} />
                    {h.note && <span className="block mt-0.5 italic">"{h.note}"</span>}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

