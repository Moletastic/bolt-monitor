import Link from 'next/link'

import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getSchedulerConfig, listProbeLocations } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

type LoadResult<T> = { data: T; error?: never } | { data?: never; error: string }

async function loadSettings<T>(loader: () => Promise<T>): Promise<LoadResult<T>> {
  try {
    return { data: await loader() }
  } catch (error) {
    return { error: error instanceof Error ? error.message : 'Unable to load settings data.' }
  }
}

export default async function ConfigPage() {
  const [schedulerResult, probesResult] = await Promise.all([
    loadSettings(getSchedulerConfig),
    loadSettings(listProbeLocations),
  ])
  const scheduler = schedulerResult.data
  const schedulerError = schedulerResult.error
  const probes = probesResult.data
  const probesError = probesResult.error
  const apiBaseConfigured = Boolean(process.env.NEXT_PUBLIC_MONITOR_API_BASE_URL)

  return (
    <AppShell currentPath="/config">
      <div className="grid gap-6">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Settings</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Control-plane overview for scheduler, probe catalog, and safe dashboard setup context.
          </p>
        </div>
        <section className="grid gap-4 lg:grid-cols-3">
          <Card>
            <CardHeader>
              <CardTitle>Scheduler</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {schedulerError ? (
                <EmptyState
                  description={`${schedulerError} Scheduler controls may be unavailable until the monitor API is reachable.`}
                  title="Scheduler unavailable"
                />
              ) : scheduler ? (
                <>
                  <div className="rounded-lg border border-border bg-surface-low p-4">
                    <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                      Recurring execution
                    </p>
                    <p className="mt-2 text-2xl font-semibold text-foreground">
                      {scheduler.recurringEnabled ? 'Enabled' : 'Disabled'}
                    </p>
                    <p className="mt-2 text-sm text-muted-foreground">
                      Updated{' '}
                      <span className="font-mono">{formatDateTime(scheduler.updatedAt)}</span>
                    </p>
                  </div>
                  <Link
                    className="inline-flex rounded-md border border-primary/40 bg-primary/10 px-3 py-2 text-sm font-semibold text-primary hover:bg-primary/20"
                    href="/admin/scheduler"
                  >
                    Open scheduler controls
                  </Link>
                </>
              ) : null}
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Probe locations</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {probesError ? (
                <EmptyState
                  description={`${probesError} Probe catalog context cannot be shown right now.`}
                  title="Probe catalog unavailable"
                />
              ) : probes ? (
                <>
                  <div className="rounded-lg border border-border bg-surface-low p-4">
                    <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                      Enabled probes
                    </p>
                    <p className="mt-2 text-2xl font-semibold text-foreground">
                      {probes.filter((probe) => probe.enabled).length}/{probes.length}
                    </p>
                  </div>
                  <div className="grid gap-2">
                    {probes.map((probe) => (
                      <div
                        className="flex items-center justify-between rounded-lg border border-border bg-surface-low px-3 py-2 text-sm"
                        key={probe.locationId}
                      >
                        <span className="font-medium text-foreground">{probe.displayName}</span>
                        <span className="text-muted-foreground">
                          {probe.enabled ? 'Enabled' : 'Disabled'}
                        </span>
                      </div>
                    ))}
                  </div>
                </>
              ) : null}
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Dashboard setup</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="rounded-lg border border-border bg-surface-low p-4">
                <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                  Monitor API base URL
                </p>
                <p className="mt-2 text-xl font-semibold text-foreground">
                  {apiBaseConfigured ? 'Configured' : 'Missing'}
                </p>
                <p className="mt-2 text-sm text-muted-foreground">
                  Value is intentionally hidden. Dashboard server fetches require this setting.
                </p>
              </div>
              <div className="rounded-lg border border-border bg-surface-low p-4 text-sm text-muted-foreground">
                <p className="font-semibold text-foreground">Bootstrap assumptions</p>
                <ul className="mt-3 space-y-2">
                  <li>Single tenant context</li>
                  <li>Service-first API is source of truth</li>
                  <li>Probe picker follows current built-in catalog</li>
                </ul>
              </div>
            </CardContent>
          </Card>
        </section>
      </div>
    </AppShell>
  )
}
