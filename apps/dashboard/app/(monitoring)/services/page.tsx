import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { FocusOnMount } from '@/components/focus-on-mount'
import { ServiceListLayout } from '@/components/service-list-layout'
import { ServiceListStatusToast } from '@/components/service-list-status-toast'
import { Feedback } from '@/components/ui/feedback'
import { ApiError, listServices } from '@/lib/api'

function DeletedServiceFeedback({ active = true }: { active?: boolean }) {
  return (
    <FocusOnMount active={active}>
      <Feedback tone="success">Service permanently deleted.</Feedback>
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

    return (
      <AppShell currentPath="/services">
        <ServiceListStatusToast services={services} />
        <div className="grid gap-6">
          <h1 className="sr-only">Services</h1>
          {query.deletedService ? <DeletedServiceFeedback /> : null}
          <ServiceListLayout services={services} />
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
