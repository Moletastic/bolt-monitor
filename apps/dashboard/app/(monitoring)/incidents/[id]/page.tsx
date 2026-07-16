import { notFound } from 'next/navigation'
import Link from 'next/link'
import { Suspense } from 'react'

import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { SamePageActionForm } from '@/components/same-page-action-form'
import { StatusChip } from '@/components/status-chip'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs } from '@/components/ui/tabs'
import { ApiError, getIncident } from '@/lib/api'
import { acknowledgeIncidentStateAction, resolveIncidentStateAction } from '@/lib/actions'
import { formatDateTime } from '@/lib/utils'
import { AlertHistoryTab } from './alert-history-tab'
import { AuditTab } from './audit-tab'
import { EscalationStateTab } from './escalation-state-tab'
import { TimelineTab } from './timeline-tab'

export default async function IncidentDetailPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>
  searchParams: Promise<{ error?: string; tab?: string }>
}) {
  const { id } = await params
  const query = await searchParams
  const activeTab = query.tab ?? 'timeline'

  try {
    const incident = await getIncident(id)
    const isOpen = incident.status === 'open' || incident.status === 'acknowledged'

    return (
      <AppShell
        breadcrumbs={[
          { label: 'Incidents', href: '/incidents' },
          { label: incident.summary || incident.incidentId },
        ]}
        currentPath="/incidents"
      >
        <div className="grid gap-6">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-[11px] font-bold uppercase tracking-[0.28em] text-muted-foreground">
                Incident detail
              </p>
              <h1 className="mt-2 text-2xl font-semibold tracking-tight">{incident.summary}</h1>
              <p className="mt-2 text-sm text-muted-foreground">
                Opened {formatDateTime(incident.openedAt)} · <StatusChip status={incident.status} />
              </p>
            </div>
            {isOpen && (
              <div className="flex gap-3">
                {incident.status === 'open' && (
                  <SamePageActionForm
                    action={acknowledgeIncidentStateAction}
                    buttonLabel="Acknowledge"
                    pendingLabel="Acknowledging..."
                    variant="secondary"
                  >
                    <input name="incidentId" type="hidden" value={incident.incidentId} />
                    <input
                      name="returnTo"
                      type="hidden"
                      value={`/incidents/${incident.incidentId}`}
                    />
                  </SamePageActionForm>
                )}
                <SamePageActionForm
                  action={resolveIncidentStateAction}
                  buttonLabel="Resolve"
                  pendingLabel="Resolving..."
                  variant="default"
                >
                  <input name="incidentId" type="hidden" value={incident.incidentId} />
                  <input
                    name="returnTo"
                    type="hidden"
                    value={`/incidents/${incident.incidentId}`}
                  />
                </SamePageActionForm>
              </div>
            )}
          </div>
          <div className="grid gap-6 xl:grid-cols-[0.9fr_1.1fr]">
            <Card>
              <CardHeader>
                <CardTitle>Details</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="rounded-lg border border-border bg-surface-low p-4">
                  <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                    Status
                  </p>
                  <p className="mt-2 text-sm text-foreground capitalize">{incident.status}</p>
                </div>
                <div className="rounded-lg border border-border bg-surface-low p-4">
                  <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                    Monitor
                  </p>
                  {incident.serviceId ? (
                    <Link
                      className="mt-2 inline-block text-sm text-primary hover:underline"
                      href={`/services/${incident.serviceId}/monitors/${incident.monitorId}`}
                    >
                      Open related monitor
                    </Link>
                  ) : (
                    <p className="mt-2 text-sm text-muted-foreground">
                      Related monitor unavailable
                    </p>
                  )}
                </div>
                <div className="rounded-lg border border-border bg-surface-low p-4">
                  <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                    Origin
                  </p>
                  <p className="mt-2 text-sm text-foreground">{incident.origin ?? '—'}</p>
                </div>
              </CardContent>
            </Card>
            <div className="space-y-4">
              <Tabs
                basePath={`/incidents/${incident.incidentId}`}
                tabs={[
                  { label: 'Timeline', href: `/incidents/${incident.incidentId}?tab=timeline` },
                  {
                    label: 'Escalation',
                    href: `/incidents/${incident.incidentId}?tab=escalation`,
                  },
                  {
                    label: 'Alert History',
                    href: `/incidents/${incident.incidentId}?tab=alerts`,
                  },
                  { label: 'Audit', href: `/incidents/${incident.incidentId}?tab=audit` },
                ]}
              />
              {activeTab === 'timeline' ? (
                <Suspense
                  fallback={<TabLoadingFallback title="Timeline" message="Loading timeline…" />}
                >
                  <TimelineTab incidentId={incident.incidentId} />
                </Suspense>
              ) : null}
              {activeTab === 'escalation' ? (
                <EscalationStateTab incidentId={incident.incidentId} />
              ) : null}
              {activeTab === 'alerts' && incident.serviceId ? (
                <Suspense
                  fallback={
                    <TabLoadingFallback title="Alert History" message="Loading alert history…" />
                  }
                >
                  <AlertHistoryTab
                    monitorId={incident.monitorId}
                    openedAt={incident.openedAt}
                    acknowledgedAt={incident.acknowledgedAt}
                    resolvedAt={incident.resolvedAt}
                    serviceId={incident.serviceId}
                  />
                </Suspense>
              ) : null}
              {activeTab === 'alerts' && !incident.serviceId ? (
                <Card>
                  <CardContent className="pt-6">
                    <EmptyState
                      description="Service context missing for this incident, so related runs cannot be loaded."
                      title="Alert history unavailable"
                    />
                  </CardContent>
                </Card>
              ) : null}
              {activeTab === 'audit' && incident.serviceId ? (
                <Suspense
                  fallback={<TabLoadingFallback title="Audit" message="Loading audit trail…" />}
                >
                  <AuditTab monitorId={incident.monitorId} serviceId={incident.serviceId} />
                </Suspense>
              ) : null}
              {activeTab === 'audit' && !incident.serviceId ? (
                <Card>
                  <CardContent className="pt-6">
                    <EmptyState
                      description="Service context missing for this incident, so audit history cannot be loaded."
                      title="Audit unavailable"
                    />
                  </CardContent>
                </Card>
              ) : null}
            </div>
          </div>
        </div>
      </AppShell>
    )
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound()
    }

    return (
      <AppShell
        breadcrumbs={[{ label: 'Incidents', href: '/incidents' }, { label: 'Incident' }]}
        currentPath="/incidents"
      >
        <EmptyState
          description={`${error instanceof Error ? error.message : 'Unable to load incident detail.'}`}
          title="Incident unavailable"
        />
      </AppShell>
    )
  }
}

function TabLoadingFallback({ title, message }: { title: string; message: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">{message}</p>
      </CardContent>
    </Card>
  )
}
