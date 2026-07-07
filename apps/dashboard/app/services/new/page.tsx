import { AppShell } from '@/components/app-shell'
import { ServiceForm } from '@/components/service-form'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
    <AppShell currentPath="/services/new">
      <div className="grid gap-6 xl:grid-cols-[1.7fr_1fr]">
        <ServiceForm error={params.error} policies={policies} />
        <Card>
          <CardHeader>
            <CardTitle>Create flow notes</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4 text-sm text-muted-foreground">
            <p>
              Services own the top-level identity and can start in draft before any monitor exists.
            </p>
            <p>
              `serviceCategory` drives one primary service icon while monitor icons remain
              frontend-derived.
            </p>
            <p>
              Nested monitor creation moves to the service detail view after the service is saved.
            </p>
          </CardContent>
        </Card>
      </div>
    </AppShell>
  )
}
