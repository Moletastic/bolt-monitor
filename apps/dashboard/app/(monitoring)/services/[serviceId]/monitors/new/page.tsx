import { notFound } from 'next/navigation'

import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { MonitorForm } from '@/components/monitor-form'
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
        <div className="grid gap-6">
          <MonitorForm error={query.error} serviceId={service.serviceId} />
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
