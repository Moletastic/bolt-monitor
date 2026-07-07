import type { Service, ServiceCardMetrics } from '@/lib/types'
import { formatDuration } from '@/lib/utils'

export function formatMetricDuration(value?: number) {
  return value === undefined ? '--' : formatDuration(value)
}

export function formatRecentUptime(value?: number) {
  if (value === undefined) {
    return '--'
  }
  if (value === 100) {
    return '100%'
  }
  return `${value.toFixed(1)}%`
}

export function formatMonitorUpCoverage(metrics: ServiceCardMetrics | undefined, service: Service) {
  const total = metrics?.monitorCount ?? service.monitorCount ?? service.monitors?.length ?? 0
  if (total === 0) {
    return 'No monitors'
  }
  const up = metrics?.upMonitorCount
  if (up === undefined) {
    const enabled = service.enabledMonitorCount
    return enabled === undefined ? `${total} monitors` : `${enabled}/${total} enabled`
  }
  return `${up}/${total} monitors up`
}

export function serviceMetricStateLabel(metrics: ServiceCardMetrics | undefined, service: Service) {
  if (service.lifecycleState === 'draft' && (service.monitorCount ?? 0) === 0) {
    return 'Pending config'
  }
  if (!metrics) {
    return 'Metrics unavailable'
  }
  if (metrics.state === 'no_monitors') {
    return 'No monitors configured'
  }
  if (metrics.state === 'no_data') {
    return 'Waiting for data...'
  }
  return null
}
