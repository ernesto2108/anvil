// Formateadores con locale es-ES para la UI del dashboard.

const dateFormatter = new Intl.DateTimeFormat('es-ES', {
  day: 'numeric',
  month: 'short',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
})

const numberFormatter = new Intl.NumberFormat('es-ES')

// formatDate convierte un string ISO-8601 a formato legible en español.
// Retorna "—" si el string está vacío.
export function formatDate(iso: string): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (isNaN(d.getTime())) return '—'
  return dateFormatter.format(d)
}

// formatDuration convierte milisegundos a texto legible (p.ej. "2m 34s", "850ms").
// Retorna "—" si ms es 0.
export function formatDuration(ms: number): string {
  if (ms === 0) return '—'
  if (ms < 1000) return `${ms}ms`

  const totalSeconds = Math.floor(ms / 1000)
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  if (hours > 0) {
    return `${hours}h ${String(minutes).padStart(2, '0')}m ${String(seconds).padStart(2, '0')}s`
  }
  if (minutes > 0) {
    return `${minutes}m ${String(seconds).padStart(2, '0')}s`
  }
  return `${seconds}s`
}

// formatTokens convierte un número a formato local es-ES (p.ej. "1.234").
export function formatTokens(n: number): string {
  return numberFormatter.format(n)
}

// formatTokensFlow convierte un número de tokens al formato compacto usado en
// los nodos del grafo de flujo (p.ej. "8.432 tk"). Mantiene el locale es-ES
// para consistencia con el resto del dashboard.
export function formatTokensFlow(n: number): string {
  return `${numberFormatter.format(n)} tk`
}

export type QAScoreTone = 'success' | 'warn' | 'fail' | 'muted'

export interface QAScoreResult {
  text: string
  tone: QAScoreTone
}

// formatQAScore clasifica el qa_score en un tono visual y texto formateado.
// null → "—" muted; >=80 success, >=60 warn, <60 fail.
export function formatQAScore(score: number | null): QAScoreResult {
  if (score === null) return { text: '—', tone: 'muted' }
  if (score >= 80) return { text: `${score}`, tone: 'success' }
  if (score >= 60) return { text: `${score}`, tone: 'warn' }
  return { text: `${score}`, tone: 'fail' }
}
