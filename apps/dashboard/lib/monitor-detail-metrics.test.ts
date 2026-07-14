import { describe, expect, it } from 'vitest'

import { buildMonitorChartPoints, buildMonitorIndicators } from '@/lib/monitor-detail-metrics'
import type { CheckRun, MonitorStatus } from '@/lib/types'

const status: MonitorStatus = {
  serviceId: 'svc-1',
  monitorId: 'mon-1',
  currentStatus: 'UP',
  consecutiveFailures: 0,
  consecutiveSuccesses: 3,
  lastCheckedAt: '2026-07-13T10:00:00Z',
  lastDurationMs: 120,
  lastOutcome: 'SUCCESS',
}

function run(overrides: Partial<CheckRun>): CheckRun {
  return {
    runId: 'run-1',
    type: 'http',
    trigger: 'scheduled',
    startedAt: '2026-07-13T10:00:00Z',
    finishedAt: '2026-07-13T10:00:01Z',
    durationMs: 100,
    outcome: 'SUCCESS',
    ...overrides,
  }
}

describe('monitor detail metrics', () => {
  it('derives monitor indicators from status and recent runs', () => {
    const indicators = buildMonitorIndicators(status, [
      run({ runId: 'run-1', durationMs: 100 }),
      run({ runId: 'run-2', durationMs: 250 }),
      run({ runId: 'run-3', durationMs: 500, outcome: 'FAILURE' }),
    ])

    expect(indicators).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ key: 'state', value: 'up', tone: 'success' }),
        expect.objectContaining({ key: 'uptime', value: '66.7%', detail: '2/3 runs passed' }),
        expect.objectContaining({ key: 'p99', value: '250 ms' }),
        expect.objectContaining({ key: 'errors', value: '33.3%', detail: '1 failed run' }),
      ])
    )
  })

  it('uses no-data values when runs or status are unavailable', () => {
    const indicators = buildMonitorIndicators(undefined, [])

    expect(indicators).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ key: 'state', value: 'Unknown', detail: 'Status unavailable' }),
        expect.objectContaining({ key: 'uptime', value: 'No data' }),
        expect.objectContaining({ key: 'p99', value: 'No data' }),
        expect.objectContaining({ key: 'errors', value: 'No data' }),
      ])
    )
  })

  it('tolerates partial persisted run and status payloads', () => {
    const partialStatus = { ...status, lastOutcome: undefined as unknown as string }
    const partialRun = run({ outcome: undefined as unknown as string, durationMs: Number.NaN })

    expect(() => buildMonitorIndicators(partialStatus, [partialRun])).not.toThrow()
    expect(() => buildMonitorChartPoints([partialRun])).not.toThrow()
    expect(buildMonitorChartPoints([partialRun])[0]).toEqual(
      expect.objectContaining({ durationMs: 0, outcome: 'unknown', success: false })
    )
  })

  it('maps latest recent runs into chronological chart datapoints with outcome context', () => {
    const points = buildMonitorChartPoints([
      run({
        runId: 'latest',
        startedAt: '2026-07-13T10:00:00Z',
        durationMs: 600,
        outcome: 'FAILURE',
        statusCode: 500,
        error: 'server error',
      }),
      run({ runId: 'old', startedAt: '2026-07-13T09:59:00Z', durationMs: 90 }),
      run({
        runId: 'older-than-window',
        startedAt: '2026-07-13T09:58:00Z',
        durationMs: 600,
        outcome: 'FAILURE',
        statusCode: 500,
        error: 'server error',
      }),
    ])

    expect(points).toEqual([
      expect.objectContaining({ runId: 'older-than-window', success: false }),
      expect.objectContaining({ runId: 'old', success: true }),
      expect.objectContaining({
        runId: 'latest',
        success: false,
        statusCode: 500,
        error: 'server error',
      }),
    ])
  })
})
