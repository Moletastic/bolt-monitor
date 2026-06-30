import { notFound } from 'next/navigation'

import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { MonitorForm } from '@/components/monitor-form'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ApiError, getService, listProbeLocations } from '@/lib/api'
import { getMonitorLocationField } from '@/lib/probe-locations'

export default async function NewServiceMonitorPage({
  params,
  searchParams,
}: {
  params: Promise<{ serviceId: string }>
  searchParams: Promise<{ error?: string }>
}) {
  const { serviceId } = await params
  const query = await searchParams

  try {
    const [service, locations] = await Promise.all([getService(serviceId), listProbeLocations()])
    const field = getMonitorLocationField(locations)

    return (
      <AppShell currentPath={`/services/${serviceId}/monitors/new`}>
        <div className="grid gap-6 xl:grid-cols-[1.7fr_1fr]">
          <MonitorForm error={query.error} locations={locations} serviceId={service.serviceId} />
          <Card>
            <CardHeader>
              <CardTitle>Create flow notes</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4 text-sm text-muted-foreground">
              <p>
                New monitor will be created under {service.name} using nested service-monitor APIs.
              </p>
              <p>
                Monitor icon stays frontend-derived from monitor protocol or type rather than
                persisted icon metadata.
              </p>
              {field.kind === 'single-fixed' ? (
                <p>
                  Probe region is pinned to{' '}
                  <span className="font-semibold text-foreground">
                    {field.location.locationId.toUpperCase()} · {field.location.displayName}
                  </span>{' '}
                  based on the current enabled catalog. Multi-region probes are not available yet.
                </p>
              ) : (
                <p>
                  Operator selects probe region from {field.locations.length} enabled catalog
                  entries.
                </p>
              )}
            </CardContent>
          </Card>
        </div>
      </AppShell>
    )
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound()
    }

    return (
      <AppShell currentPath={`/services/${serviceId}/monitors/new`}>
        <EmptyState
          description={`${error instanceof Error ? error.message : 'Unable to load monitor create flow.'}`}
          title="Create monitor unavailable"
        />
      </AppShell>
    )
  }
}
