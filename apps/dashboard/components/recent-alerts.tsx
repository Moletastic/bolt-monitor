import Link from 'next/link'
import { AlertOctagon, AlertTriangle, CheckCircle2 } from 'lucide-react'

import { IncidentTimestamp } from '@/components/incident-timestamp'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { Incident } from '@/lib/types'

function incidentStatusPresentation(status: string) {
  const normalized = status.toLowerCase()
  if (normalized === 'resolved') {
    return {
      Icon: CheckCircle2,
      className: 'text-status-up',
      ariaLabel: 'Resolved incident',
    }
  }
  if (normalized === 'acknowledged') {
    return {
      Icon: AlertTriangle,
      className: 'text-status-warn',
      ariaLabel: 'Acknowledged incident',
    }
  }
  return {
    Icon: AlertOctagon,
    className: 'text-status-down',
    ariaLabel: 'Open incident',
  }
}

export function RecentAlerts({
  incidents,
  serviceId,
  limit,
}: {
  incidents: Incident[]
  serviceId: string
  limit?: number
}) {
  if (incidents.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Recent alerts</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            No recent alerts for monitors under this service.
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent alerts</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        {incidents.map((incident) => {
          const { Icon, className, ariaLabel } = incidentStatusPresentation(incident.status)
          return (
            <Link
              className="flex items-start gap-3 rounded-md border border-transparent px-2 py-2 transition-colors hover:border-border hover:bg-surface-low"
              href={`/services/${serviceId}/monitors/${incident.monitorId}?tab=incidents`}
              key={incident.incidentId}
            >
              <Icon
                aria-hidden="true"
                aria-label={ariaLabel}
                className={`mt-0.5 h-5 w-5 flex-shrink-0 ${className}`}
              />
              <div className="min-w-0 flex-1 space-y-1">
                <p className="truncate text-xs font-medium text-foreground">{incident.summary}</p>
                <p className="text-xs text-muted-foreground">
                  <IncidentTimestamp iso={incident.openedAt} />
                </p>
              </div>
            </Link>
          )
        })}
        {limit !== undefined && incidents.length >= limit ? (
          <Link
            className="inline-flex pt-1 text-xs font-semibold text-primary hover:underline"
            href={`/services/${serviceId}/monitors/${incidents[0]?.monitorId}?tab=incidents`}
          >
            View more in incident context →
          </Link>
        ) : null}
      </CardContent>
    </Card>
  )
}
