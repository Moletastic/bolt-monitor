import Link from 'next/link'

import { AppShell } from '@/components/app-shell'
import { ServiceForm } from '@/components/service-form'
import { ApiError, getService, listEscalationPolicies } from '@/lib/api'

export default async function EditServicePage({
  params,
  searchParams,
}: {
  params: Promise<{ serviceId: string }>
  searchParams: Promise<{ error?: string }>
}) {
  const { serviceId } = await params
  const query = await searchParams

  let policies: Awaited<ReturnType<typeof listEscalationPolicies>> = []
  try {
    policies = await listEscalationPolicies()
  } catch {
    policies = []
  }

  try {
    const service = await getService(serviceId)
    if (service.lifecycleState === 'archived') {
      return (
        <AppShell
          breadcrumbs={[
            { label: 'Services', href: '/services' },
            { label: service.name || 'Service', href: `/services/${serviceId}` },
            { label: 'Edit' },
          ]}
          currentPath={`/services/${serviceId}/edit`}
        >
          <div className="grid gap-6">
            <div className="space-y-2">
              <h1 className="text-3xl font-semibold tracking-tight text-foreground">
                Edit service
              </h1>
              <p className="rounded-md border border-status-warn/30 bg-status-warn/10 px-3 py-2 text-sm text-status-warn">
                Archived services cannot be edited. Recreate or restore the service to make changes.
              </p>
              <Link
                className="inline-flex font-semibold text-primary hover:underline"
                href={`/services/${serviceId}`}
              >
                Back to service
              </Link>
            </div>
          </div>
        </AppShell>
      )
    }

    return (
      <AppShell
        breadcrumbs={[
          { label: 'Services', href: '/services' },
          { label: service.name || 'Service', href: `/services/${serviceId}` },
          { label: 'Edit' },
        ]}
        currentPath={`/services/${serviceId}/edit`}
      >
        <div className="grid gap-6">
          <div className="space-y-2">
            <h1 className="text-3xl font-semibold tracking-tight text-foreground">Edit service</h1>
            <p className="max-w-2xl text-sm text-muted-foreground">
              Update the identity and alert routing for this service.
            </p>
          </div>
          <ServiceForm error={query.error} policies={policies} service={service} />
        </div>
      </AppShell>
    )
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      return (
        <AppShell
          breadcrumbs={[{ label: 'Services', href: '/services' }, { label: 'Edit' }]}
          currentPath={`/services/${serviceId}/edit`}
        >
          <div className="grid gap-6">
            <p className="rounded-md border border-status-down/30 bg-status-down/10 px-3 py-2 text-sm text-status-down">
              Service not found.
            </p>
            <Link
              className="inline-flex font-semibold text-primary hover:underline"
              href="/services"
            >
              Back to services
            </Link>
          </div>
        </AppShell>
      )
    }
    throw error
  }
}
