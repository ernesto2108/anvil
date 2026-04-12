import { RefreshCw } from 'lucide-react'

interface TopbarProps {
  title: string
  actions?: React.ReactNode
}

export function Topbar({ title, actions }: TopbarProps) {
  return (
    <header className="h-12 flex items-center justify-between px-6 border-b border-border bg-background">
      <h1 className="text-sm font-semibold text-foreground">{title}</h1>

      <div className="flex items-center gap-3">
        {actions}
        <div className="flex items-center gap-1.5">
          <span className="w-1.5 h-1.5 rounded-full bg-success" />
          <span className="font-mono text-[11px] text-muted-foreground">Live</span>
        </div>
        <RefreshCw size={14} className="text-muted-foreground" />
      </div>
    </header>
  )
}
