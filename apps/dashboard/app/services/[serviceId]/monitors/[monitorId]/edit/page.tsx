import Link from 'next/link'
import { notFound } from 'next/navigation'

import { AppShell } from '@/components/app-shell'
import { MonitorForm } from '@/components/monitor-form'
import { ApiError, getMonitor, getService } from '@/lib/api'

export default async function EditServiceMonitorPage({
  params,
  searchParams,
}: {
  params: Promise<{ serviceId: string; monitorId: string }>
  searchParams: Promise<{ error?: string }>
}) {
  const { serviceId, monitorId } = await params
  const query = await searchParams

  try {
    const [service, monitor] = await Promise.all([
      getService(serviceId),
      getMonitor(serviceId, monitorId),
    ])

    return (
      <AppShell
        breadcrumbs={[
          { label: 'Services', href: '/services' },
          { label: service.name || 'Service', href: `/services/${serviceId}` },
          {
            label: monitor.name || 'Monitor',
            href: `/services/${serviceId}/monitors/${monitorId}`,
          },
          { label: 'Edit' },
        ]}
        currentPath={`/services/${serviceId}/monitors/${monitorId}/edit`}
      >
        <div className="grid gap-6">
          <div className="space-y-2">
            <h1 className="text-3xl font-semibold tracking-tight text-foreground">Edit monitor</h1>
            <p className="max-w-2xl text-sm text-muted-foreground">
              Update the monitor identity, request, and validation settings.
            </p>
          </div>
          <MonitorForm error={query.error} monitor={monitor} serviceId={serviceId} />
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
          { label: 'Monitor', href: `/services/${serviceId}/monitors/${monitorId}` },
          { label: 'Edit' },
        ]}
        currentPath={`/services/${serviceId}/monitors/${monitorId}/edit`}
      >
        <div className="grid gap-6">
          <p className="rounded-md border border-status-down/30 bg-status-down/10 px-3 py-2 text-sm text-status-down">
            {error instanceof Error ? error.message : 'Unable to load monitor edit flow.'}
          </p>
          <Link
            className="inline-flex font-semibold text-primary hover:underline"
            href={`/services/${serviceId}/monitors/${monitorId}`}
          >
            Back to monitor
          </Link>
        </div>
      </AppShell>
    )
  }
}
