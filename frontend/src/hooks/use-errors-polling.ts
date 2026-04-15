import { useCallback, useEffect, useRef, useState } from 'react'
import { getErrorGroups, getErrorsCounts, type ErrorGroupDTO, type ErrorsCountDTO } from '@/lib/wails'

const POLL_INTERVAL_MS = 5000

export interface ErrorsFilter {
  status: string
  search: string
}

interface UseErrorsPollingResult {
  errors: ErrorGroupDTO[] | null
  counts: ErrorsCountDTO | null
  fetchError: boolean
  refresh: () => void
}

export function useErrorsPolling(filter: ErrorsFilter): UseErrorsPollingResult {
  const [errors, setErrors] = useState<ErrorGroupDTO[] | null>(null)
  const [counts, setCounts] = useState<ErrorsCountDTO | null>(null)
  const [fetchError, setFetchError] = useState(false)
  const [refreshTick, setRefreshTick] = useState(0)

  const refresh = useCallback(() => setRefreshTick((t) => t + 1), [])

  const mountedRef = useRef(true)
  const initialFetchDoneRef = useRef(false)

  useEffect(() => {
    mountedRef.current = true
    initialFetchDoneRef.current = false
    setErrors(null)
    setCounts(null)
    setFetchError(false)

    async function fetchData() {
      try {
        const [data, c] = await Promise.all([
          getErrorGroups({ status: filter.status, search: filter.search }),
          getErrorsCounts(),
        ])
        if (!mountedRef.current) return
        setErrors(data)
        setCounts(c)
        setFetchError(false)
        initialFetchDoneRef.current = true
      } catch (err) {
        if (!mountedRef.current) return
        if (!initialFetchDoneRef.current) {
          console.error('[useErrorsPolling] error en carga inicial:', err)
          setErrors([])
          setFetchError(true)
          initialFetchDoneRef.current = true
        } else {
          console.warn('[useErrorsPolling] error en refetch:', err)
        }
      }
    }

    fetchData()
    const interval = setInterval(fetchData, POLL_INTERVAL_MS)

    return () => {
      mountedRef.current = false
      clearInterval(interval)
    }
  }, [filter.status, filter.search, refreshTick])

  return { errors, counts, fetchError, refresh }
}
