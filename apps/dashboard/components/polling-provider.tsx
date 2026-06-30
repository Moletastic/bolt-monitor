'use client'

import { useEffect, useRef, useCallback } from 'react'
import { useRouter } from 'next/navigation'

interface PollingProviderProps {
  intervalMs?: number
}

export function PollingProvider({ intervalMs = 5000 }: PollingProviderProps) {
  const router = useRouter()
  const intervalRef = useRef<NodeJS.Timeout | null>(null)
  const isVisibleRef = useRef(true)

  const startPolling = useCallback(() => {
    if (intervalRef.current) return

    intervalRef.current = setInterval(() => {
      if (isVisibleRef.current) {
        router.refresh()
      }
    }, intervalMs)
  }, [router, intervalMs])

  const stopPolling = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
  }, [])

  const handleVisibilityChange = useCallback(() => {
    if (document.visibilityState === 'visible') {
      isVisibleRef.current = true
      router.refresh()
      startPolling()
    } else {
      isVisibleRef.current = false
      stopPolling()
    }
  }, [router, startPolling, stopPolling])

  useEffect(() => {
    isVisibleRef.current = document.visibilityState === 'visible'

    if (document.visibilityState === 'visible') {
      startPolling()
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)

    return () => {
      stopPolling()
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [handleVisibilityChange, startPolling, stopPolling])

  return null
}
