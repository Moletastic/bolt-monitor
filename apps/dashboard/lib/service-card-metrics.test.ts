import { describe, expect, it } from 'vitest'

import {
  formatMetricDuration,
  formatMonitorUpCoverage,
  formatRecentUptime,
  serviceMetricStateLabel,
} from '@/lib/service-card-metrics'
import type { Service } from '@/lib/types'

const baseService: Service = {
  tenantId: 'DEFAULT',
  serviceId: 'auth',
  name: 'Auth',
  lifecycleState: 'active',
  monitorCount: 2,
  enabledMonitorCount: 2,
  rollupStatus: 'up',
}

describe('service-card-metrics', () => {
  it('formats unavailable values honestly', () => {
    expect(formatMetricDuration(undefined)).toBe('--')
    expect(formatRecentUptime(undefined)).toBe('--')
  })

  it('formats recent uptime without implying a long window', () => {
    expect(formatRecentUptime(100)).toBe('100%')
    expect(formatRecentUptime(66.666)).toBe('66.7%')
  })

  it('uses monitor-up coverage from metrics', () => {
    expect(
      formatMonitorUpCoverage(
        {
          state: 'ready',
          sampleCount: 3,
          successCount: 2,
          monitorCount: 2,
          upMonitorCount: 1,
        },
        baseService
      )
    ).toBe('1/2 monitors up')
  })

  it('labels no-monitor and no-data states', () => {
    expect(
      serviceMetricStateLabel(
        {
          state: 'no_monitors',
          sampleCount: 0,
          successCount: 0,
          monitorCount: 0,
          upMonitorCount: 0,
        },
        { ...baseService, lifecycleState: 'draft', monitorCount: 0 }
      )
    ).toBe('Pending config')
    expect(
      serviceMetricStateLabel(
        { state: 'no_data', sampleCount: 0, successCount: 0, monitorCount: 1, upMonitorCount: 0 },
        baseService
      )
    ).toBe('Waiting for data...')
    expect(
      serviceMetricStateLabel(
        { state: 'ready', sampleCount: 3, successCount: 3, monitorCount: 1, upMonitorCount: 1 },
        baseService
      )
    ).toBeNull()
  })
})
