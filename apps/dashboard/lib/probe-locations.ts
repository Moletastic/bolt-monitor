import type { ProbeLocation } from '@/lib/types'

export type MonitorLocationField =
  | { kind: 'single-fixed'; location: ProbeLocation }
  | { kind: 'multi'; locations: ProbeLocation[] }

export function getMonitorLocationField(locations: ProbeLocation[]): MonitorLocationField {
  const enabled = locations.filter((location) => location.enabled)
  if (enabled.length === 1) {
    return { kind: 'single-fixed', location: enabled[0] }
  }
  return { kind: 'multi', locations: enabled }
}
