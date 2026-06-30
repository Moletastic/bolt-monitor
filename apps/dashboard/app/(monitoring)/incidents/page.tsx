import Link from 'next/link'

import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { StatusChip } from '@/components/status-chip'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ApiError } from '@/lib/api'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { listIncidents } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

const statusFilters = [
  { label: 'All', value: '' },
  { label: 'Open', value: 'open' },
  { label: 'Closed', value: 'closed' },
]

function getEmptyDescription(statusFilter: string) {
  if (statusFilter === 'open') {
    return 'No open incidents. Monitor execution has no active failures requiring operator action.'
  }
  if (statusFilter === 'closed') {
    return 'No closed incident history found for this filter yet.'
  }
  return 'No incidents recorded yet. Incidents are created by monitor execution when checks fail, not manually from this view.'
}

export default async function IncidentsPage({
  searchParams,
}: {
  searchParams: Promise<{ status?: string; error?: string }>
}) {
  const params = await searchParams
  const statusFilter = params.status ?? ''

  let incidents: Awaited<ReturnType<typeof listIncidents>> = []
  let loadError: string | undefined

  try {
    incidents = await listIncidents(statusFilter || undefined)
  } catch (error) {
    loadError = error instanceof ApiError ? error.message : 'Unable to load incidents.'
  }

  return (
    <AppShell currentPath="/incidents">
      <div className="grid gap-6">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Incidents</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Track and manage monitor incidents across all active monitors.
          </p>
        </div>
        <div className="flex gap-2">
          {statusFilters.map((filter) => (
            <Link
              key={filter.value}
              href={filter.value ? `/incidents?status=${filter.value}` : '/incidents'}
              className={`rounded-md border px-3 py-1.5 text-sm font-medium transition-colors ${
                statusFilter === filter.value
                  ? 'border-primary/40 bg-primary/10 text-primary'
                  : 'border-border bg-transparent text-muted-foreground hover:bg-surface-low hover:text-foreground'
              }`}
            >
              {filter.label}
            </Link>
          ))}
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Incident list</CardTitle>
          </CardHeader>
          <CardContent>
            {loadError ? (
              <EmptyState
                description={`${loadError} Check local API connectivity and incident API availability.`}
                title="Incidents unavailable"
              />
            ) : incidents.length === 0 ? (
              <EmptyState description={getEmptyDescription(statusFilter)} title="No incidents" />
            ) : (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Opened</TableHead>
                      <TableHead>Summary</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Origin</TableHead>
                      <TableHead className="text-right">Action</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {incidents.map((incident) => {
                      const isExhausted = incident.type === 'escalation.exhausted'
                      return (
                        <TableRow
                          key={incident.incidentId}
                          className={isExhausted ? 'bg-status-down/10' : undefined}
                        >
                          <TableCell className="font-mono text-xs">
                            {formatDateTime(incident.openedAt)}
                          </TableCell>
                          <TableCell className="max-w-xs truncate">
                            <div className="flex flex-col gap-1">
                              <span>{incident.summary}</span>
                              {isExhausted ? (
                                <span className="inline-flex w-fit items-center rounded-full border border-status-down/40 bg-status-down/15 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-status-down">
                                  Escalation exhausted
                                </span>
                              ) : null}
                            </div>
                          </TableCell>
                          <TableCell>
                            <StatusChip status={incident.status} />
                          </TableCell>
                          <TableCell>
                            <div className="flex flex-col gap-1 text-xs">
                              <span>{incident.origin ?? '—'}</span>
                              {isExhausted && incident.originalIncidentId ? (
                                <Link
                                  className="text-primary hover:underline"
                                  href={`/incidents/${incident.originalIncidentId}`}
                                >
                                  Original incident
                                </Link>
                              ) : null}
                            </div>
                          </TableCell>
                          <TableCell className="text-right">
                            <div className="flex justify-end gap-3">
                              <Link
                                className="text-primary hover:underline"
                                href={`/incidents/${incident.incidentId}`}
                              >
                                Open incident
                              </Link>
                              {incident.serviceId && (
                                <Link
                                  className="text-primary hover:underline"
                                  href={`/services/${incident.serviceId}/monitors/${incident.monitorId}`}
                                >
                                  Open monitor
                                </Link>
                              )}
                            </div>
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
      </div>
    </AppShell>
  )
}
