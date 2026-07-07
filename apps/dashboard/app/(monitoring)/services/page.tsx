import Link from 'next/link'

import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { FocusOnMount } from '@/components/focus-on-mount'
import { ServiceOverviewCard } from '@/components/service-overview-card'
import { ServiceIcon } from '@/components/service-icon'
import { ServiceListStatusToast } from '@/components/service-list-status-toast'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ApiError, listServices } from '@/lib/api'

function DeletedServiceFeedback({ active = true }: { active?: boolean }) {
  return (
    <FocusOnMount active={active}>
      <p
        className="rounded-md border border-status-up/30 bg-status-up/10 px-3 py-2 text-sm text-status-up"
        role="status"
      >
        Service permanently deleted.
      </p>
    </FocusOnMount>
  )
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
          <div className="grid gap-6">
            {query.deletedService ? <DeletedServiceFeedback /> : null}
            <EmptyState
              actionHref="/services/new"
              actionLabel="Create first service"
              description="No services configured yet. Create a draft service first, then add nested monitors from the service detail view."
              title="No services yet"
            />
          </div>
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
          {query.deletedService ? <DeletedServiceFeedback /> : null}
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
                        <ServiceIcon serviceCategory={service.serviceCategory} size="sm" />
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
          <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            {services.map((service) => (
              <ServiceOverviewCard key={service.serviceId} service={service} />
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
