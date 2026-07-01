import Link from 'next/link'

import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { MonitorTrafficLights } from '@/components/monitor-traffic-light'
import { ServiceIcon } from '@/components/service-icon'
import { ServiceListStatusToast } from '@/components/service-list-status-toast'
import { StatusChip } from '@/components/status-chip'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ApiError, listServices } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

type ServiceSummary = Awaited<ReturnType<typeof listServices>>[number]

function formatCoverage(service: ServiceSummary) {
  const total = service.monitorCount ?? service.monitors?.length ?? 0
  const enabled = service.enabledMonitorCount

  if (enabled === undefined) {
    return `${total} monitor${total === 1 ? '' : 's'}`
  }

  return `${enabled}/${total} enabled`
}

export default async function ServicesPage({
  searchParams,
}: {
  searchParams: Promise<{ deletedService?: string }>
}) {
  const query = await searchParams

  try {
    const services = await listServices()

    if (services.length === 0) {
      return (
        <AppShell currentPath="/services">
          <EmptyState
            actionHref="/services/new"
            actionLabel="Create first service"
            description="No services configured yet. Create a draft service first, then add nested monitors from the service detail view."
            title="No services yet"
          />
        </AppShell>
      )
    }

    const activeCount = services.filter((service) => service.lifecycleState === 'active').length
    const draftCount = services.filter((service) => service.lifecycleState === 'draft').length
    const downCount = services.filter(
      (service) => service.rollupStatus?.toUpperCase() === 'DOWN'
    ).length

    return (
      <AppShell currentPath="/services">
        <ServiceListStatusToast services={services} />
        <div className="grid gap-6">
          <h1 className="sr-only">Services</h1>
          {query.deletedService ? (
            <p className="rounded-md border border-status-up/30 bg-status-up/10 px-3 py-2 text-sm text-status-up">
              Service permanently deleted.
            </p>
          ) : null}
          <section className="grid gap-4 xl:grid-cols-[2fr_1fr_1fr]">
            <Card className="overflow-hidden">
              <CardHeader>
                <CardTitle>Service operations</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <h2 className="text-3xl font-semibold tracking-tight text-foreground">
                    {services.length} tracked services
                  </h2>
                  <p className="mt-2 max-w-2xl text-sm text-muted-foreground">
                    Services now anchor the overview. Drill into a service to manage its metadata,
                    child monitors, recent runs, and lifecycle state.
                  </p>
                </div>
                <div className="grid gap-2 sm:grid-cols-3">
                  {services.slice(0, 12).map((service) => (
                    <Link
                      className="rounded-lg border border-border/80 bg-surface-low px-3 py-2 hover:border-primary/40"
                      href={`/services/${service.serviceId}`}
                      key={service.serviceId}
                    >
                      <div className="flex items-center gap-3">
                        <ServiceIcon size="sm" technologyKey={service.technologyKey} />
                        <div className="min-w-0">
                          <p className="truncate text-sm font-semibold text-foreground">
                            {service.name}
                          </p>
                          <p className="mt-1 text-xs uppercase tracking-[0.2em] text-muted-foreground">
                            {service.rollupStatus ?? service.lifecycleState}
                          </p>
                        </div>
                      </div>
                    </Link>
                  ))}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Active</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="font-mono text-3xl font-semibold text-foreground">{activeCount}</p>
                <p className="mt-2 text-sm text-muted-foreground">
                  Services currently in active lifecycle state.
                </p>
              </CardContent>
            </Card>
            <div className="grid gap-4">
              <Card>
                <CardHeader>
                  <CardTitle>Drafts</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="font-mono text-3xl font-semibold text-foreground">{draftCount}</p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    Services that still need monitor coverage or activation.
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>Currently down</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="font-mono text-3xl font-semibold text-status-down">{downCount}</p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    Based on derived rollup status from enabled child monitors.
                  </p>
                </CardContent>
              </Card>
            </div>
          </section>
          <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {services.map((service) => (
              <Card
                className="transition-colors hover:border-primary/40 hover:bg-surface-low/70"
                key={service.serviceId}
              >
                <Link
                  aria-label={`Open ${service.name}`}
                  className="block focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                  href={`/services/${service.serviceId}`}
                >
                  <CardContent className="space-y-4 pt-6">
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex items-start gap-3">
                        <ServiceIcon technologyKey={service.technologyKey} />
                        <div>
                          <p className="text-lg font-semibold text-foreground">{service.name}</p>
                          <p className="mt-1 text-xs uppercase tracking-[0.2em] text-muted-foreground">
                            {service.lifecycleState}
                          </p>
                        </div>
                      </div>
                      <StatusChip status={service.rollupStatus ?? service.lifecycleState} />
                    </div>
                    <MonitorTrafficLights monitors={service.monitors ?? []} />
                    <p className="min-h-10 text-sm text-muted-foreground">
                      {service.description || 'No service description yet.'}
                    </p>
                    <dl className="grid grid-cols-2 gap-3 text-sm">
                      <div>
                        <dt className="text-muted-foreground">Lifecycle</dt>
                        <dd className="mt-1 font-semibold text-foreground capitalize">
                          {service.lifecycleState}
                        </dd>
                      </div>
                      <div>
                        <dt className="text-muted-foreground">Technology</dt>
                        <dd className="mt-1 font-semibold text-foreground">
                          {service.technologyKey ?? 'None'}
                        </dd>
                      </div>
                      <div>
                        <dt className="text-muted-foreground">Coverage</dt>
                        <dd className="mt-1 font-semibold text-foreground">
                          {formatCoverage(service)}
                        </dd>
                      </div>
                      <div>
                        <dt className="text-muted-foreground">Updated</dt>
                        <dd className="mt-1 font-mono text-sm text-foreground">
                          {formatDateTime(service.updatedAt)}
                        </dd>
                      </div>
                    </dl>
                  </CardContent>
                </Link>
              </Card>
            ))}
          </section>
        </div>
      </AppShell>
    )
  } catch (error) {
    const message = error instanceof ApiError ? error.message : 'Unable to load service overview.'
    return (
      <AppShell currentPath="/services">
        <EmptyState
          actionHref="/services/new"
          actionLabel="Open create form"
          description={`${message} Check NEXT_PUBLIC_MONITOR_API_BASE_URL and local monitor API availability.`}
          title="Overview unavailable"
        />
      </AppShell>
    )
  }
}
