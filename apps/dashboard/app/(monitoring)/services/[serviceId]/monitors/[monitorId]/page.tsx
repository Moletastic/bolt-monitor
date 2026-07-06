import Link from 'next/link'
import { notFound } from 'next/navigation'

import { AppShell } from '@/components/app-shell'
import { DeleteResourceForm } from '@/components/delete-resource-form'
import { EmptyState } from '@/components/empty-state'
import { MonitorForm } from '@/components/monitor-form'
import { MonitorProtocolBadge } from '@/components/monitor-protocol-badge'
import { QueryFeedbackBanner } from '@/components/query-feedback-banner'
import { SamePageActionForm } from '@/components/same-page-action-form'
import { StatusChip } from '@/components/status-chip'
import { SubmitButton } from '@/components/submit-button'
import { UnavailableCard } from '@/components/unavailable-card'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs } from '@/components/ui/tabs'
import {
  ApiError,
  getMonitor,
  getMonitorAudit,
  getMonitorIncidents,
  getMonitorRuns,
  getMonitorStatus,
  getService,
  listProbeLocations,
} from '@/lib/api'
import {
  deleteMonitorAction,
  toggleMaintenanceModeAction,
  toggleMonitorStateAction,
  triggerManualRunAction,
} from '@/lib/actions'
import {
  formatDateTime,
  formatDuration,
  formatMonitorCadence,
  formatOutcome,
  formatProbeLocations,
} from '@/lib/utils'

