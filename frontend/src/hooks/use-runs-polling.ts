import { useCallback, useEffect, useRef, useState } from 'react'
import { getRuns, type RunDTO } from '@/lib/wails'

const POLL_INTERVAL_MS = 2000

export interface RunsFilter {
  status: string
  startDate: string
  endDate: string
  project: string
}

interface UseRunsPollingResult {
  runs: RunDTO[] | null
  error: boolean
  refresh: () => void
}

/**
 * useRunsPolling — carga la lista de runs y la refresca cada 2s.
 *
 * Comportamiento:
 * - Primera carga: `runs` es null hasta que el fetch completa (muestra skeleton).
 * - Refetches posteriores: actualizan `runs` en background sin flash de skeleton.
 * - Error en primer fetch: `error` es true y `runs` queda null.
 * - Error en refetch posterior: log a consola, mantiene datos previos (no bloquea UI).
 * - Cleanup en unmount: el interval se limpia y se cancela cualquier setState pendiente.
 * - Cambio de filtros: reinicia el estado a null (skeleton) y hace un nuevo fetch.
 */
export function useRunsPolling(filter: RunsFilter): UseRunsPollingResult {
  const [runs, setRuns] = useState<RunDTO[] | null>(null)
  const [error, setError] = useState(false)
  const [refreshTick, setRefreshTick] = useState(0)

  const refresh = useCallback(() => setRefreshTick((t) => t + 1), [])

  // Ref para rastrear si el componente está montado y evitar setState tras unmount.
  const mountedRef = useRef(true)
  // Ref que indica si el primer fetch ya completó (para distinguir initial vs refetch).
  const initialFetchDoneRef = useRef(false)

  useEffect(() => {
    mountedRef.current = true
    initialFetchDoneRef.current = false
    setRuns(null)
    setError(false)

    async function fetchRuns() {
      try {
        const data = await getRuns({
          limit: 100,
          offset: 0,
          status: filter.status,
          startDate: filter.startDate,
          endDate: filter.endDate,
          project: filter.project,
        })
        if (!mountedRef.current) return
        setRuns(data)
        setError(false)
        initialFetchDoneRef.current = true
      } catch (err) {
        if (!mountedRef.current) return
        if (!initialFetchDoneRef.current) {
          // Primer fetch fallido: mostrar error.
          console.error('[useRunsPolling] error en carga inicial:', err)
          setRuns([])
          setError(true)
          initialFetchDoneRef.current = true
        } else {
          // Refetch fallido: solo log, mantener datos previos.
          console.warn('[useRunsPolling] error en refetch (manteniendo datos previos):', err)
        }
      }
    }

    // Fetch inicial inmediato.
    fetchRuns()

    // Polling cada 2s.
    const interval = setInterval(fetchRuns, POLL_INTERVAL_MS)

    return () => {
      mountedRef.current = false
      clearInterval(interval)
    }
  }, [filter.status, filter.startDate, filter.endDate, filter.project, refreshTick])

  return { runs, error, refresh }
}
