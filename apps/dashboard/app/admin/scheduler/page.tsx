import { AppShell } from '@/components/app-shell'
import { SchedulerConfigForm } from '@/components/scheduler-config-form'
import { UnavailableCard } from '@/components/unavailable-card'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getSchedulerConfig } from '@/lib/api'

export default async function SchedulerPage() {
  let config
  try {
    config = await getSchedulerConfig()
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unable to load scheduler config.'
    return (
      <AppShell currentPath="/admin/scheduler">
        <div className="grid gap-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Scheduler</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Manage recurring monitor execution settings.
            </p>
          </div>
          <UnavailableCard message={message} title="Scheduler unavailable" />
        </div>
      </AppShell>
    )
  }

  return (
    <AppShell currentPath="/admin/scheduler">
      <div className="grid gap-6">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Scheduler</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Manage recurring monitor execution settings.
          </p>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Recurring execution</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="rounded-lg border border-border bg-surface-low p-4">
              <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                Status
              </p>
              <p className="mt-2 text-xl font-semibold text-foreground">
                {config.recurringEnabled ? 'Enabled' : 'Disabled'}
              </p>
            </div>
            <SchedulerConfigForm recurringEnabled={config.recurringEnabled} />
          </CardContent>
        </Card>
      </div>
    </AppShell>
  )
}
