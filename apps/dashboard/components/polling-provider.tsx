'use client'

import { startTransition, useEffect, useRef, useCallback } from 'react'
import { useRouter } from 'next/navigation'

interface PollingProviderProps {
  intervalMs?: number
}

export function PollingProvider({ intervalMs = 5000 }: PollingProviderProps) {
  const router = useRouter()
  const intervalRef = useRef<NodeJS.Timeout | null>(null)
  const isVisibleRef = useRef(true)
  const lastVisibilityStateRef = useRef<DocumentVisibilityState | null>(null)

  const refreshDashboard = useCallback(() => {
    startTransition(() => {
      router.refresh()
    })
  }, [router])

  const startPolling = useCallback(() => {
    if (intervalRef.current) return

    intervalRef.current = setInterval(() => {
      if (isVisibleRef.current) {
        refreshDashboard()
      }
    }, intervalMs)
  }, [refreshDashboard, intervalMs])

  const stopPolling = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
  }, [])

  const handleVisibilityChange = useCallback(() => {
    const nextVisibilityState = document.visibilityState
    if (lastVisibilityStateRef.current === nextVisibilityState) return
    lastVisibilityStateRef.current = nextVisibilityState

    if (nextVisibilityState === 'visible') {
      isVisibleRef.current = true
      refreshDashboard()
      startPolling()
    } else {
      isVisibleRef.current = false
      stopPolling()
    }
  }, [refreshDashboard, startPolling, stopPolling])

  useEffect(() => {
    lastVisibilityStateRef.current = document.visibilityState
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
