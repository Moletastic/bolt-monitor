import { notFound } from 'next/navigation'
import { Pencil } from 'lucide-react'
import Link from 'next/link'
import type { ReactNode } from 'react'

import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { MonitorDetailActionsMenu } from '@/components/monitor-detail-actions-menu'
import {
  MobileMonitorIndicatorPicker,
  MonitorIndicatorCard,
} from '@/components/monitor-indicator-picker'
import { MonitorProtocolBadge } from '@/components/monitor-protocol-badge'
import { MonitorRunTimelineChart } from '@/components/monitor-run-timeline-chart'
import { QueryFeedbackBanner } from '@/components/query-feedback-banner'
import { StatusChip } from '@/components/status-chip'
import { SubmitButton } from '@/components/submit-button'
import { UnavailableCard } from '@/components/unavailable-card'
import { buttonVariants } from '@/components/ui/button'
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
} from '@/lib/api'
import { triggerManualRunAction } from '@/lib/actions'
import { buildMonitorChartPoints, buildMonitorIndicators } from '@/lib/monitor-detail-metrics'
import type { Monitor } from '@/lib/types'
import { formatDateTime, formatDuration, formatMonitorCadence } from '@/lib/utils'

type Runs = Awaited<ReturnType<typeof getMonitorRuns>>
type Incidents = Awaited<ReturnType<typeof getMonitorIncidents>>
type AuditEvents = Awaited<ReturnType<typeof getMonitorAudit>>

function RunsTab({ runs }: { runs: Runs }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent runs</CardTitle>
      </CardHeader>
      <CardContent>
        {runs.length === 0 ? (
          <EmptyState description="No run history returned yet." title="No runs yet" />
        ) : (
          <>
            <div className="grid gap-3 md:hidden">
              {runs.map((run) => (
                <div className="rounded-lg border border-border bg-surface-low p-4" key={run.runId}>
                  <div className="flex items-start justify-between gap-3">
                    <StatusChip status={run.outcome} />
                    <span className="font-mono text-xs text-muted-foreground">
                      {run.statusCode ?? 'n/a'}
                    </span>
                  </div>
                  <p className="mt-3 font-mono text-sm text-foreground">
                    {formatDateTime(run.startedAt)}
                  </p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {formatDuration(run.durationMs)} · {run.trigger}
                  </p>
                  {run.error && (
                    <p className="mt-2 break-words font-mono text-xs text-status-down">
                      {run.error}
                    </p>
                  )}
                </div>
              ))}
            </div>
            <div className="hidden overflow-x-auto md:block">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Started</TableHead>
                    <TableHead>Outcome</TableHead>
                    <TableHead>Duration</TableHead>
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
                      <TableCell>{run.trigger}</TableCell>
                      <TableCell>{run.statusCode ?? 'n/a'}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}

function IncidentsTab({ incidents }: { incidents: Incidents }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Incidents</CardTitle>
      </CardHeader>
      <CardContent>
        {incidents.length === 0 ? (
          <EmptyState description="No incidents recorded for this monitor." title="No incidents" />
        ) : (
          <>
            <div className="grid gap-3 md:hidden">
              {incidents.map((incident) => (
                <div
                  className="rounded-lg border border-border bg-surface-low p-4"
                  key={incident.incidentId}
                >
                  <div className="flex items-start justify-between gap-3">
                    <StatusChip status={incident.status} />
                    <span className="text-xs text-muted-foreground">{incident.origin ?? '—'}</span>
                  </div>
                  <p className="mt-3 font-medium text-foreground">{incident.summary}</p>
                  <p className="mt-1 font-mono text-xs text-muted-foreground">
                    {formatDateTime(incident.openedAt)}
                  </p>
                </div>
              ))}
            </div>
            <div className="hidden overflow-x-auto md:block">
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
          </>
        )}
      </CardContent>
    </Card>
  )
}

function AuditTab({ events }: { events: AuditEvents }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Audit log</CardTitle>
      </CardHeader>
      <CardContent>
        {events.length === 0 ? (
          <EmptyState description="No audit events recorded yet." title="No events" />
        ) : (
          <>
            <div className="grid gap-3 md:hidden">
              {events.map((event) => (
                <div
                  className="rounded-lg border border-border bg-surface-low p-4"
                  key={event.auditId}
                >
                  <p className="font-medium text-foreground">{event.eventType}</p>
                  <p className="mt-2 font-mono text-xs text-muted-foreground">
                    {formatDateTime(event.occurredAt)}
                  </p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {event.actor ?? '—'} · {event.origin ?? '—'}
                  </p>
                </div>
              ))}
            </div>
            <div className="hidden overflow-x-auto md:block">
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
          </>
        )}
      </CardContent>
    </Card>
  )
}

