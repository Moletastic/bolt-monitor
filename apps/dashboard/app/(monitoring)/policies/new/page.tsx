import { AppShell } from '@/components/app-shell'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { EscalationPolicyForm } from '@/components/escalation-policy-form'
import { listNotificationChannels } from '@/lib/api'

export default async function NewPolicyPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>
}) {
  const channels = await listNotificationChannels().catch(() => [])
  return (
    <AppShell
      breadcrumbs={[
        { label: 'Notification routes', href: '/policies' },
        { label: 'Create route' },
      ]}
      currentPath="/policies"
    >
      <div className="grid gap-6">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">New notification route</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Order the channels that fire when an incident opens. Each step waits for the previous
            one.
          </p>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Policy details</CardTitle>
          </CardHeader>
          <CardContent>
            <EscalationPolicyForm
              availableChannels={channels}
              errorHref="/policies/new?error=1"
              mode="create"
              returnTo="/policies"
              searchParams={searchParams}
            />
          </CardContent>
        </Card>
      </div>
    </AppShell>
  )
}
