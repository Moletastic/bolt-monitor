import Link from 'next/link'
import { notFound } from 'next/navigation'
import { Activity, AlertOctagon, Archive, Clock, Pencil, Plus } from 'lucide-react'

import { AppShell } from '@/components/app-shell'
import { ArchiveServiceButton } from '@/components/archive-service-button'
import { DeleteServiceConfirmDialog } from '@/components/delete-service-confirm-dialog'
import { EmptyState } from '@/components/empty-state'
import { FocusOnMount } from '@/components/focus-on-mount'
import { MonitorTable } from '@/components/monitor-table'
import { RecentAlerts } from '@/components/recent-alerts'
import { ServiceIcon } from '@/components/service-icon'
import { StatusChip } from '@/components/status-chip'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ApiError, getService, listServiceIncidents } from '@/lib/api'
import { formatMetricDuration, formatRecentUptime } from '@/lib/service-card-metrics'
import type { Service, ServiceCardMetrics } from '@/lib/types'
import type { ServiceIconTone } from '@/components/service-icon'

const RECENT_ALERTS_LIMIT = 5

function serviceIconTone(
  service: Pick<Service, 'rollupStatus' | 'lifecycleState'>
): ServiceIconTone {
  if (service.lifecycleState === 'archived') {
    return 'unknown'
  }
  const rollup = (service.rollupStatus ?? '').toUpperCase()
  if (rollup === 'UP' || rollup === 'SUCCESS') {
    return 'up'
  }
  if (rollup === 'DOWN' || rollup === 'FAILED') {
    return 'down'
  }
  if (rollup === 'DEGRADED' || rollup === 'RECOVERING') {
    return 'degraded'
  }
  return 'unknown'
}

function formatErrorRate(metrics: ServiceCardMetrics | undefined): string {
  if (!metrics || metrics.state !== 'ready' || metrics.sampleCount === 0) {
    return 'No data'
  }
  const failed = metrics.sampleCount - metrics.successCount
  const rate = (failed / metrics.sampleCount) * 100
  if (rate === 0) {
    return '0%'
  }
  return `${rate.toFixed(2)}%`
}

function formatUptime(metrics: ServiceCardMetrics | undefined): string {
  if (!metrics || metrics.state !== 'ready' || metrics.recentUptimePct === undefined) {
    return 'No data'
  }
  return formatRecentUptime(metrics.recentUptimePct)
}

function formatP99(metrics: ServiceCardMetrics | undefined): string {
  if (!metrics || metrics.state !== 'ready' || metrics.p99LatencyMs === undefined) {
    return 'No data'
  }
  return formatMetricDuration(metrics.p99LatencyMs)
}

