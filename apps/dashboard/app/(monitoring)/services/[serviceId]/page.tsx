import Link from 'next/link'
import { notFound } from 'next/navigation'

import { AppShell } from '@/components/app-shell'
import { ArchiveServiceButton } from '@/components/archive-service-button'
import { DeleteResourceForm } from '@/components/delete-resource-form'
import { EmptyState } from '@/components/empty-state'
import { MonitorTable } from '@/components/monitor-table'
import { ServiceIcon } from '@/components/service-icon'
import { ServiceForm } from '@/components/service-form'
import { StatusChip } from '@/components/status-chip'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ApiError, getService, listEscalationPolicies } from '@/lib/api'
import { deleteServiceAction } from '@/lib/actions'
import { formatDateTime } from '@/lib/utils'

type ServiceDetail = Awaited<ReturnType<typeof getService>>

function getEnabledMonitorCount(service: ServiceDetail) {
  return (
    service.enabledMonitorCount ??
    service.monitors?.filter((monitor) => monitor.enabled).length ??
    0
  )
}

function getMonitorCount(service: ServiceDetail) {
  return service.monitorCount ?? service.monitors?.length ?? 0
}

function policyName(
  policyId: string,
  policies: Awaited<ReturnType<typeof listEscalationPolicies>>
) {
  const match = policies.find((policy) => policy.policyId === policyId)
  return match?.name ?? policyId
}

function describeBusinessHours(businessHours: NonNullable<ServiceDetail['businessHours']>) {
  const days = businessHours.daysOfWeek.map((day) => dayLabels[day] ?? String(day)).join(', ')
  return `${businessHours.timezone} · ${businessHours.startHour}:00–${businessHours.endHour}:00 · ${days}`
}

const dayLabels: Record<number, string> = {
  0: 'Sun',
  1: 'Mon',
  2: 'Tue',
  3: 'Wed',
  4: 'Thu',
  5: 'Fri',
  6: 'Sat',
}

function buildSetupSignals(service: ServiceDetail) {
  const total = getMonitorCount(service)
  const enabled = getEnabledMonitorCount(service)
  const signals: { label: string; tone: 'down' | 'warn' | 'info' }[] = []

  if ((service.rollupStatus ?? '').toUpperCase() === 'DOWN') {
    signals.push({ label: 'Service rollup is down', tone: 'down' })
  }
  if (service.lifecycleState === 'draft') {
    signals.push({ label: 'Service is still draft', tone: 'info' })
  }
  if (total === 0) {
    signals.push({ label: 'No monitor coverage yet', tone: 'warn' })
  } else if (enabled < total) {
    const disabled = total - enabled
    signals.push({
      label: `${disabled} disabled monitor${disabled === 1 ? '' : 's'}`,
      tone: 'warn',
    })
  }

  return signals
}