function getMonitorTarget(monitor: Monitor) {
  return monitor.http?.target ?? String(monitor.type ?? 'monitor').toUpperCase()
}

function getServiceMonitorCount(service: Awaited<ReturnType<typeof getService>>) {
  return service.monitorCount ?? service.monitors?.length ?? 0
}

function getStatusDotClass(status?: string) {
  switch (status?.toUpperCase()) {
    case 'UP':
      return 'bg-status-up shadow-[0_0_0_4px_hsl(var(--status-up)/0.12)]'
    case 'DOWN':
      return 'bg-status-down shadow-[0_0_0_4px_hsl(var(--status-down)/0.12)]'
    case 'DEGRADED':
    case 'RECOVERING':
      return 'bg-status-warn shadow-[0_0_0_4px_hsl(var(--status-warn)/0.12)]'
    default:
      return 'bg-status-unknown shadow-[0_0_0_4px_hsl(var(--status-unknown)/0.12)]'
  }
}

function MetricBadge({ children }: { children: ReactNode }) {
  return (
    <span className="inline-flex items-center rounded-md border border-border bg-surface-low px-2.5 py-1 text-xs font-semibold text-muted-foreground">
      {children}
    </span>
  )
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

  const [statusResult, runsResult, incidentsResult, eventsResult] = await Promise.all([
    loadSection(() => getMonitorStatus(serviceId, monitorId)),
    loadSection(() => getMonitorRuns(serviceId, monitorId)),
    loadSection(() => getMonitorIncidents(serviceId, monitorId)),
    loadSection(() => getMonitorAudit(serviceId, monitorId)),
  ])

  const monitorDeleteBlocked =
    service.lifecycleState === 'active' && getServiceMonitorCount(service) <= 1

  const status = statusResult.data
  const runs = runsResult.data ?? []
  const incidents = incidentsResult.data ?? []
  const events = eventsResult.data ?? []
  const indicators = buildMonitorIndicators(status, runs)
  const chartPoints = buildMonitorChartPoints(runs)
  const returnTo = `/services/${serviceId}/monitors/${monitor.monitorId}`

  return (
    <AppShell
      breadcrumbs={[
        { label: 'Services', href: '/services' },
        { label: service.name || 'Service', href: `/services/${serviceId}` },
        { label: monitor.name || 'Monitor' },
      ]}
      currentPath={`/services/${serviceId}/monitors/${monitorId}`}
    >
      <h1 className="sr-only">{monitor.name}</h1>
      <div className="grid gap-6">
        <Card>
          <CardContent className="space-y-5 pt-6">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div className="min-w-0 space-y-3">
                <div className="flex min-w-0 items-center gap-3">
                  <span
                    aria-label={`Status indicator for ${status?.currentStatus ?? 'UNKNOWN'}`}
                    className={`h-3 w-3 flex-shrink-0 rounded-full ${getStatusDotClass(status?.currentStatus)}`}
                    role="img"
                  />
                  <h2 className="min-w-0 truncate text-3xl font-semibold tracking-tight text-foreground">
                    {monitor.name}
                  </h2>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <MonitorProtocolBadge type={monitor.type} />
                  <span className="font-mono text-sm font-semibold text-foreground">
                    {monitor.http?.method ?? String(monitor.type ?? 'monitor').toUpperCase()}
                  </span>
                  <span className="break-all font-mono text-sm text-muted-foreground">
                    {getMonitorTarget(monitor)}
                  </span>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <MetricBadge>{formatMonitorCadence(monitor.intervalSeconds ?? 0)}</MetricBadge>
                  <MetricBadge>{formatDuration(monitor.http?.timeoutMs ?? 0)} timeout</MetricBadge>
                </div>
                {status?.currentStatus === 'DEGRADED' && (
                  <p className="text-sm text-status-warn">
                    {status.consecutiveFailures}/{monitor.failureThreshold} failures until incident
                  </p>
                )}
                {status?.currentStatus === 'RECOVERING' && (
                  <p className="text-sm text-status-warn">
                    {status.consecutiveSuccesses}/{monitor.recoveryThreshold} successes until
                    recovery
                  </p>
                )}
              </div>
              <div className="flex items-center justify-end gap-2 self-stretch lg:self-start">
                {monitor.enabled && status && (
                  <form action={triggerManualRunAction} className="min-w-0 flex-1 lg:flex-none">
                    <input name="serviceId" type="hidden" value={serviceId} />
                    <input name="monitorId" type="hidden" value={monitor.monitorId} />
                    <input name="returnTo" type="hidden" value={returnTo} />
                    <SubmitButton
                      className="w-full gap-2 border-primary text-primary hover:bg-primary/10 lg:w-auto"
                      iconName="play"
                      pendingLabel="Running..."
                      type="submit"
                      variant="outline"
                    >
                      Run now
                    </SubmitButton>
                  </form>
                )}
                <Link
                  aria-label="Edit monitor"
                  className={buttonVariants({
                    variant: 'outline',
                    className: 'gap-2 px-3 lg:px-4',
                  })}
                  href={`${returnTo}/edit`}
                >
                  <Pencil aria-hidden="true" className="h-4 w-4" />
                  <span className="hidden lg:inline">Edit</span>
                </Link>
                <MonitorDetailActionsMenu
                  deleteDisabled={monitorDeleteBlocked}
                  deleteDisabledReason="This is the only monitor on an active service. Disable or archive the service before permanently deleting this monitor."
                  enabled={monitor.enabled}
                  inMaintenance={status?.currentStatus === 'MAINTENANCE'}
                  monitorId={monitor.monitorId}
                  monitorName={monitor.name}
                  returnTo={returnTo}
                  serviceId={serviceId}
                />
              </div>
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
            {statusResult.error && (
              <UnavailableCard message={statusResult.error} title="Status unavailable" />
            )}
            {status?.lastError && (
              <div className="rounded-lg border border-status-down/30 bg-status-down/10 p-4">
                <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-status-down">
                  Latest error
                </p>
                <p className="mt-2 break-words font-mono text-sm text-status-down">
                  {status.lastError}
                </p>
              </div>
            )}
          </CardContent>
        </Card>

        <div className="hidden gap-4 md:grid md:grid-cols-2 xl:grid-cols-4">
          {indicators.map((indicator) => (
            <MonitorIndicatorCard indicator={indicator} key={indicator.key} />
          ))}
        </div>
        <MobileMonitorIndicatorPicker indicators={indicators} />

        <section>
          <Card>
            <CardHeader>
              <CardTitle>Run timeline</CardTitle>
            </CardHeader>
            <CardContent>
              {runsResult.data ? (
                chartPoints.length === 0 ? (
                  <EmptyState
                    description="No recent run datapoints returned yet."
                    title="No chart data"
                  />
                ) : (
                  <MonitorRunTimelineChart indicators={indicators} points={chartPoints} />
                )
              ) : (
                <UnavailableCard message={runsResult.error} title="Performance unavailable" />
              )}
            </CardContent>
          </Card>
        </section>

        <section className="grid gap-6">
          <div className="space-y-4">
            <Tabs
              basePath={returnTo}
              tabs={[
                {
                  iconName: 'history',
                  label: 'Runs',
                  href: `${returnTo}?tab=runs`,
                },
                {
                  iconName: 'incidents',
                  label: 'Incidents',
                  href: `${returnTo}?tab=incidents`,
                },
                {
                  iconName: 'audit',
                  label: 'Audit',
                  href: `${returnTo}?tab=audit`,
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
        </section>
      </div>
    </AppShell>
  )
}