function RunsTab({ runs }: { runs: Awaited<ReturnType<typeof getMonitorRuns>> }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent runs</CardTitle>
      </CardHeader>
      <CardContent>
        {runs.length === 0 ? (
          <EmptyState description="No run history returned yet." title="No runs yet" />
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Started</TableHead>
                  <TableHead>Outcome</TableHead>
                  <TableHead>Duration</TableHead>
                  <TableHead>Probe</TableHead>
                  <TableHead>Trigger</TableHead>
                  <TableHead>Status code</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {runs.map((run) => (
                  <TableRow key={run.runId}>
                    <TableCell className="font-mono text-xs">
                      {formatDateTime(run.startedAt)}
                    </TableCell>
                    <TableCell>
                      <StatusChip status={run.outcome} />
                    </TableCell>
                    <TableCell className="font-mono">{formatDuration(run.durationMs)}</TableCell>
                    <TableCell>{run.probeLocationId.toUpperCase()}</TableCell>
                    <TableCell>{run.trigger}</TableCell>
                    <TableCell>{run.statusCode ?? 'n/a'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function IncidentsTab({
  incidents,
}: {
  incidents: Awaited<ReturnType<typeof getMonitorIncidents>>
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Incidents</CardTitle>
      </CardHeader>
      <CardContent>
        {incidents.length === 0 ? (
          <EmptyState description="No incidents recorded for this monitor." title="No incidents" />
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Opened</TableHead>
                  <TableHead>Summary</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Origin</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {incidents.map((incident) => (
                  <TableRow key={incident.incidentId}>
                    <TableCell className="font-mono text-xs">
                      {formatDateTime(incident.openedAt)}
                    </TableCell>
                    <TableCell className="max-w-xs truncate">{incident.summary}</TableCell>
                    <TableCell>
                      <StatusChip status={incident.status} />
                    </TableCell>
                    <TableCell>{incident.origin ?? '—'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function AuditTab({ events }: { events: Awaited<ReturnType<typeof getMonitorAudit>> }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Audit log</CardTitle>
      </CardHeader>
      <CardContent>
        {events.length === 0 ? (
          <EmptyState description="No audit events recorded yet." title="No events" />
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>When</TableHead>
                  <TableHead>Event</TableHead>
                  <TableHead>Actor</TableHead>
                  <TableHead>Origin</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {events.map((event) => (
                  <TableRow key={event.auditId}>
                    <TableCell className="font-mono text-xs">
                      {formatDateTime(event.occurredAt)}
                    </TableCell>
                    <TableCell className="font-medium">{event.eventType}</TableCell>
                    <TableCell>{event.actor ?? '—'}</TableCell>
                    <TableCell>{event.origin ?? '—'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function getMonitorTarget(monitor: Awaited<ReturnType<typeof getMonitor>>) {
  return monitor.http?.target ?? formatProbeLocations(monitor.probeLocations)
}

function getServiceMonitorCount(service: Awaited<ReturnType<typeof getService>>) {
  return service.monitorCount ?? service.monitors?.length ?? 0
}

type LoadResult<T> = { data: T; error?: never } | { data?: never; error: string }

async function loadSection<T>(loader: () => Promise<T>): Promise<LoadResult<T>> {
  try {
    return { data: await loader() }
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Unable to load section data.' }
  }
}

export default async function ServiceMonitorDetailPage({
  params,
  searchParams,
}: {
  params: Promise<{ serviceId: string; monitorId: string }>
  searchParams: Promise<{
    tab?: string
    created?: string
    updated?: string
    error?: string
    run?: string
  }>
}) {
  const { serviceId, monitorId } = await params
  const query = await searchParams
  const activeTab = query.tab ?? 'runs'

  let service
  let monitor
  try {
    const result = await Promise.all([getService(serviceId), getMonitor(serviceId, monitorId)])
    service = result[0]
    monitor = result[1]
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound()
    }
    throw error
  }

  const [statusResult, runsResult, incidentsResult, eventsResult, probeLocationsResult] =
    await Promise.all([
      loadSection(() => getMonitorStatus(serviceId, monitorId)),
      loadSection(() => getMonitorRuns(serviceId, monitorId)),
      loadSection(() => getMonitorIncidents(serviceId, monitorId)),
      loadSection(() => getMonitorAudit(serviceId, monitorId)),
      loadSection(() => listProbeLocations()),
    ])

  const monitorDeleteBlocked =
    service.lifecycleState === 'active' && getServiceMonitorCount(service) <= 1

  const status = statusResult.data
  const runs = runsResult.data ?? []
  const incidents = incidentsResult.data ?? []
  const events = eventsResult.data ?? []
  const probeLocations = probeLocationsResult.data ?? []

  return (
    <AppShell currentPath={`/services/${serviceId}/monitors/${monitorId}`}>
      <h1 className="sr-only">{monitor.name}</h1>
      <div className="grid gap-6">
        <section className="grid gap-6 xl:grid-cols-[1.3fr_0.7fr]">
          <Card>
            <CardHeader>
              <CardTitle>Current status</CardTitle>
            </CardHeader>
            <CardContent className="space-y-5">
              {status ? (
                <>
                  <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
                    <div className="flex items-start gap-4">
                      <MonitorProtocolBadge type={monitor.type} />
                      <div>
                        <p className="text-[11px] font-bold uppercase tracking-[0.28em] text-muted-foreground">
                          {monitor.type.toUpperCase()} monitor
                        </p>
                        <h2 className="mt-2 text-3xl font-semibold tracking-tight text-foreground">
                          {monitor.name}
                        </h2>
                        <p className="mt-2 max-w-2xl break-all font-mono text-sm text-muted-foreground">
                          {monitor.http?.method ?? monitor.type.toUpperCase()}{' '}
                          {getMonitorTarget(monitor)}
                        </p>
                        <Link
                          className="mt-3 inline-block text-sm font-medium text-primary hover:underline"
                          href={`/services/${serviceId}`}
                        >
                          Back to service
                        </Link>
                      </div>
                    </div>
                    <StatusChip status={status.currentStatus} />
                    {status.currentStatus === 'DEGRADED' && (
                      <p className="text-sm text-status-warn">
                        {status.consecutiveFailures}/{monitor.failureThreshold} failures until
                        incident
                      </p>
                    )}
                    {status.currentStatus === 'RECOVERING' && (
                      <p className="text-sm text-status-warn">
                        {status.consecutiveSuccesses}/{monitor.recoveryThreshold} successes until
                        recovery
                      </p>
                    )}
                  </div>
                  {(query.created || query.updated || query.error || query.run) && (
                    <QueryFeedbackBanner
                      message={
                        query.error ||
                        (query.run
                          ? 'Manual run triggered.'
                          : query.created
                            ? 'Monitor created.'
                            : 'Monitor updated.')
                      }
                      tone={query.error ? 'error' : 'success'}
                    />
                  )}
                  <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                    <div className="rounded-lg border border-border bg-surface-low p-4">
                      <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                        Last outcome
                      </p>
                      <p className="mt-2 text-xl font-semibold text-foreground">
                        {formatOutcome(status.lastOutcome)}
                      </p>
                    </div>
                    <div className="rounded-lg border border-border bg-surface-low p-4">
                      <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                        Last check
                      </p>
                      <p className="mt-2 font-mono text-xl font-semibold text-foreground">
                        {formatDateTime(status.lastCheckedAt)}
                      </p>
                    </div>
                    <div className="rounded-lg border border-border bg-surface-low p-4">
                      <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                        Duration
                      </p>
                      <p className="mt-2 font-mono text-xl font-semibold text-foreground">
                        {formatDuration(status.lastDurationMs)}
                      </p>
                    </div>
                    <div className="rounded-lg border border-border bg-surface-low p-4">
                      <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                        Probe
                      </p>
                      <p className="mt-2 text-xl font-semibold text-foreground">
                        {status.lastProbeLocationId ?? formatProbeLocations(monitor.probeLocations)}
                      </p>
                    </div>
                    <div className="rounded-lg border border-border bg-surface-low p-4">
                      <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                        Cadence
                      </p>
                      <p className="mt-2 text-xl font-semibold text-foreground">
                        {formatMonitorCadence(monitor.intervalSeconds)}
                      </p>
                    </div>
                    <div className="rounded-lg border border-border bg-surface-low p-4">
                      <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                        Enabled
                      </p>
                      <p className="mt-2 text-xl font-semibold text-foreground">
                        {monitor.enabled ? 'Yes' : 'No'}
                      </p>
                    </div>
                  </div>
                  {status.lastError && (
                    <div className="rounded-lg border border-status-down/30 bg-status-down/10 p-4">
                      <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-status-down">
                        Latest error
                      </p>
                      <p className="mt-2 break-words font-mono text-sm text-status-down">
                        {status.lastError}
                      </p>
                    </div>
                  )}
                </>
              ) : (
                <UnavailableCard message={statusResult.error} title="Status unavailable" />
              )}
              <div className="flex justify-end gap-3">
                {monitor.enabled && status && (
                  <form action={triggerManualRunAction}>
                    <input name="serviceId" type="hidden" value={serviceId} />
                    <input name="monitorId" type="hidden" value={monitor.monitorId} />
                    <input
                      name="returnTo"
                      type="hidden"
                      value={`/services/${serviceId}/monitors/${monitor.monitorId}`}
                    />
                    <SubmitButton type="submit" variant="secondary">
                      Run now
                    </SubmitButton>
                  </form>
                )}
                <SamePageActionForm
                  action={toggleMonitorStateAction}
                  buttonLabel={monitor.enabled ? 'Disable monitor' : 'Enable monitor'}
                  pendingLabel={monitor.enabled ? 'Disabling monitor...' : 'Enabling monitor...'}
                  variant={monitor.enabled ? 'outline' : 'default'}
                >
                  <input name="serviceId" type="hidden" value={serviceId} />
                  <input name="monitorId" type="hidden" value={monitor.monitorId} />
                  <input name="enabled" type="hidden" value={monitor.enabled ? 'false' : 'true'} />
                  <input
                    name="returnTo"
                    type="hidden"
                    value={`/services/${serviceId}/monitors/${monitor.monitorId}`}
                  />
                </SamePageActionForm>
                <form action={toggleMaintenanceModeAction}>
                  <input name="serviceId" type="hidden" value={serviceId} />
                  <input name="monitorId" type="hidden" value={monitor.monitorId} />
                  <input
                    name="enabled"
                    type="hidden"
                    value={status?.currentStatus === 'MAINTENANCE' ? 'false' : 'true'}
                  />
                  <input
                    name="returnTo"
                    type="hidden"
                    value={`/services/${serviceId}/monitors/${monitor.monitorId}`}
                  />
                  <SubmitButton variant="outline">
                    {status?.currentStatus === 'MAINTENANCE'
                      ? 'Exit maintenance'
                      : 'Enter maintenance'}
                  </SubmitButton>
                </form>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Configuration</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="rounded-lg border border-border bg-surface-low p-4">
                <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                  Type
                </p>
                <p className="mt-2 font-semibold text-foreground">{monitor.type.toUpperCase()}</p>
              </div>
              <div className="rounded-lg border border-border bg-surface-low p-4">
                <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                  Target
                </p>
                <p className="mt-2 break-all font-mono text-foreground">
                  {getMonitorTarget(monitor)}
                </p>
              </div>
              <div className="rounded-lg border border-border bg-surface-low p-4">
                <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                  Interval / Timeout
                </p>
                <p className="mt-2 text-foreground">
                  {formatMonitorCadence(monitor.intervalSeconds)} · {monitor.http?.timeoutMs ?? 0}
                  ms timeout
                </p>
              </div>
            </CardContent>
          </Card>
        </section>
        <section className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
          <div className="space-y-4">
            <Tabs
              basePath={`/services/${serviceId}/monitors/${monitorId}`}
              tabs={[
                { label: 'Runs', href: `/services/${serviceId}/monitors/${monitorId}?tab=runs` },
                {
                  label: 'Incidents',
                  href: `/services/${serviceId}/monitors/${monitorId}?tab=incidents`,
                },
                {
                  label: 'Audit',
                  href: `/services/${serviceId}/monitors/${monitorId}?tab=audit`,
                },
              ]}
            />
            {activeTab === 'runs' &&
              (runsResult.data ? (
                <RunsTab runs={runs} />
              ) : (
                <UnavailableCard message={runsResult.error} title="Recent runs unavailable" />
              ))}
            {activeTab === 'incidents' &&
              (incidentsResult.data ? (
                <IncidentsTab incidents={incidents} />
              ) : (
                <UnavailableCard message={incidentsResult.error} title="Incidents unavailable" />
              ))}
            {activeTab === 'audit' &&
              (eventsResult.data ? (
                <AuditTab events={events} />
              ) : (
                <UnavailableCard message={eventsResult.error} title="Audit log unavailable" />
              ))}
          </div>
          <div className="space-y-6">
            <MonitorForm
              error={query.error}
              locations={probeLocations}
              monitor={{ ...monitor, status }}
              serviceId={serviceId}
            />
            <Card className="border-status-down/30">
              <CardHeader>
                <CardTitle>Delete monitor</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <p className="text-sm text-muted-foreground">
                  Permanently deletes this monitor configuration from active management views. Use
                  disable when you only need to pause checks and preserve the monitor.
                </p>
                {monitorDeleteBlocked ? (
                  <p className="rounded-md border border-status-warn/30 bg-status-warn/10 px-3 py-2 text-sm text-status-warn">
                    This is the only monitor on an active service. Disable or archive the service
                    before permanently deleting this monitor.
                  </p>
                ) : null}
                <DeleteResourceForm
                  action={deleteMonitorAction}
                  confirmMessage={`Permanently delete ${monitor.name}? This cannot be undone.`}
                  disabled={monitorDeleteBlocked}
                  label="Delete monitor"
                >
                  <input name="serviceId" type="hidden" value={serviceId} />
                  <input name="monitorId" type="hidden" value={monitor.monitorId} />
                  <input
                    name="returnTo"
                    type="hidden"
                    value={`/services/${serviceId}/monitors/${monitor.monitorId}`}
                  />
                </DeleteResourceForm>
              </CardContent>
            </Card>
          </div>
        </section>
      </div>
    </AppShell>
  )
}
