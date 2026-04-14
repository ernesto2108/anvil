import { useState, useCallback } from 'react'
import { FileText, Terminal, Bot } from 'lucide-react'
import type { PromptDTO, TurnStatsDTO } from '@/lib/wails'

interface PromptTimelineProps {
  prompts: PromptDTO[]
  turnStats: TurnStatsDTO[]
}

function formatRelativeTime(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  const diffMs = Date.now() - d.getTime()
  const diffSec = Math.floor(diffMs / 1000)
  if (diffSec < 60) return 'hace un momento'
  const diffMin = Math.floor(diffSec / 60)
  if (diffMin < 60) return `hace ${diffMin}m`
  const diffHr = Math.floor(diffMin / 60)
  if (diffHr < 24) return `hace ${diffHr}h`
  const diffDay = Math.floor(diffHr / 24)
  return `hace ${diffDay}d`
}

export function PromptTimeline({ prompts, turnStats }: PromptTimelineProps) {
  // Single prompt: simple card (current behaviour)
  if (prompts.length <= 1) {
    const prompt = prompts[0]
    if (!prompt) return null
    return (
      <div className="rounded-lg border border-border bg-card px-4 py-3 max-h-[120px] overflow-y-auto">
        <p className="text-[11px] font-medium uppercase tracking-wide text-text-muted">Prompt</p>
        <p className="mt-1 text-sm text-foreground whitespace-pre-wrap">{prompt.prompt}</p>
      </div>
    )
  }

  // Multi-prompt: timeline
  return <MultiPromptTimeline prompts={prompts} turnStats={turnStats} />
}

function MultiPromptTimeline({ prompts, turnStats }: { prompts: PromptDTO[]; turnStats: TurnStatsDTO[] }) {
  // Sequence #1 expanded by default
  const [expandedSet, setExpandedSet] = useState<Set<number>>(() => new Set([1]))

  const toggle = useCallback((seq: number) => {
    setExpandedSet((prev) => {
      const next = new Set(prev)
      next.has(seq) ? next.delete(seq) : next.add(seq)
      return next
    })
  }, [])

  const allExpanded = expandedSet.size === prompts.length

  const toggleAll = useCallback(() => {
    if (allExpanded) {
      setExpandedSet(new Set())
    } else {
      setExpandedSet(new Set(prompts.map((p) => p.sequence)))
    }
  }, [allExpanded, prompts])

  const statsMap = new Map<number, TurnStatsDTO>()
  for (const s of turnStats) statsMap.set(s.turnNumber, s)

  return (
    <div>
      {/* Header */}
      <div className="flex items-center gap-2 mb-2">
        <span className="text-[11px] font-medium uppercase tracking-wide text-text-muted">
          Prompts ({prompts.length})
        </span>
        <button
          type="button"
          onClick={toggleAll}
          className="text-[11px] text-purple-400 hover:text-purple-300 transition-colors ml-auto"
        >
          {allExpanded ? 'colapsar todo' : 'expandir todo'}
        </button>
      </div>

      {/* Timeline */}
      <div className="relative">
        {/* Vertical line */}
        <div className="absolute left-[9px] top-4 bottom-4 w-0.5 bg-purple-500/20" />

        <div className="space-y-2">
          {prompts.map((p) => {
            const isExpanded = expandedSet.has(p.sequence)
            const stats = statsMap.get(p.sequence)
            const relTime = formatRelativeTime(p.timestamp)

            return (
              <div key={p.sequence} className="relative flex gap-3">
                {/* Badge */}
                <div className="shrink-0 z-10 mt-2.5">
                  <span className="inline-flex items-center justify-center w-[22px] h-[22px] rounded-full bg-purple-950 border border-purple-500/40 text-[10px] font-bold font-mono text-purple-300">
                    #{p.sequence}
                  </span>
                </div>

                {/* Card */}
                <div className={`flex-1 rounded-lg border overflow-hidden transition-colors ${
                  isExpanded
                    ? 'border-purple-500/40 bg-card'
                    : 'border-border bg-card'
                }`}>
                  <button
                    type="button"
                    onClick={() => toggle(p.sequence)}
                    className="w-full text-left"
                  >
                    <div className={`px-3 py-2 ${isExpanded ? 'border-l-2 border-purple-500/60' : ''}`}>
                      {isExpanded ? (
                        <p className="font-mono text-xs text-foreground whitespace-pre-wrap break-words">
                          {p.prompt}
                        </p>
                      ) : (
                        <p className="font-mono text-xs text-muted-foreground truncate">
                          {p.prompt}
                        </p>
                      )}
                    </div>
                  </button>

                  {/* Turn stats + timestamp */}
                  <div className="flex items-center gap-3 px-3 py-1.5 bg-secondary/20 border-t border-border/60">
                    {stats && (
                      <>
                        <StatPill icon={<FileText size={10} />} count={stats.filesCount} />
                        <StatPill icon={<Terminal size={10} />} count={stats.toolUsesCount} />
                        <StatPill icon={<Bot size={10} />} count={stats.agentsCount} />
                      </>
                    )}
                    {relTime && (
                      <span className="ml-auto text-[10px] font-mono text-text-muted">{relTime}</span>
                    )}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}

function StatPill({ icon, count }: { icon: React.ReactNode; count: number }) {
  return (
    <span className="inline-flex items-center gap-1 text-[10px] font-mono text-text-muted">
      <span className="text-muted-foreground">{icon}</span>
      {count}
    </span>
  )
}
