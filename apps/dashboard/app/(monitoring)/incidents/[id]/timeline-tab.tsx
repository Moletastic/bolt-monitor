import { EmptyState } from '@/components/empty-state'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getIncidentActivities } from '@/lib/api'
import type { IncidentActivity } from '@/lib/types'
import { formatDateTime } from '@/lib/utils'

interface TimelineTabProps {
  incidentId: string
}

export async function TimelineTab({ incidentId }: TimelineTabProps) {
  let activities: IncidentActivity[] = []
  let error: string | null = null

  try {
    activities = [...(await getIncidentActivities(incidentId))].sort((a, b) =>
      a.timestamp.localeCompare(b.timestamp)
    )
  } catch (cause) {
    error = cause instanceof Error ? cause.message : 'Unable to load incident activity.'
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Timeline</CardTitle>
      </CardHeader>
      <CardContent>
        {error ? (
          <div className="rounded-md border border-status-down/30 bg-status-down/10 px-3 py-2 text-sm text-status-down">
            {error}
          </div>
        ) : activities.length === 0 ? (
          <EmptyState
            description="No activity recorded for this incident yet."
            title="No activity"
          />
        ) : (
          <div className="space-y-3">
            {activities.map((activity) => (
              <div
                key={activity.activityId}
                className="rounded-lg border border-border bg-surface-low p-4"
              >
                <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
                  {activity.action.replaceAll('_', ' ')}
                </p>
                <p className="mt-2 text-sm text-foreground">{formatDateTime(activity.timestamp)}</p>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
