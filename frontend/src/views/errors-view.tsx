import { useState } from 'react'
import { ChevronRight, Search, ShieldCheck, SearchX, TrendingUp, TrendingDown, Minus, X } from 'lucide-react'
import { ResolutionBadge } from '@/components/resolution-badge'
import { formatDate } from '@/lib/format'
import { useErrorsPolling, type ErrorsFilter } from '@/hooks/use-errors-polling'
import type { ErrorGroupDTO } from '@/lib/wails'

interface ErrorsViewProps {
  onGroupClick: (groupId: string) => void
}

const STATUS_OPTIONS = [
  { value: '', label: 'Todos' },
  { value: 'new', label: 'New' },
  { value: 'investigating', label: 'Investigating' },
  { value: 'resolved', label: 'Resolved' },
  { value: 'ignored', label: 'Ignored' },
] as const

function LoadingSkeleton() {
  return (
    <div className="p-6 space-y-3">
      {[0, 1, 2].map((i) => (
        <div key={i} className="h-10 rounded-md bg-secondary/50 animate-pulse" />
      ))}
    </div>
  )
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center h-full min-h-[400px] gap-4 px-6">
      <ShieldCheck size={48} className="text-muted-foreground" />
      <div className="text-center space-y-1">
        <p className="text-sm font-medium text-foreground">Sin errores detectados</p>
        <p className="text-sm text-muted-foreground">
          Todos los runs se ejecutaron correctamente.
        </p>
      </div>
    </div>
  )
}

function EmptyFilterState({ onClear }: { onClear: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center h-full min-h-[400px] gap-4 px-6">
      <SearchX size={36} className="text-muted-foreground" />
      <div className="text-center space-y-1">
        <p className="text-sm text-foreground">Sin resultados para los filtros seleccionados</p>
        <button
          onClick={onClear}
          className="text-xs text-[hsl(var(--brand-text))] hover:underline"
        >
          Limpiar filtros
        </button>
      </div>
    </div>
  )
}

function ErrorGroupRow({ group, onClick }: { group: ErrorGroupDTO; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="w-full flex items-center gap-3 rounded-md px-3 py-2.5 hover:bg-secondary/50 transition-colors cursor-pointer text-left group"
    >
      <ResolutionBadge status={group.resolutionStatus} />

      <div className="flex-1 min-w-0">
        <p className="text-xs font-medium text-foreground truncate">{group.title}</p>
        <p className="text-[10px] font-mono text-[hsl(var(--text-muted))] truncate">{group.fingerprint}</p>
      </div>

      <div className="w-[80px] shrink-0 flex items-center gap-1 justify-end">
        <span className="text-xs font-mono text-foreground">{group.occurrenceCount}</span>
        <TrendIndicator count={group.occurrenceCount} />
      </div>

      <div className="w-[140px] shrink-0 text-right">
        <span className="text-xs text-[hsl(var(--text-secondary))]">{formatDate(group.lastSeenAt)}</span>
      </div>

      <ChevronRight size={14} className="shrink-0 text-muted-foreground group-hover:text-foreground transition-colors" />
    </button>
  )
}

function TrendIndicator({ count }: { count: number }) {
  if (count > 3) return <TrendingUp size={12} className="text-[hsl(var(--trend-up))]" />
  if (count === 1) return <TrendingDown size={12} className="text-[hsl(var(--trend-down))]" />
  return <Minus size={12} className="text-[hsl(var(--trend-flat))]" />
}

export function ErrorsView({ onGroupClick }: ErrorsViewProps) {
  const [filter, setFilter] = useState<ErrorsFilter>({ status: '', search: '' })
  const [searchInput, setSearchInput] = useState('')

  const { errors, counts } = useErrorsPolling(filter)

  const hasFilters = filter.status !== '' || filter.search !== ''

  const clearFilters = () => {
    setFilter({ status: '', search: '' })
    setSearchInput('')
  }

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setFilter((f) => ({ ...f, search: searchInput }))
  }

  if (errors === null) return <LoadingSkeleton />

  return (
    <div className="h-full flex flex-col">
      {/* Filter bar */}
      <div className="flex items-center gap-3 px-6 pt-6 pb-2 flex-wrap">
        <select
          value={filter.status}
          onChange={(e) => setFilter((f) => ({ ...f, status: e.target.value }))}
          className="h-8 rounded-md border border-border bg-background px-2 text-xs text-foreground"
        >
          {STATUS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>

        <form onSubmit={handleSearchSubmit} className="relative">
          <Search size={12} className="absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            onBlur={() => setFilter((f) => ({ ...f, search: searchInput }))}
            placeholder="Buscar en mensajes de error..."
            className="h-8 w-[240px] rounded-md border border-border bg-background px-2 pl-7 text-xs text-foreground focus:ring-1 focus:ring-ring focus:outline-none"
          />
        </form>

        {hasFilters && (
          <button
            onClick={clearFilters}
            className="h-8 px-2 rounded-md text-xs text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors"
          >
            <X size={14} />
          </button>
        )}
      </div>

      {/* Summary bar */}
      {counts && (
        <div className="flex items-center gap-4 px-6 py-2">
          <span className="text-[11px] font-mono">
            <span className="text-foreground font-semibold">{counts.new}</span>{' '}
            <span className="text-muted-foreground">new</span>
          </span>
          <div className="w-px h-3 bg-border" />
          <span className="text-[11px] font-mono">
            <span className="text-foreground font-semibold">{counts.investigating}</span>{' '}
            <span className="text-muted-foreground">investigating</span>
          </span>
          <div className="w-px h-3 bg-border" />
          <span className="text-[11px] font-mono">
            <span className="text-foreground font-semibold">{counts.weekTotal}</span>{' '}
            <span className="text-muted-foreground">esta semana</span>
          </span>
        </div>
      )}

      {/* Error list */}
      <div className="flex-1 overflow-y-auto px-3 pb-6">
        {errors.length === 0 && !hasFilters && <EmptyState />}
        {errors.length === 0 && hasFilters && <EmptyFilterState onClear={clearFilters} />}
        {errors.map((group) => (
          <ErrorGroupRow
            key={group.id}
            group={group}
            onClick={() => onGroupClick(group.id)}
          />
        ))}
      </div>
    </div>
  )
}
