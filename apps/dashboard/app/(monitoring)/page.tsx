import Link from 'next/link'
import { compareDesc, parseISO } from 'date-fns'
import type { ReactNode } from 'react'

import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { MonitorTrafficLights } from '@/components/monitor-traffic-light'
import { ServiceIcon } from '@/components/service-icon'
import { StatusChip } from '@/components/status-chip'
import { UnavailableCard } from '@/components/unavailable-card'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { getSchedulerConfig, listIncidents, listServices } from '@/lib/api'
import type { Incident, SchedulerConfigResponse, Service } from '@/lib/types'
import { formatDateTime } from '@/lib/utils'

type LoadResult<T> = { data: T; error?: never } | { data?: never; error: string }

type AttentionItem = {
  href: string
  label: string
  detail: ReactNode
  tone: 'down' | 'warn' | 'info'
}

async function loadOverview<T>(loader: () => Promise<T>): Promise<LoadResult<T>> {
  try {
    return { data: await loader() }
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Unable to load dashboard data.' }
  }
}

function normalizeStatus(status?: string) {
  return status?.toUpperCase() ?? 'UNKNOWN'
}

function isOpenIncident(incident: Incident) {
  const status = normalizeStatus(incident.status)
  return status !== 'CLOSED' && status !== 'RESOLVED'
}

function getMonitorCount(service: Service) {
  return service.monitorCount ?? service.monitors?.length ?? 0
}

function getEnabledMonitorCount(service: Service) {
  return (
    service.enabledMonitorCount ?? service.monitors?.filter((monitor) => monitor.enabled).length
  )
}

function buildAttentionItems({
  services,
  incidents,
  scheduler,
}: {
  services: Service[]
  incidents: Incident[]
  scheduler?: SchedulerConfigResponse
}) {
  const downServices = services.filter(
    (service) => normalizeStatus(service.rollupStatus) === 'DOWN'
  )
  const openIncidents = incidents.filter(isOpenIncident)
  const uncoveredServices = services.filter((service) => getMonitorCount(service) === 0)
  const disabledCoverageServices = services.filter((service) => {
    const monitorCount = getMonitorCount(service)
    const enabledCount = getEnabledMonitorCount(service)
    return monitorCount > 0 && enabledCount !== undefined && enabledCount < monitorCount
  })
  const draftServices = services.filter((service) => service.lifecycleState === 'draft')
  const items: AttentionItem[] = []

  if (scheduler && !scheduler.recurringEnabled) {
    items.push({
      href: '/admin/scheduler',
      label: 'Scheduler disabled',
      detail: 'Recurring monitor execution is off.',
      tone: 'down',
    })
  }

  for (const service of downServices.slice(0, 4)) {
    items.push({
      href: `/services/${service.serviceId}`,
      label: `${service.name} is down`,
      detail: `${getMonitorCount(service)} monitor${getMonitorCount(service) === 1 ? '' : 's'} tracked.`,
      tone: 'down',
    })
  }

  for (const incident of openIncidents.slice(0, 4)) {
    items.push({
      href: incident.serviceId
        ? `/services/${incident.serviceId}/monitors/${incident.monitorId}`
        : '/incidents?status=open',
      label: incident.summary,
      detail: (
        <span>
          Incident {incident.status} opened{' '}
          <span className="font-mono text-xs">{formatDateTime(incident.openedAt)}</span>.
        </span>
      ),
      tone: 'warn',
    })
  }

  for (const service of uncoveredServices.slice(0, 3)) {
    items.push({
      href: `/services/${service.serviceId}/monitors/new`,
      label: `${service.name} has no monitors`,
      detail: 'Add monitor coverage before relying on service health.',
      tone: 'warn',
    })
  }

  for (const service of disabledCoverageServices.slice(0, 3)) {
    const enabledCount = getEnabledMonitorCount(service) ?? 0
    items.push({
      href: `/services/${service.serviceId}`,
      label: `${service.name} has disabled coverage`,
      detail: `${enabledCount}/${getMonitorCount(service)} monitors enabled.`,
      tone: 'warn',
    })
  }

  for (const service of draftServices.slice(0, 3)) {
    items.push({
      href: `/services/${service.serviceId}`,
      label: `${service.name} is draft`,
      detail: 'Finish setup or activate when coverage is ready.',
      tone: 'info',
    })
  }

  return items.slice(0, 8)
}

function SummaryCard({
  title,
  value,
  description,
  tone = 'default',
}: {
  title: string
  value: string | number
  description: string
  tone?: 'default' | 'down' | 'warn' | 'up'
}) {
  const toneClass = {
    default: 'text-foreground',
    down: 'text-status-down',
    warn: 'text-status-warn',
    up: 'text-status-up',
  }[tone]

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className={`font-mono text-3xl font-semibold ${toneClass}`}>{value}</p>
        <p className="mt-2 text-sm text-muted-foreground">{description}</p>
      </CardContent>
    </Card>
  )
}

