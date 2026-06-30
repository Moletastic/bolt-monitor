'use client'

import { useEffect, useState } from 'react'

import { EmptyState } from '@/components/empty-state'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getIncidentActivities } from '@/lib/api'
import type { IncidentActivity } from '@/lib/types'
import { formatDateTime } from '@/lib/utils'

interface TimelineTabProps {
  incidentId: string
}

export function TimelineTab({ incidentId }: TimelineTabProps) {
  const [activities, setActivities] = useState<IncidentActivity[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function loadActivities() {
      setLoading(true)
      setError(null)
      try {
        const next = await getIncidentActivities(incidentId)
        if (cancelled) {
          return
        }
        setActivities([...next].sort((a, b) => a.timestamp.localeCompare(b.timestamp)))
      } catch (err) {
        if (cancelled) {
          return
        }
        setError(err instanceof Error ? err.message : 'Unable to load incident activity.')
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void loadActivities()

    return () => {
      cancelled = true
    }
  }, [incidentId])

  return (
    <Card>
      <CardHeader>
        <CardTitle>Timeline</CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <p className="text-sm text-muted-foreground">Loading timeline…</p>
        ) : error ? (
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
