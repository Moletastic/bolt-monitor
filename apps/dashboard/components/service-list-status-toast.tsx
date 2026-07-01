'use client'

import { useEffect, useRef } from 'react'
import { toast } from '@/hooks/use-toast'

interface ServiceStatusToastProps {
  services: Array<{
    serviceId: string
    name: string
    rollupStatus?: string | null
  }>
}

const STORAGE_KEY = 'service-status-tracker'

interface ServiceStatusTracker {
  downServices: string[]
  notifiedDown: string[]
  notifiedUp: string[]
}

function defaultTracker(): ServiceStatusTracker {
  return { downServices: [], notifiedDown: [], notifiedUp: [] }
}

function stringArray(value: unknown): string[] {
  return Array.isArray(value)
    ? value.filter((item): item is string => typeof item === 'string')
    : []
}

function parseTracker(raw: string | null): ServiceStatusTracker {
  if (!raw) return defaultTracker()
  let parsed: unknown
  try {
    parsed = JSON.parse(raw) as unknown
  } catch {
    return defaultTracker()
  }
  if (!parsed || typeof parsed !== 'object') return defaultTracker()
  const record = parsed as Record<string, unknown>
  return {
    downServices: stringArray(record.downServices),
    notifiedDown: stringArray(record.notifiedDown),
    notifiedUp: stringArray(record.notifiedUp),
  }
}

export function ServiceListStatusToast({ services }: ServiceStatusToastProps) {
  const prevTrackerRef = useRef<ServiceStatusTracker | null>(null)

  useEffect(() => {
    const stored = sessionStorage.getItem(STORAGE_KEY)
    const tracker = parseTracker(stored)

    const currentDownServices = services
      .filter((s) => s.rollupStatus?.toUpperCase() === 'DOWN')
      .map((s) => s.serviceId)
    const currentUpServices = services
      .filter((s) => s.rollupStatus?.toUpperCase() === 'UP')
      .map((s) => s.serviceId)

    for (const service of services) {
      const wasDown = tracker.downServices.includes(service.serviceId)
      const isDown = currentDownServices.includes(service.serviceId)
      const isUp = currentUpServices.includes(service.serviceId)

      if (wasDown && isUp && !tracker.notifiedUp.includes(service.serviceId)) {
        toast({
          title: `${service.name} is UP again`,
          variant: 'default',
        })
        tracker.notifiedUp.push(service.serviceId)
        tracker.notifiedDown = tracker.notifiedDown.filter((id) => id !== service.serviceId)
      } else if (!wasDown && isDown && !tracker.notifiedDown.includes(service.serviceId)) {
        toast({
          title: `${service.name} is DOWN`,
          variant: 'destructive',
        })
        tracker.notifiedDown.push(service.serviceId)
      }
    }

    tracker.downServices = currentDownServices
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(tracker))
    prevTrackerRef.current = tracker
  }, [services])

  return null
}
