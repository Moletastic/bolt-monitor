import Link from 'next/link'

import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ApiError, listEscalationPolicies } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

export default async function PoliciesPage() {
  let policies: Awaited<ReturnType<typeof listEscalationPolicies>> = []
  let loadError: string | undefined

  try {
    policies = await listEscalationPolicies()
  } catch (error) {
    loadError = error instanceof ApiError ? error.message : 'Unable to load notification routes.'
  }

  return (
    <AppShell currentPath="/policies">
      <div className="grid gap-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Notification routes</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Order the channels that fire when an incident opens. Each step waits for the previous
              one.
            </p>
          </div>
          <Link
            className="rounded-md border border-primary/40 bg-primary/10 px-3 py-2 text-sm font-semibold text-primary hover:bg-primary/20"
            href="/policies/new"
          >
            Create route
          </Link>
        </div>
        {loadError ? (
          <EmptyState
            description={`${loadError} Check local API connectivity and policy API availability.`}
            title="Routes unavailable"
          />
        ) : policies.length === 0 ? (
          <EmptyState
            actionHref="/policies/new"
            actionLabel="Create your first route"
            description="Routes decide who hears about incidents and when. Start with one that pages the on-call engineer."
            title="No routes yet"
          />
        ) : (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {policies.map((policy) => {
              const totalSteps =
                policy.businessHoursPath.steps.length + policy.offHoursPath.steps.length
              const channelCount = totalSteps
              return (
                <Link
                  className="block transition-colors"
                  href={`/policies/${policy.policyId}`}
                  key={policy.policyId}
                >
                  <Card className="h-full transition-colors hover:border-primary/40 hover:bg-surface-low/70">
                    <CardHeader>
                      <CardTitle>{policy.name}</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-3 text-sm">
                      {policy.description ? (
                        <p className="line-clamp-2 text-muted-foreground">{policy.description}</p>
                      ) : (
                        <p className="italic text-muted-foreground">No description provided.</p>
                      )}
                      <dl className="grid grid-cols-2 gap-3">
                        <div>
                          <dt className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                            Steps
                          </dt>
                          <dd className="mt-1 font-mono text-base text-foreground">{totalSteps}</dd>
                        </div>
                        <div>
                          <dt className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                            Channels
                          </dt>
                          <dd className="mt-1 font-mono text-base text-foreground">
                            {channelCount}
                          </dd>
                        </div>
                        <div className="col-span-2">
                          <dt className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                            Updated
                          </dt>
                          <dd className="mt-1 text-foreground">
                            {formatDateTime(policy.updatedAt)}
                          </dd>
                        </div>
                      </dl>
                    </CardContent>
                  </Card>
                </Link>
              )
            })}
          </div>
        )}
      </div>
    </AppShell>
  )
}
