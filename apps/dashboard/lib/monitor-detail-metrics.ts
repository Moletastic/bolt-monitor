import type { CheckRun, MonitorStatus } from '@/lib/types'

export type MonitorIndicatorKey = 'state' | 'uptime' | 'p99' | 'errors'

export interface MonitorIndicator {
  key: MonitorIndicatorKey
  label: string
  value: string
  detail: string
  tone: 'neutral' | 'success' | 'warning' | 'danger'
}

export interface MonitorChartPoint {
  runId: string
  startedAt: string
  durationMs: number
  outcome: string
  success: boolean
  statusCode?: number
  error?: string
}

function isSuccessfulRun(run: CheckRun) {
  return String(run.outcome ?? '').toUpperCase() === 'SUCCESS'
}

function safeDurationMs(run: CheckRun) {
  return Number.isFinite(run.durationMs) ? run.durationMs : 0
}

function percentile(values: number[], percentileRank: number) {
  if (values.length === 0) {
    return undefined
  }

  const sorted = [...values].sort((left, right) => left - right)
  const index = Math.ceil((percentileRank / 100) * sorted.length) - 1
  return sorted[Math.max(0, Math.min(index, sorted.length - 1))]
}

function formatPercent(value: number) {
  return `${value.toFixed(value >= 99.95 || value === 0 || value === 100 ? 0 : 1)}%`
}

function pluralize(count: number, singular: string, plural = `${singular}s`) {
  return `${count} ${count === 1 ? singular : plural}`
}

export function buildMonitorChartPoints(runs: CheckRun[]): MonitorChartPoint[] {
  return runs
    .slice(0, 30)
    .reverse()
    .map((run) => ({
      runId: run.runId,
      startedAt: run.startedAt,
      durationMs: safeDurationMs(run),
      outcome: run.outcome || 'unknown',
      success: isSuccessfulRun(run),
      statusCode: run.statusCode,
      error: run.error,
    }))
}

export function buildMonitorIndicators(
  status: MonitorStatus | undefined,
  runs: CheckRun[]
): MonitorIndicator[] {
  const sampleCount = runs.length
  const successCount = runs.filter(isSuccessfulRun).length
  const failureCount = Math.max(0, sampleCount - successCount)
  const successfulDurations = runs.filter(isSuccessfulRun).map(safeDurationMs)
  const p99LatencyMs = percentile(successfulDurations, 99)
  const uptimePct = sampleCount === 0 ? undefined : (successCount / sampleCount) * 100
  const errorPct = sampleCount === 0 ? undefined : (failureCount / sampleCount) * 100

  return [
    {
      key: 'state',
      label: 'Current state',
      value: status?.currentStatus ? status.currentStatus.toLowerCase() : 'Unknown',
      detail: status
        ? `${(status.lastOutcome || 'unknown').toLowerCase()} · ${status.consecutiveFailures ?? 0} consecutive failures`
        : 'Status unavailable',
      tone:
        status?.currentStatus === 'UP'
          ? 'success'
          : status?.currentStatus === 'DOWN'
            ? 'danger'
            : status?.currentStatus === 'DEGRADED' || status?.currentStatus === 'RECOVERING'
              ? 'warning'
              : 'neutral',
    },
    {
      key: 'uptime',
      label: 'Recent uptime',
      value: uptimePct === undefined ? 'No data' : formatPercent(uptimePct),
      detail: sampleCount === 0 ? 'No recent runs' : `${successCount}/${sampleCount} runs passed`,
      tone: uptimePct === undefined ? 'neutral' : uptimePct >= 99 ? 'success' : 'warning',
    },
    {
      key: 'p99',
      label: 'P99 latency',
      value: p99LatencyMs === undefined ? 'No data' : `${p99LatencyMs} ms`,
      detail:
        successfulDurations.length === 0
          ? 'No successful runs'
          : `Recent tail from ${pluralize(successfulDurations.length, 'success', 'successes')}`,
      tone: p99LatencyMs === undefined ? 'neutral' : 'success',
    },
    {
      key: 'errors',
      label: 'Error rate',
      value: errorPct === undefined ? 'No data' : formatPercent(errorPct),
      detail: sampleCount === 0 ? 'No recent runs' : pluralize(failureCount, 'failed run'),
      tone: errorPct === undefined ? 'neutral' : errorPct === 0 ? 'success' : 'danger',
    },
  ]
}
