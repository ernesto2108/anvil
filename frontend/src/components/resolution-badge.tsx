import { CircleAlert, Search, CircleCheck, EyeOff } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ResolutionBadgeProps {
  status: string
}

interface ResolutionConfig {
  label: string
  icon: React.ReactNode
  className: string
}

function getConfig(status: string): ResolutionConfig {
  switch (status) {
    case 'new':
      return {
        label: 'New',
        icon: <CircleAlert size={11} />,
        className: 'text-[hsl(var(--status-new))] bg-[hsl(var(--status-new-bg))] border-[hsl(var(--status-new-border))]',
      }
    case 'investigating':
      return {
        label: 'Investigating',
        icon: <Search size={11} />,
        className: 'text-[hsl(var(--status-investigating))] bg-[hsl(var(--status-investigating-bg))] border-[hsl(var(--status-investigating-border))]',
      }
    case 'resolved':
      return {
        label: 'Resolved',
        icon: <CircleCheck size={11} />,
        className: 'text-[hsl(var(--status-resolved))] bg-[hsl(var(--status-resolved-bg))] border-[hsl(var(--status-resolved-border))]',
      }
    case 'ignored':
      return {
        label: 'Ignored',
        icon: <EyeOff size={11} />,
        className: 'text-[hsl(var(--status-ignored))] bg-[hsl(var(--status-ignored-bg))] border-[hsl(var(--status-ignored-border))]',
      }
    default:
      return {
        label: status,
        icon: <CircleAlert size={11} />,
        className: 'text-muted-foreground bg-secondary border-border',
      }
  }
}

export function ResolutionBadge({ status }: ResolutionBadgeProps) {
  const config = getConfig(status)
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full px-2 py-1 text-[11px] font-medium border',
        config.className,
      )}
    >
      {config.icon}
      {config.label}
    </span>
  )
}
