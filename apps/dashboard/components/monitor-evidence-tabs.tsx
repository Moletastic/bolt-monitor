'use client'

import { startTransition, useState } from 'react'

import {
  loadMonitorHistoryPageAction,
  type MonitorHistoryActionData,
  type MonitorHistoryKind,
} from '@/lib/actions'
import { actionErrorMessage, type ActionState } from '@/lib/action-state'
import type { AuditEvent, CheckRun, Incident } from '@/lib/types'
import { formatDateTime, formatDuration } from '@/lib/utils'
import { EmptyState } from '@/components/empty-state'
import { StatusChip } from '@/components/status-chip'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

type Page<T> = { items: T[]; nextCursor?: string; loaded: boolean; error?: string }
type Pages = { runs: Page<CheckRun>; incidents: Page<Incident>; audit: Page<AuditEvent> }

export function MonitorEvidenceTabs({
  activeTab,
  initialRuns,
  initialRunsCursor,
  serviceId,
  monitorId,
}: {
  activeTab: string
  initialRuns: CheckRun[]
  initialRunsCursor?: string
  serviceId: string
  monitorId: string
}) {
  const [active, setActive] = useState<MonitorHistoryKind>(
    activeTab === 'incidents' || activeTab === 'audit' ? activeTab : 'runs'
  )
  const [pending, setPending] = useState(false)
  const [pages, setPages] = useState<Pages>({
    runs: { items: initialRuns, nextCursor: initialRunsCursor, loaded: true },
    incidents: { items: [], loaded: false },
    audit: { items: [], loaded: false },
  })

  function request(kind: MonitorHistoryKind, cursor?: string) {
    if (pending) return
    setPending(true)
    const formData = new FormData()
    formData.set('kind', kind)
    formData.set('serviceId', serviceId)
    formData.set('monitorId', monitorId)
    if (cursor) formData.set('cursor', cursor)
    startTransition(async () => {
      const result = await loadMonitorHistoryPageAction(
        { status: 'idle' } as ActionState<MonitorHistoryActionData>,
        formData
      )
      setPending(false)
      if (result.status === 'error') {
        setPages((current) => ({
          ...current,
          [kind]: { ...current[kind], error: actionErrorMessage(result.error) },
        }))
        return
      }
      if (result.status !== 'success' || result.data.kind !== kind) return
      setPages((current) => ({
        ...current,
        [kind]: {
          items: cursor
            ? [...current[kind].items, ...result.data.page.items]
            : result.data.page.items,
          nextCursor: result.data.page.nextCursor,
          loaded: true,
        },
      }))
    })
  }

  function select(kind: MonitorHistoryKind) {
    setActive(kind)
    if (!pages[kind].loaded) request(kind)
  }

  const page = pages[active]
  return (
    <section className="space-y-4">
      <div aria-label="Monitor history" className="flex gap-2" role="tablist">
        {(['runs', 'incidents', 'audit'] as const).map((kind) => (
          <Button
            aria-selected={active === kind}
            key={kind}
            onClick={() => select(kind)}
            role="tab"
            size="sm"
            type="button"
            variant={active === kind ? 'default' : 'outline'}
          >
            {kind === 'runs' ? 'Runs' : kind === 'incidents' ? 'Incidents' : 'Audit'}
          </Button>
        ))}
      </div>
      <Card role="tabpanel">
        <CardHeader>
          <CardTitle>
            {active === 'runs' ? 'Recent runs' : active === 'incidents' ? 'Incidents' : 'Audit log'}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {!page.loaded || (pending && page.items.length === 0) ? (
            <p className="text-sm text-muted-foreground">Loading history...</p>
          ) : page.items.length === 0 ? (
            <EmptyState description="No history returned yet." title="No history" />
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    {active === 'runs' ? (
                      <>
                        <TableHead>Started</TableHead>
                        <TableHead>Outcome</TableHead>
                        <TableHead>Duration</TableHead>
                      </>
                    ) : active === 'incidents' ? (
                      <>
                        <TableHead>Opened</TableHead>
                        <TableHead>Summary</TableHead>
                        <TableHead>Status</TableHead>
                      </>
                    ) : (
                      <>
                        <TableHead>When</TableHead>
                        <TableHead>Event</TableHead>
                        <TableHead>Origin</TableHead>
                      </>
                    )}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {active === 'runs' &&
                    pages.runs.items.map((run) => (
                      <TableRow key={run.runId}>
                        <TableCell className="font-mono text-xs">
                          {formatDateTime(run.startedAt)}
                        </TableCell>
                        <TableCell>
                          <StatusChip status={run.outcome} />
                        </TableCell>
                        <TableCell>{formatDuration(run.durationMs)}</TableCell>
                      </TableRow>
                    ))}
                  {active === 'incidents' &&
                    pages.incidents.items.map((incident) => (
                      <TableRow key={incident.incidentId}>
                        <TableCell className="font-mono text-xs">
                          {formatDateTime(incident.openedAt)}
                        </TableCell>
                        <TableCell>{incident.summary}</TableCell>
                        <TableCell>
                          <StatusChip status={incident.status} />
                        </TableCell>
                      </TableRow>
                    ))}
                  {active === 'audit' &&
                    pages.audit.items.map((event) => (
                      <TableRow key={event.auditId}>
                        <TableCell className="font-mono text-xs">
                          {formatDateTime(event.occurredAt)}
                        </TableCell>
                        <TableCell>{event.eventType}</TableCell>
                        <TableCell>{event.origin ?? '—'}</TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </div>
          )}
          {page.error && <p className="mt-4 text-sm text-status-down">{page.error}</p>}
          {(page.nextCursor || page.error) && (
            <Button
              className="mt-4 w-full"
              disabled={pending}
              onClick={() => request(active, page.nextCursor)}
              type="button"
              variant="outline"
            >
              {pending ? 'Loading...' : page.error ? 'Try again' : 'Load more'}
            </Button>
          )}
        </CardContent>
      </Card>
    </section>
  )
}