export default async function ServiceDetailPage({
  params,
  searchParams,
}: {
  params: Promise<{ serviceId: string }>
  searchParams: Promise<{
    created?: string
    updated?: string
    archived?: string
    error?: string
    deletedMonitor?: string
  }>
}) {
  const { serviceId } = await params
  const query = await searchParams

  try {
    const [service] = await Promise.all([getService(serviceId)])
    const monitors = service.monitors ?? []
    const isArchived = service.lifecycleState === 'archived'
    const serviceDeleteBlocked = service.lifecycleState === 'active'

    const recentAlerts = await listServiceIncidents(serviceId, RECENT_ALERTS_LIMIT).catch(() => [])

    const metrics = service.cardMetrics
    const showFeedbackBanner = Boolean(
      query.created || query.updated || query.archived || query.error || query.deletedMonitor
    )
    const feedbackMessage = query.error
      ? query.error
      : query.deletedMonitor
        ? 'Monitor permanently deleted.'
        : query.archived
          ? 'Service archived.'
          : query.created
            ? 'Service created.'
            : 'Service updated.'

    return (
      <AppShell
        breadcrumbs={[
          { label: 'Services', href: '/services' },
          { label: service.name || 'Service' },
        ]}
        currentPath={`/services/${serviceId}`}
      >
        <div className="grid gap-6">
          <header className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
            <div className="space-y-1">
              <h1 className="sr-only">{service.name}</h1>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Link
                aria-disabled={isArchived}
                className="inline-flex items-center gap-2 rounded-md border border-border bg-transparent px-3 py-2 text-sm font-semibold text-foreground transition-colors hover:bg-surface-low disabled:pointer-events-none disabled:opacity-50"
                href={isArchived ? '#' : `/services/${serviceId}/edit`}
                aria-label="Edit service"
                title={isArchived ? 'Archived services cannot be edited' : 'Edit service'}
              >
                <Pencil aria-hidden="true" className="h-4 w-4" />
                Edit service
              </Link>
              {!isArchived ? (
                <ArchiveServiceButton serviceId={service.serviceId} serviceName={service.name} />
              ) : (
                <span
                  aria-disabled="true"
                  className="inline-flex cursor-not-allowed items-center gap-2 rounded-md border border-border bg-transparent px-3 py-2 text-sm font-semibold text-muted-foreground opacity-50"
                  title="Archived services are already archived"
                >
                  <Archive aria-hidden="true" className="h-4 w-4" />
                  Archive service
                </span>
              )}
              {isArchived ? (
                <span
                  aria-disabled="true"
                  className="inline-flex cursor-not-allowed items-center gap-2 rounded-md border border-border bg-transparent px-3 py-2 text-sm font-semibold text-muted-foreground opacity-50"
                  title="Archived services cannot have new monitors"
                >
                  <Plus aria-hidden="true" className="h-4 w-4" />
                  Create monitor
                </span>
              ) : (
                <Link
                  aria-label="Create monitor"
                  className="inline-flex items-center gap-2 rounded-md border border-primary/40 bg-primary/10 px-3 py-2 text-sm font-semibold text-primary transition-colors hover:bg-primary/20"
                  href={`/services/${serviceId}/monitors/new`}
                >
                  <Plus aria-hidden="true" className="h-4 w-4" />
                  Create monitor
                </Link>
              )}
            </div>
          </header>

          <Card>
            <CardContent className="space-y-5 p-5 md:p-6">
              <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
                <div className="flex min-w-0 items-center gap-5">
                  <ServiceIcon
                    serviceCategory={service.serviceCategory}
                    size="xl"
                    tone={serviceIconTone(service)}
                  />
                  <div className="min-w-0 space-y-2">
                    <h2 className="text-2xl font-semibold tracking-tight text-foreground md:text-3xl">
                      {service.name}
                    </h2>
                    <p className="max-w-2xl text-sm italic text-muted-foreground">
                      {service.description || 'No description'}
                    </p>
                  </div>
                </div>
                <div className="flex flex-shrink-0 flex-wrap items-center gap-2">
                  <StatusChip status={service.rollupStatus ?? service.lifecycleState} />
                  {isArchived ? (
                    <span className="rounded-full border border-status-warn/30 bg-status-warn/10 px-2.5 py-1 text-[11px] font-bold uppercase tracking-[0.2em] text-status-warn">
                      Archived · Read-only
                    </span>
                  ) : null}
                </div>
              </div>
              <dl className="grid grid-cols-1 gap-4 border-t border-border pt-4 sm:grid-cols-3">
                <div className="rounded-lg border border-border bg-surface-low p-4">
                  <dt className="flex items-center gap-2 text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                    <Activity aria-hidden="true" className="h-3.5 w-3.5" />
                    Uptime
                  </dt>
                  <dd className="mt-2 font-mono text-xl font-semibold text-foreground">
                    {formatUptime(metrics)}
                  </dd>
                </div>
                <div className="rounded-lg border border-border bg-surface-low p-4">
                  <dt className="flex items-center gap-2 text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                    <Clock aria-hidden="true" className="h-3.5 w-3.5" />
                    P99 latency
                  </dt>
                  <dd className="mt-2 font-mono text-xl font-semibold text-foreground">
                    {formatP99(metrics)}
                  </dd>
                </div>
                <div className="rounded-lg border border-border bg-surface-low p-4">
                  <dt className="flex items-center gap-2 text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                    <AlertOctagon aria-hidden="true" className="h-3.5 w-3.5" />
                    Error rate
                  </dt>
                  <dd className="mt-2 font-mono text-xl font-semibold text-foreground">
                    {formatErrorRate(metrics)}
                  </dd>
                </div>
              </dl>
              {showFeedbackBanner ? (
                <FocusOnMount active={Boolean(query.deletedMonitor)}>
                  <p
                    className={`rounded-md border px-3 py-2 text-sm ${query.error ? 'border-status-down/30 bg-status-down/10 text-status-down' : 'border-status-up/30 bg-status-up/10 text-status-up'}`}
                    role={query.error ? 'alert' : 'status'}
                  >
                    {feedbackMessage}
                  </p>
                </FocusOnMount>
              ) : null}
            </CardContent>
          </Card>

          <div className="grid items-stretch gap-6 xl:grid-cols-10">
            <section className="order-2 min-w-0 xl:order-1 xl:col-span-7">
              <Card className="h-full">
                <CardHeader>
                  <CardTitle>Monitor overview</CardTitle>
                </CardHeader>
                <CardContent>
                  {monitors.length === 0 ? (
                    isArchived ? (
                      <EmptyState
                        description="This archived service has no monitors. New monitor creation is disabled while the service remains archived."
                        title="No monitors yet"
                      />
                    ) : (
                      <EmptyState
                        actionHref={`/services/${serviceId}/monitors/new`}
                        actionLabel="Create first monitor"
                        description="Draft services can exist without monitors. Add a nested monitor to activate real status, runs, and incident evidence for this service."
                        title="No monitors yet"
                      />
                    )
                  ) : (
                    <MonitorTable
                      monitors={monitors}
                      readOnly={isArchived}
                      returnTo={`/services/${serviceId}`}
                    />
                  )}
                </CardContent>
              </Card>
            </section>
            <section className="order-1 min-w-0 xl:order-2 xl:col-span-3">
              <RecentAlerts
                incidents={recentAlerts}
                limit={RECENT_ALERTS_LIMIT}
                serviceId={serviceId}
              />
            </section>
          </div>

          <div className="rounded-xl border border-destructive p-5 md:p-6">
            <div className="grid gap-5">
              <div className="space-y-2">
                <h2 className="flex items-center gap-2 text-base font-semibold tracking-tight text-foreground">
                  <AlertOctagon aria-hidden="true" className="h-5 w-5 text-destructive" />
                  Danger Zone
                </h2>
                <ul className="list-disc space-y-1 pl-5 text-sm text-muted-foreground">
                  <li>
                    Deleting a service is permanent and cannot be undone. All monitor configuration
                    for this service is removed from active management views. Use archive when you
                    need a reversible state change.
                  </li>
                  {serviceDeleteBlocked ? (
                    <li>
                      Active services cannot be deleted. Archive the service or remove active
                      monitor coverage first.
                    </li>
                  ) : null}
                </ul>
              </div>
              <div className="flex justify-end">
                <DeleteServiceConfirmDialog
                  disabled={serviceDeleteBlocked}
                  returnTo={`/services/${serviceId}`}
                  serviceId={service.serviceId}
                  serviceName={service.name}
                />
              </div>
            </div>
          </div>
        </div>
      </AppShell>
    )
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound()
    }

    return (
      <AppShell
        breadcrumbs={[{ label: 'Services', href: '/services' }, { label: 'Service' }]}
        currentPath={`/services/${serviceId}`}
      >
        <EmptyState
          description={`${error instanceof Error ? error.message : 'Unable to load service detail.'} Check local API connectivity and service identifier.`}
          title="Service detail unavailable"
        />
      </AppShell>
    )
  }
}
