import { useState } from 'react'
import { ChevronRight } from 'lucide-react'

interface TurnGroupProps {
  turnNumber: number
  promptText: string
  count: number
  countLabel: string
  defaultOpen?: boolean
  children: React.ReactNode
}

export function TurnGroup({ turnNumber, promptText, count, countLabel, defaultOpen = false, children }: TurnGroupProps) {
  const [open, setOpen] = useState(defaultOpen)

  return (
    <div className="rounded-lg border border-border overflow-hidden">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center gap-2 px-3 py-2 bg-card hover:bg-secondary/40 transition-colors text-left"
      >
        <ChevronRight
          size={14}
          className={`shrink-0 text-text-muted transition-transform duration-150 ${open ? 'rotate-90' : ''}`}
        />
        {/* Purple turn badge */}
        <span className="shrink-0 inline-flex items-center justify-center w-5 h-5 rounded-full bg-purple-950 border border-purple-500/40 text-[10px] font-bold font-mono text-purple-300">
          {turnNumber}
        </span>
        {/* Prompt text truncated */}
        <span className="flex-1 truncate font-mono text-xs text-muted-foreground">
          {promptText}
        </span>
        {/* Count label */}
        <span className="shrink-0 text-[11px] font-mono text-text-muted">
          {count} {countLabel}
        </span>
      </button>

      {open && (
        <div className="border-t border-border">
          {children}
        </div>
      )}
    </div>
  )
}