function AttentionQueue({ items }: { items: AttentionItem[] }) {
  const toneClass = {
    down: 'border-status-down/30 bg-status-down/10 text-status-down',
    warn: 'border-status-warn/30 bg-status-warn/10 text-status-warn',
    info: 'border-primary/30 bg-primary/10 text-primary',
  }
  const toneLabel = {
    down: 'Action needed',
    warn: 'At risk',
    info: 'Heads-up',
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Attention queue</CardTitle>
      </CardHeader>
      <CardContent>
        {items.length === 0 ? (
          <p className="rounded-lg border border-status-up/30 bg-status-up/10 px-4 py-3 text-sm text-status-up">
            All clear across loaded modules.
          </p>
        ) : (
          <div className="grid gap-3">
            {items.map((item) => (
              <Link
                className="rounded-lg border border-border bg-surface-low p-4 transition-colors hover:border-primary/40"
                href={item.href}
                key={`${item.href}-${item.label}`}
              >
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div>
                    <p className="font-semibold text-foreground">{item.label}</p>
                    <p className="mt-1 text-sm text-muted-foreground">{item.detail}</p>
                  </div>
                  <span
                    className={`w-fit rounded-full border px-2.5 py-1 text-[11px] font-bold uppercase tracking-[0.2em] ${toneClass[item.tone]}`}
                  >
                    {toneLabel[item.tone]}
                  </span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function ServiceHealthMatrix({ services }: { services: Service[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Service health matrix</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Service</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Lifecycle</TableHead>
                <TableHead>Coverage</TableHead>
                <TableHead>Updated</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {services.map((service) => {
                const monitorCount = getMonitorCount(service)
                const enabledCount = getEnabledMonitorCount(service)
                return (
                  <TableRow key={service.serviceId}>
                    <TableCell>
                      <Link
                        className="flex items-center gap-3 text-foreground hover:text-primary"
                        href={`/services/${service.serviceId}`}
                      >
                        <ServiceIcon serviceCategory={service.serviceCategory} />
                        <span>
                          <span className="block font-semibold">{service.name}</span>
                          <span className="block text-xs uppercase tracking-[0.2em] text-muted-foreground">
                            {service.lifecycleState}
                          </span>
                        </span>
                      </Link>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-col gap-1">
                        <StatusChip status={service.rollupStatus ?? service.lifecycleState} />
                        <MonitorTrafficLights monitors={service.monitors ?? []} />
                      </div>
                    </TableCell>
                    <TableCell className="capitalize">{service.lifecycleState}</TableCell>
                    <TableCell>
                      {enabledCount === undefined
                        ? `${monitorCount} monitor${monitorCount === 1 ? '' : 's'}`
                        : `${enabledCount}/${monitorCount} enabled`}
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                      {formatDateTime(service.updatedAt)}
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  )
}

function RecentIncidents({ incidents }: { incidents: Incident[] }) {
  const recent = [...incidents]
    .sort((a, b) => compareDesc(parseISO(a.openedAt), parseISO(b.openedAt)))
    .slice(0, 5)

  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent incidents</CardTitle>
      </CardHeader>
      <CardContent>
        {recent.length === 0 ? (
          <p className="text-sm text-muted-foreground">No incidents recorded yet.</p>
        ) : (
          <div className="grid gap-3">
            {recent.map((incident) => (
              <Link
                className="rounded-lg border border-border bg-surface-low p-3 hover:border-primary/40"
                href={
                  incident.serviceId
                    ? `/services/${incident.serviceId}/monitors/${incident.monitorId}`
                    : '/incidents'
                }
                key={incident.incidentId}
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="line-clamp-1 text-sm font-semibold text-foreground">
                      {incident.summary}
                    </p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Opened <span className="font-mono">{formatDateTime(incident.openedAt)}</span>
                    </p>
                  </div>
                  <StatusChip status={incident.status} />
                </div>
              </Link>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function SetupGaps({
  draftCount,
  uncoveredCount,
  disabledCoverageCount,
  scheduler,
}: {
  draftCount: number
  uncoveredCount: number
  disabledCoverageCount: number
  scheduler?: SchedulerConfigResponse
}) {
  const gaps = [
    { label: 'Draft services', value: draftCount, href: '/services' },
    { label: 'Services without monitors', value: uncoveredCount, href: '/services' },
    {
      label: 'Services with disabled monitor coverage',
      value: disabledCoverageCount,
      href: '/services',
    },
    {
      label: 'Scheduler recurring execution',
      value: scheduler?.recurringEnabled ? 'Enabled' : 'Disabled',
      href: '/admin/scheduler',
    },
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle>Setup gaps</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-3">
        {gaps.map((gap) => (
          <Link
            className="flex items-center justify-between gap-3 rounded-lg border border-border bg-surface-low px-3 py-2 hover:border-primary/40"
            href={gap.href}
            key={gap.label}
          >
            <span className="text-sm text-muted-foreground">{gap.label}</span>
            <span className="font-mono text-sm font-semibold text-foreground">{gap.value}</span>
          </Link>
        ))}
      </CardContent>
    </Card>
  )
}

export default async function DashboardHome() {
  const [servicesResult, incidentsResult, schedulerResult] = await Promise.all([
    loadOverview(listServices),
    loadOverview(() => listIncidents()),
    loadOverview(getSchedulerConfig),
  ])

  const services = servicesResult.data ?? []
  const incidents = incidentsResult.data ?? []
  const scheduler = schedulerResult.data
  const servicesLoaded = !servicesResult.error
  const incidentsLoaded = !incidentsResult.error
  const schedulerLoaded = !schedulerResult.error

  const downCount = services.filter(
    (service) => normalizeStatus(service.rollupStatus) === 'DOWN'
  ).length
  const draftCount = services.filter((service) => service.lifecycleState === 'draft').length
  const uncoveredCount = services.filter((service) => getMonitorCount(service) === 0).length
  const disabledCoverageCount = services.filter((service) => {
    const monitorCount = getMonitorCount(service)
    const enabledCount = getEnabledMonitorCount(service)
    return monitorCount > 0 && enabledCount !== undefined && enabledCount < monitorCount
  }).length
  const openIncidentCount = incidents.filter(isOpenIncident).length
  const attentionItems = buildAttentionItems({ services, incidents, scheduler })

  return (
    <AppShell currentPath="/">
      <div className="grid gap-6">
        <section className="grid gap-3">
          <p className="text-[11px] font-bold uppercase tracking-[0.28em] text-primary">
            Operational overview
          </p>
          <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
            <div>
              <h1 className="text-3xl font-semibold tracking-tight text-foreground">
                Operator overview
              </h1>
              <p className="mt-2 max-w-3xl text-sm text-muted-foreground">
                Start here to see service rollups, incident pressure, scheduler state, and setup
                gaps before drilling into modules.
              </p>
            </div>
            <Link
              className="w-fit rounded-md border border-primary/40 bg-primary/10 px-3 py-2 text-sm font-semibold text-primary hover:bg-primary/20"
              href="/services/new"
            >
              Create service
            </Link>
          </div>
        </section>

        {servicesLoaded && services.length === 0 ? (
          <EmptyState
            actionHref="/services/new"
            actionLabel="Create first service"
            description="No services exist yet. Create a service before treating dashboard health totals as meaningful monitoring coverage."
            title="No monitored services yet"
          />
        ) : null}

        <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {servicesLoaded ? (
            <SummaryCard
              description="Services with derived DOWN rollup status."
              title="Services down"
              tone={downCount > 0 ? 'down' : 'up'}
              value={downCount}
            />
          ) : (
            <UnavailableCard message={servicesResult.error} title="Service health unavailable" />
          )}
          {incidentsLoaded ? (
            <SummaryCard
              description="Incidents not closed or resolved."
              title="Open incidents"
              tone={openIncidentCount > 0 ? 'warn' : 'up'}
              value={openIncidentCount}
            />
          ) : (
            <UnavailableCard message={incidentsResult.error} title="Incidents unavailable" />
          )}
          {servicesLoaded ? (
            <SummaryCard
              description="Services still in draft lifecycle state."
              title="Draft services"
              tone={draftCount > 0 ? 'warn' : 'up'}
              value={draftCount}
            />
          ) : (
            <UnavailableCard
              message="Service setup gaps require service data."
              title="Drafts unavailable"
            />
          )}
          {schedulerLoaded ? (
            <SummaryCard
              description="Recurring monitor execution control."
              title="Scheduler"
              tone={scheduler?.recurringEnabled ? 'up' : 'down'}
              value={scheduler?.recurringEnabled ? 'Enabled' : 'Disabled'}
            />
          ) : (
            <UnavailableCard message={schedulerResult.error} title="Scheduler unavailable" />
          )}
        </section>

        {(servicesLoaded || incidentsLoaded || schedulerLoaded) && (
          <AttentionQueue items={attentionItems} />
        )}

        {servicesLoaded && services.length > 0 ? <ServiceHealthMatrix services={services} /> : null}

        <section className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
          {incidentsLoaded ? (
            <RecentIncidents incidents={incidents} />
          ) : (
            <UnavailableCard message={incidentsResult.error} title="Recent incidents unavailable" />
          )}
          {servicesLoaded ? (
            <SetupGaps
              disabledCoverageCount={disabledCoverageCount}
              draftCount={draftCount}
              scheduler={scheduler}
              uncoveredCount={uncoveredCount}
            />
          ) : (
            <UnavailableCard
              message="Setup-gap details require service data."
              title="Setup gaps unavailable"
            />
          )}
        </section>
      </div>
    </AppShell>
  )
}
