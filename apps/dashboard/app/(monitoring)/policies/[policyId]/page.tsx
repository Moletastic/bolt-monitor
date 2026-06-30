import Link from 'next/link'

import { AppShell } from '@/components/app-shell'
import { EscalationPolicyForm } from '@/components/escalation-policy-form'
import { UnavailableCard } from '@/components/unavailable-card'
import { getEscalationPolicy, listNotificationChannels } from '@/lib/api'

type Params = Promise<{ policyId: string }>

type SearchParams = Promise<{ error?: string; updated?: string }>

export default async function EditEscalationPolicyPage({
  params,
  searchParams,
}: {
  params: Params
  searchParams: SearchParams
}) {
  const { policyId } = await params
  const { error, updated } = await searchParams

  let policy
  try {
    policy = await getEscalationPolicy(policyId)
  } catch (fetchError) {
    const message =
      fetchError instanceof Error ? fetchError.message : 'Unable to load escalation policy.'
    return (
      <AppShell currentPath="/policies">
        <div className="grid gap-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-foreground">
              Notification route
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Update route steps, channel order, and business hours window.
            </p>
          </div>
          <UnavailableCard message={message} title="Escalation policy unavailable" />
          <Link className="text-sm text-primary hover:underline" href="/policies">
            Back to routes
          </Link>
        </div>
      </AppShell>
    )
  }

  const channels = await listNotificationChannels().catch(() => [])

  return (
    <AppShell currentPath="/policies">
      <div className="grid gap-6">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">
            Edit notification route
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Update route steps, channel order, and business hours window.
          </p>
        </div>

        {updated ? (
          <p className="rounded-md border border-status-up/30 bg-status-up/10 px-3 py-2 text-sm text-status-up">
            Notification route saved.
          </p>
        ) : null}

        {error ? (
          <p className="rounded-md border border-status-down/30 bg-status-down/10 px-3 py-2 text-sm text-status-down">
            {error}
          </p>
        ) : null}

        <EscalationPolicyForm
          availableChannels={channels}
          mode="edit"
          policyId={policy.policyId}
          initialName={policy.name}
          initialDescription={policy.description ?? ''}
          initialBusinessHoursPath={policy.businessHoursPath}
          initialOffHoursPath={policy.offHoursPath}
          initialBusinessHours={undefined}
          returnTo={`/policies/${policy.policyId}`}
          errorHref={`/policies/${policy.policyId}?error=1`}
          searchParams={Promise.resolve(searchParams)}
        />
      </div>
    </AppShell>
  )
}