function SignalBadge({ label, tone }: { label: string; tone: 'down' | 'warn' | 'info' }) {
  const className = {
    down: 'border-status-down/30 bg-status-down/10 text-status-down',
    warn: 'border-status-warn/30 bg-status-warn/10 text-status-warn',
    info: 'border-primary/30 bg-primary/10 text-primary',
  }[tone]

  return (
    <span className={`rounded-full border px-2.5 py-1 text-xs font-semibold ${className}`}>
      {label}
    </span>
  )
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
    const [service, policies] = await Promise.all([
      getService(serviceId),
      listEscalationPolicies().catch(() => []),
    ])
    const monitors = service.monitors ?? []
    const monitorCount = getMonitorCount(service)
    const enabledMonitorCount = getEnabledMonitorCount(service)
    const setupSignals = buildSetupSignals(service)
    const isArchived = service.lifecycleState === 'archived'

    const serviceDeleteBlocked = service.lifecycleState === 'active'

    return (
      <AppShell currentPath={`/services/${serviceId}`}>
        <h1 className="sr-only">{service.name}</h1>
        <div className="grid gap-6">
          <section className="grid gap-6 xl:grid-cols-[1.3fr_0.7fr]">
            <Card>
              <CardHeader>
                <CardTitle>Service summary</CardTitle>
              </CardHeader>
              <CardContent className="space-y-5">
                <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
                  <div className="flex items-start gap-4">
                    <ServiceIcon size="lg" technologyKey={service.technologyKey} />
                    <div>
                      <p className="text-[11px] font-bold uppercase tracking-[0.28em] text-muted-foreground">
                        {service.technologyKey ?? 'service'} · {service.lifecycleState}
                      </p>
                      <h2 className="mt-2 text-3xl font-semibold tracking-tight text-foreground">
                        {service.name}
                      </h2>
                      <p className="mt-2 max-w-2xl text-sm text-muted-foreground">
                        {service.description ||
                          'No service description yet. Add nested monitors below to build real rollup coverage for this service.'}
                      </p>
                    </div>
                  </div>
                  <div className="flex flex-wrap justify-start gap-2 md:justify-end">
                    <StatusChip status={service.rollupStatus ?? service.lifecycleState} />
                    <span className="rounded-full border border-border bg-surface-low px-2.5 py-1 text-[11px] font-bold uppercase tracking-[0.2em] text-muted-foreground">
                      {service.lifecycleState}
                    </span>
                    {isArchived ? (
                      <span className="rounded-full border border-status-warn/30 bg-status-warn/10 px-2.5 py-1 text-[11px] font-bold uppercase tracking-[0.2em] text-status-warn">
                        Archived · Read-only
                      </span>
                    ) : null}
                  </div>
                </div>
                {(query.created ||
                  query.updated ||
                  query.archived ||
                  query.error ||
                  query.deletedMonitor) && (
                  <p
                    className={`rounded-md border px-3 py-2 text-sm ${query.error ? 'border-status-down/30 bg-status-down/10 text-status-down' : 'border-status-up/30 bg-status-up/10 text-status-up'}`}
                  >
                    {query.error ??
                      (query.deletedMonitor
                        ? 'Monitor permanently deleted.'
                        : query.archived
                          ? 'Service archived.'
                          : query.created
                            ? 'Service created.'
                            : 'Service updated.')}
                  </p>
                )}
                {setupSignals.length > 0 && (
                  <div className="flex flex-wrap gap-2">
                    {setupSignals.map((signal) => (
                      <SignalBadge key={signal.label} label={signal.label} tone={signal.tone} />
                    ))}
                  </div>
                )}
                <div className="grid gap-4 md:grid-cols-3">
                  <div className="rounded-lg border border-border bg-surface-low p-4">
                    <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                      Technology
                    </p>
                    <p className="mt-2 text-xl font-semibold text-foreground">
                      {service.technologyKey ?? 'None'}
                    </p>
                  </div>
                  <div className="rounded-lg border border-border bg-surface-low p-4">
                    <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                      Monitors
                    </p>
                    <p className="mt-2 text-xl font-semibold text-foreground">
                      {monitorCount} total
                    </p>
                  </div>
                  <div className="rounded-lg border border-border bg-surface-low p-4">
                    <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                      Coverage
                    </p>
                    <p className="mt-2 text-xl font-semibold text-foreground">
                      {enabledMonitorCount}/{monitorCount} enabled
                    </p>
                  </div>
                  <div className="rounded-lg border border-border bg-surface-low p-4">
                    <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                      Lifecycle
                    </p>
                    <p className="mt-2 text-xl font-semibold text-foreground capitalize">
                      {service.lifecycleState}
                    </p>
                  </div>
                  <div className="rounded-lg border border-border bg-surface-low p-4">
                    <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                      Updated
                    </p>
                    <p className="mt-2 font-mono text-xl font-semibold text-foreground">
                      {formatDateTime(service.updatedAt)}
                    </p>
                  </div>
                </div>
                <div className="rounded-lg border border-border bg-surface-low p-4">
                  <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                    Notification route
                  </p>
                  {service.escalationPolicyId ? (
                    <div className="mt-2 space-y-1 text-sm">
                      <Link
                        className="font-semibold text-primary hover:underline"
                        href={`/policies/${service.escalationPolicyId}`}
                      >
                        {policyName(service.escalationPolicyId, policies)}
                      </Link>
                      {service.businessHours ? (
                        <p className="text-xs text-muted-foreground">
                          {describeBusinessHours(service.businessHours)}
                        </p>
                      ) : (
                        <p className="text-xs text-muted-foreground">Uses off-hours path 24/7.</p>
                      )}
                    </div>
                  ) : (
                    <div className="mt-2 space-y-2 text-sm text-muted-foreground">
                      <p>No notification route assigned.</p>
                      <Link
                        className="inline-flex font-semibold text-primary hover:underline"
                        href="/policies/new"
                      >
                        Assign a route
                      </Link>
                    </div>
                  )}
                </div>
                <div className="flex flex-wrap justify-end gap-3">
                  {!isArchived ? (
                    <ArchiveServiceButton
                      serviceId={service.serviceId}
                      serviceName={service.name}
                    />
                  ) : null}
                  {isArchived ? (
                    <span
                      aria-disabled="true"
                      className="cursor-not-allowed rounded-md border border-border bg-surface-low px-3 py-2 text-sm font-semibold text-muted-foreground opacity-60"
                      title="Archived services cannot have new monitors."
                    >
                      Create monitor
                    </span>
                  ) : (
                    <Link
                      className="rounded-md border border-primary/40 bg-primary/10 px-3 py-2 text-sm font-semibold text-primary hover:bg-primary/20"
                      href={`/services/${serviceId}/monitors/new`}
                    >
                      Create monitor
                    </Link>
                  )}
                </div>
              </CardContent>
            </Card>
            <div className="grid gap-6">
              {isArchived ? (
                <Card>
                  <CardHeader>
                    <CardTitle>Service is archived</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-2 text-sm text-muted-foreground">
                    <p>
                      This service is in the <strong className="text-foreground">archived</strong>{' '}
                      lifecycle state. Editing is disabled while it remains archived.
                    </p>
                    <p>
                      To make changes again, recreate the service or restore it from archived state
                      in your source of truth.
                    </p>
                  </CardContent>
                </Card>
              ) : (
                <ServiceForm error={query.error} policies={policies} service={service} />
              )}
              <Card className="border-status-down/30">
                <CardHeader>
                  <CardTitle>Delete service</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <p className="text-sm text-muted-foreground">
                    Permanently deletes this service and its monitor configuration from active
                    management views. Use archive when you need a reversible state change.
                  </p>
                  {serviceDeleteBlocked ? (
                    <p className="rounded-md border border-status-warn/30 bg-status-warn/10 px-3 py-2 text-sm text-status-warn">
                      Active services cannot be deleted. Archive the service or remove active
                      monitor coverage first.
                    </p>
                  ) : null}
                  <DeleteResourceForm
                    action={deleteServiceAction}
                    confirmMessage={`Permanently delete ${service.name}? This cannot be undone.`}
                    disabled={serviceDeleteBlocked}
                    label="Delete service"
                  >
                    <input name="serviceId" type="hidden" value={service.serviceId} />
                    <input name="returnTo" type="hidden" value={`/services/${serviceId}`} />
                  </DeleteResourceForm>
                </CardContent>
              </Card>
            </div>
          </section>
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
        </div>
      </AppShell>
    )
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound()
    }

    return (
      <AppShell currentPath={`/services/${serviceId}`}>
        <EmptyState
          description={`${error instanceof Error ? error.message : 'Unable to load service detail.'} Check local API connectivity and service identifier.`}
          title="Service detail unavailable"
        />
      </AppShell>
    )
  }
}
