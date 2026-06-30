'use client'

import { useEffect, useState } from 'react'
import { differenceInMilliseconds, isValid, parseISO } from 'date-fns'

import { EmptyState } from '@/components/empty-state'
import { StatusChip } from '@/components/status-chip'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { listMonitorRuns } from '@/lib/api'
import type { CheckRun } from '@/lib/types'
import { formatDateTime, formatDuration } from '@/lib/utils'

interface AlertHistoryTabProps {
  serviceId: string
  monitorId: string
  openedAt: string
  acknowledgedAt?: string
  resolvedAt?: string
}

type TransitionLabel = 'Opened' | 'Acknowledged' | 'Resolved'

function getTransitionRunMap(
  runs: CheckRun[],
  timestamps: Array<{ label: TransitionLabel; timestamp?: string }>
) {
  const labelsByRunId = new Map<string, TransitionLabel[]>()

  for (const entry of timestamps) {
    if (!entry.timestamp || runs.length === 0) {
      continue
    }
    const target = parseISO(entry.timestamp)
    if (!isValid(target)) {
      continue
    }
    let nearestRun: CheckRun | null = null
    let nearestDistance = Number.POSITIVE_INFINITY
    for (const run of runs) {
      const runTime = parseISO(run.finishedAt)
      if (!isValid(runTime)) {
        continue
      }
      const distance = Math.abs(differenceInMilliseconds(runTime, target))
      if (distance < nearestDistance) {
        nearestDistance = distance
        nearestRun = run
      }
    }
    if (!nearestRun) {
      continue
    }
    const current = labelsByRunId.get(nearestRun.runId) ?? []
    current.push(entry.label)
    labelsByRunId.set(nearestRun.runId, current)
  }

  return labelsByRunId
}

export function AlertHistoryTab({
  serviceId,
  monitorId,
  openedAt,
  acknowledgedAt,
  resolvedAt,
}: AlertHistoryTabProps) {
  const [runs, setRuns] = useState<CheckRun[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function loadRuns() {
      setLoading(true)
      setError(null)
      try {
        const next = await listMonitorRuns(serviceId, monitorId)
        if (cancelled) {
          return
        }
        const filtered = next
          .filter((run) => {
            if (run.finishedAt < openedAt) {
              return false
            }
            if (resolvedAt && run.finishedAt > resolvedAt) {
              return false
            }
            return true
          })
          .sort((a, b) => a.finishedAt.localeCompare(b.finishedAt))
        setRuns(filtered)
      } catch (err) {
        if (cancelled) {
          return
        }
        setError(err instanceof Error ? err.message : 'Unable to load alert history.')
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void loadRuns()

    return () => {
      cancelled = true
    }
  }, [monitorId, openedAt, resolvedAt, serviceId])

  const transitionRuns = getTransitionRunMap(runs, [
    { label: 'Opened', timestamp: openedAt },
    { label: 'Acknowledged', timestamp: acknowledgedAt },
    { label: 'Resolved', timestamp: resolvedAt },
  ])

  return (
    <Card>
      <CardHeader>
        <CardTitle>Alert History</CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <p className="text-sm text-muted-foreground">Loading alert history…</p>
        ) : error ? (
          <div className="rounded-md border border-status-down/30 bg-status-down/10 px-3 py-2 text-sm text-status-down">
            {error}
          </div>
        ) : runs.length === 0 ? (
          <EmptyState description="No monitor runs overlap this incident window." title="No runs" />
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Finished</TableHead>
                  <TableHead>Outcome</TableHead>
                  <TableHead>Duration</TableHead>
                  <TableHead>Probe</TableHead>
                  <TableHead>Trigger</TableHead>
                  <TableHead>Incident markers</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {runs.map((run) => {
                  const markers = transitionRuns.get(run.runId) ?? []
                  return (
                    <TableRow key={run.runId}>
                      <TableCell className="font-mono text-xs">
                        {formatDateTime(run.finishedAt)}
                      </TableCell>
                      <TableCell>
                        <StatusChip status={run.outcome} />
                      </TableCell>
                      <TableCell className="font-mono">{formatDuration(run.durationMs)}</TableCell>
                      <TableCell>{run.probeLocationId.toUpperCase()}</TableCell>
                      <TableCell>{run.trigger}</TableCell>
                      <TableCell>
                        {markers.length === 0 ? (
                          '—'
                        ) : (
                          <div className="flex flex-wrap gap-2">
                            {markers.map((marker) => (
                              <span
                                key={`${run.runId}-${marker}`}
                                className="rounded-full border border-border bg-background px-2 py-0.5 text-xs text-foreground"
                              >
                                {marker}
                              </span>
                            ))}
                          </div>
                        )}
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
