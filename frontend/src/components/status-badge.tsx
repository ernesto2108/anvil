import { CircleCheck, CircleX, CircleDashed, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface StatusBadgeProps {
  status: string
}

interface StatusConfig {
  label: string
  icon: React.ReactNode
  className: string
}

function getStatusConfig(status: string): StatusConfig {
  switch (status) {
    case 'success':
    case 'completed':
      return {
        label: 'Success',
        icon: <CircleCheck size={11} />,
        className:
          'text-success bg-success-bg border-success-border',
      }
    case 'failed':
    case 'fail':
    case 'error':
      return {
        label: 'Failed',
        icon: <CircleX size={11} />,
        className: 'text-fail bg-fail-bg border-fail-border',
      }
    case 'running':
    case 'in_progress':
    case 'in-progress':
      return {
        label: 'Running',
        icon: <Loader2 size={11} className="motion-safe:animate-spin" />,
        className:
          'text-running bg-running-bg border-running-border',
      }
    default:
      return {
        label: 'Pending',
        icon: <CircleDashed size={11} />,
        className:
          'text-pending bg-pending-bg border-pending-border',
      }
  }
}

export function StatusBadge({ status }: StatusBadgeProps) {
  const config = getStatusConfig(status)
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
