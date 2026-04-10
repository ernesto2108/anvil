import { cn } from '@/lib/utils'
import { formatQAScore } from '@/lib/format'

interface QABadgeProps {
  score: number | null
  className?: string
}

// QABadge muestra el qa_score (escala 0-10) con color según su tono.
// Cuando score es null renderiza "—" en tono muted.
export function QABadge({ score, className }: QABadgeProps) {
  const { text, tone } = formatQAScore(score)
  const colorClass = {
    success: 'text-success',
    warn: 'text-running',
    fail: 'text-fail',
    muted: 'text-muted-foreground',
  }[tone]
  return <span className={cn('font-mono text-xs', colorClass, className)}>{text}</span>
}
