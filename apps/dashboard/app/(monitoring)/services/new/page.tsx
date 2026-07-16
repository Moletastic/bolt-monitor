import { AppShell } from '@/components/app-shell'
import { ServiceForm } from '@/components/service-form'
import { listEscalationPolicies } from '@/lib/api'

export default async function NewServicePage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>
}) {
  const params = await searchParams
  let policies: Awaited<ReturnType<typeof listEscalationPolicies>> = []
  try {
    policies = await listEscalationPolicies()
  } catch {
    policies = []
  }

  return (
    <AppShell
      breadcrumbs={[{ label: 'Services', href: '/services' }, { label: 'Create service' }]}
      currentPath="/services/new"
    >
      <div className="grid gap-6">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight text-foreground">Create service</h1>
          <p className="max-w-2xl text-sm text-muted-foreground">
            Define the identity and alert routing for a new service.
          </p>
        </div>
        <ServiceForm error={params.error} policies={policies} />
      </div>
    </AppShell>
  )
}
