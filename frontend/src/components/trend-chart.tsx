import type { TrendPointDTO } from '@/lib/wails'

interface TrendMiniChartProps {
  data: TrendPointDTO[]
  width?: number
  height?: number
}

export function TrendMiniChart({ data, width = 120, height = 32 }: TrendMiniChartProps) {
  if (!data || data.length === 0) return null

  const max = Math.max(...data.map((d) => d.count), 1)
  const padding = 2

  const points = data.map((d, i) => {
    const x = padding + (i / (data.length - 1)) * (width - padding * 2)
    const y = height - padding - (d.count / max) * (height - padding * 2)
    return `${x},${y}`
  })

  const linePath = `M ${points.join(' L ')}`

  // Area path: line + bottom-right + bottom-left
  const areaPath = `${linePath} L ${width - padding},${height - padding} L ${padding},${height - padding} Z`

  return (
    <svg width={width} height={height} className="shrink-0">
      <path d={areaPath} fill="hsl(var(--fail-bg))" opacity={0.5} />
      <path d={linePath} fill="none" stroke="hsl(var(--fail))" strokeWidth={1.5} />
    </svg>
  )
}
