import Link from 'next/link'

import { MonitorProtocolBadge } from '@/components/monitor-protocol-badge'
import { StatusChip } from '@/components/status-chip'
import { SubmitButton } from '@/components/submit-button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { toggleMonitorAction } from '@/lib/actions'
import type { Monitor } from '@/lib/types'
import { formatDateTime, formatDuration, formatProbeLocations } from '@/lib/utils'

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
    return monitor.http?.target ?? formatProbeLocations(monitor.probeLocations)
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between gap-3 space-y-0">
        <CardTitle>Monitor overview</CardTitle>
        {readOnly ? (
          <span className="rounded-full border border-status-warn/30 bg-status-warn/10 px-2.5 py-1 text-[11px] font-bold uppercase tracking-[0.2em] text-status-warn">
            Read-only
          </span>
        ) : null}
      </CardHeader>
      <CardContent className="space-y-4">
        {readOnly ? (
          <p className="rounded-md border border-status-warn/30 bg-status-warn/10 px-3 py-2 text-sm text-status-warn">
            Monitor enable/disable is disabled because the parent service is archived.
          </p>
        ) : null}
        <div className="grid gap-4 md:hidden">
          {monitors.map((monitor) => (
            <Link
              className="block rounded-lg border border-border bg-surface-low p-4 transition-colors hover:border-primary/40 hover:bg-surface-low/80"
              href={`/services/${monitor.serviceId}/monitors/${monitor.monitorId}`}
              key={monitor.monitorId}
            >
              <div className="mb-4 flex items-start justify-between gap-4">
                <div className="flex items-start gap-3">
                  <MonitorProtocolBadge type={monitor.type} />
                  <div>
                    <span className="text-lg font-semibold text-foreground">{monitor.name}</span>
                    <p className="mt-1 break-all text-xs text-muted-foreground">
                      {getTarget(monitor)}
                    </p>
                  </div>
                </div>
                <StatusChip status={monitor.status?.currentStatus} />
              </div>
              <dl className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <dt className="text-muted-foreground">Protocol</dt>
                  <dd className="mt-1 font-semibold text-foreground">
                    {monitor.type.toUpperCase()}
                  </dd>
                </div>
                <div>
                  <dt className="text-muted-foreground">State</dt>
                  <dd className="mt-1 font-semibold text-foreground">
                    {monitor.enabled ? 'Enabled' : 'Disabled'}
                  </dd>
                </div>
                <div>
                  <dt className="text-muted-foreground">Probe</dt>
                  <dd className="mt-1 font-semibold text-foreground">
                    {formatProbeLocations(monitor.probeLocations)}
                  </dd>
                </div>
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
              <form
                action={toggleMonitorAction}
                className="mt-4 flex justify-end pointer-events-auto"
              >
                <input name="serviceId" type="hidden" value={monitor.serviceId} />
                <input name="monitorId" type="hidden" value={monitor.monitorId} />
                <input name="enabled" type="hidden" value={monitor.enabled ? 'false' : 'true'} />
                <input name="returnTo" type="hidden" value={returnTo} />
                <SubmitButton
                  disabled={readOnly}
                  size="sm"
                  variant={monitor.enabled ? 'outline' : 'default'}
                >
                  {monitor.enabled ? 'Disable' : 'Enable'}
                </SubmitButton>
              </form>
            </Link>
          ))}
        </div>
        <div className="hidden overflow-x-auto md:block">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Protocol</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Enabled</TableHead>
                <TableHead>Last check</TableHead>
                <TableHead>Duration</TableHead>
                <TableHead>Probe location</TableHead>
                <TableHead className="text-right">Action</TableHead>
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
                        <div className="max-w-xs truncate text-xs text-muted-foreground">
                          {getTarget(monitor)}
                        </div>
                      </div>
                    </Link>
                  </TableCell>
                  <TableCell>
                    <Link href={`/services/${monitor.serviceId}/monitors/${monitor.monitorId}`}>
                      {monitor.type.toUpperCase()}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <Link href={`/services/${monitor.serviceId}/monitors/${monitor.monitorId}`}>
                      <StatusChip status={monitor.status?.currentStatus} />
                    </Link>
                  </TableCell>
                  <TableCell>
                    <Link href={`/services/${monitor.serviceId}/monitors/${monitor.monitorId}`}>
                      {monitor.enabled ? 'Enabled' : 'Disabled'}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <Link href={`/services/${monitor.serviceId}/monitors/${monitor.monitorId}`}>
                      {formatDateTime(monitor.status?.lastCheckedAt)}
                    </Link>
                  </TableCell>
                  <TableCell className="font-mono">
                    <Link href={`/services/${monitor.serviceId}/monitors/${monitor.monitorId}`}>
                      {formatDuration(monitor.status?.lastDurationMs)}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <Link href={`/services/${monitor.serviceId}/monitors/${monitor.monitorId}`}>
                      {monitor.status?.lastProbeLocationId ??
                        formatProbeLocations(monitor.probeLocations)}
                    </Link>
                  </TableCell>
                  <TableCell className="text-right">
                    <form action={toggleMonitorAction} className="inline-flex">
                      <input name="serviceId" type="hidden" value={monitor.serviceId} />
                      <input name="monitorId" type="hidden" value={monitor.monitorId} />
                      <input
                        name="enabled"
                        type="hidden"
                        value={monitor.enabled ? 'false' : 'true'}
                      />
                      <input name="returnTo" type="hidden" value={returnTo} />
                      <SubmitButton
                        disabled={readOnly}
                        size="sm"
                        variant={monitor.enabled ? 'outline' : 'default'}
                      >
                        {monitor.enabled ? 'Disable' : 'Enable'}
                      </SubmitButton>
                    </form>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  )
}
