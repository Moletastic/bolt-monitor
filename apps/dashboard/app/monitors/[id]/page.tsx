import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'

export default async function LegacyMonitorDetailPage() {
  return (
    <AppShell currentPath="/services">
      <EmptyState
        actionHref="/services"
        actionLabel="Open services"
        description="Monitor detail now lives under its parent service route. Open the Services module and drill in from the owning service."
        title="Monitor route moved"
      />
    </AppShell>
  )
}
