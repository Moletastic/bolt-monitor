import Link from 'next/link'

import { MonitorActionsMenu } from '@/components/monitor-actions-menu'
import { MonitorProtocolBadge } from '@/components/monitor-protocol-badge'
import { StatusChip } from '@/components/status-chip'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { Monitor } from '@/lib/types'
import { formatDateTime, formatDuration } from '@/lib/utils'

export function MonitorTable({
  monitors,
  returnTo,
  readOnly = false,
}: {
  monitors: Monitor[]
  returnTo: string
  readOnly?: boolean
}) {
  function getTarget(monitor: Monitor) {
    return monitor.http?.target ?? ''
  }

  return (
    <div className="grid gap-4">
      {readOnly ? (
        <p className="rounded-md border border-status-warn/30 bg-status-warn/10 px-3 py-2 text-sm text-status-warn">
          Monitor enable/disable is disabled because the parent service is archived.
        </p>
      ) : null}
      <div className="grid gap-4 md:hidden">
        {monitors.map((monitor) => (
          <div
            className="rounded-lg border border-border bg-surface-low p-4 transition-colors hover:border-primary/40 hover:bg-surface-low/80"
            key={monitor.monitorId}
          >
            <div className="mb-4 flex items-start justify-between gap-4">
              <div className="flex items-start gap-3">
                <MonitorProtocolBadge type={monitor.type} />
                <div>
                  <Link
                    className="text-lg font-semibold text-foreground hover:text-primary hover:underline"
                    href={`/services/${monitor.serviceId}/monitors/${monitor.monitorId}`}
                  >
                    {monitor.name}
                  </Link>
                  <p className="mt-1 break-all font-mono text-xs text-muted-foreground">
                    {getTarget(monitor)}
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-2">
                <StatusChip status={monitor.status?.currentStatus} />
                <MonitorActionsMenu
                  disabled={readOnly}
                  enabled={monitor.enabled}
                  monitorId={monitor.monitorId}
                  returnTo={returnTo}
                  serviceId={monitor.serviceId}
                />
              </div>
            </div>
            <dl className="grid grid-cols-2 gap-3 text-sm">
              <div>
                <dt className="text-muted-foreground">Last check</dt>
                <dd className="mt-1 text-foreground">
                  {formatDateTime(monitor.status?.lastCheckedAt)}
                </dd>
              </div>
              <div>
                <dt className="text-muted-foreground">Duration</dt>
                <dd className="mt-1 font-mono text-foreground">
                  {formatDuration(monitor.status?.lastDurationMs)}
                </dd>
              </div>
            </dl>
          </div>
        ))}
      </div>
      <div className="hidden overflow-x-auto md:block">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Duration</TableHead>
              <TableHead>Last check</TableHead>
              <TableHead className="w-12" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {monitors.map((monitor) => (
              <TableRow className="cursor-pointer hover:bg-muted/50" key={monitor.monitorId}>
                <TableCell>
                  <Link
                    className="flex items-center gap-3"
                    href={`/services/${monitor.serviceId}/monitors/${monitor.monitorId}`}
                  >
                    <MonitorProtocolBadge type={monitor.type} />
                    <div className="space-y-1">
                      <span className="font-semibold text-foreground hover:text-primary">
                        {monitor.name}
                      </span>
                      <div className="max-w-xs truncate font-mono text-xs text-muted-foreground">
                        {getTarget(monitor)}
                      </div>
                    </div>
                  </Link>
                </TableCell>
                <TableCell>
                  <Link href={`/services/${monitor.serviceId}/monitors/${monitor.monitorId}`}>
                    <StatusChip status={monitor.status?.currentStatus} />
                  </Link>
                </TableCell>
                <TableCell className="font-mono">
                  <Link href={`/services/${monitor.serviceId}/monitors/${monitor.monitorId}`}>
                    {formatDuration(monitor.status?.lastDurationMs)}
                  </Link>
                </TableCell>
                <TableCell className="font-mono">
                  <Link href={`/services/${monitor.serviceId}/monitors/${monitor.monitorId}`}>
                    {formatDateTime(monitor.status?.lastCheckedAt)}
                  </Link>
                </TableCell>
                <TableCell className="w-12">
                  <MonitorActionsMenu
                    disabled={readOnly}
                    enabled={monitor.enabled}
                    monitorId={monitor.monitorId}
                    returnTo={returnTo}
                    serviceId={monitor.serviceId}
                  />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
