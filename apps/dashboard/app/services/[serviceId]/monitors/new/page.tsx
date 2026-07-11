import { notFound } from 'next/navigation'

import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { MonitorForm } from '@/components/monitor-form'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ApiError, getService } from '@/lib/api'

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
    const service = await getService(serviceId)

    return (
      <AppShell
        breadcrumbs={[
          { label: 'Services', href: '/services' },
          { label: service.name || 'Service', href: `/services/${service.serviceId}` },
          { label: 'Create monitor' },
        ]}
        currentPath={`/services/${serviceId}/monitors/new`}
      >
        <div className="grid gap-6 xl:grid-cols-[1.7fr_1fr]">
          <MonitorForm error={query.error} serviceId={service.serviceId} />
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
              <p>Monitor payloads are submitted without execution-location selection.</p>
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
      <AppShell
        breadcrumbs={[
          { label: 'Services', href: '/services' },
          { label: 'Service', href: `/services/${serviceId}` },
          { label: 'Create monitor' },
        ]}
        currentPath={`/services/${serviceId}/monitors/new`}
      >
        <EmptyState
          description={`${error instanceof Error ? error.message : 'Unable to load monitor create flow.'}`}
          title="Create monitor unavailable"
        />
      </AppShell>
    )
  }
}
