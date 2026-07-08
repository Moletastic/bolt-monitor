import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'

export default function AuditTrailPage() {
  return (
    <AppShell
      breadcrumbs={[{ label: 'Incidents', href: '/incidents' }, { label: 'Audit trail' }]}
      currentPath="/audit-trail"
    >
      <EmptyState
        actionHref="/incidents"
        actionLabel="Open incidents"
        description="Audit trail view will be expanded here. Today, audit history is surfaced per incident and per monitor. Open Incidents for the current view of record."
        title="Audit trail not yet a destination"
      />
    </AppShell>
  )
}
