'use client'

import { useEffect, useState } from 'react'

import { EmptyState } from '@/components/empty-state'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { listMonitorAuditEvents, listServiceAuditEvents } from '@/lib/api'
import type { AuditEvent } from '@/lib/types'
import { formatDateTime } from '@/lib/utils'

interface AuditTabProps {
  serviceId: string
  monitorId: string
}

export function AuditTab({ serviceId, monitorId }: AuditTabProps) {
  const [monitorEvents, setMonitorEvents] = useState<AuditEvent[]>([])
  const [serviceEvents, setServiceEvents] = useState<AuditEvent[]>([])
  const [monitorError, setMonitorError] = useState<string | null>(null)
  const [serviceError, setServiceError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false

    async function loadAudit() {
      setLoading(true)
      setMonitorError(null)
      setServiceError(null)

      const [monitorResult, serviceResult] = await Promise.allSettled([
        listMonitorAuditEvents(serviceId, monitorId),
        listServiceAuditEvents(serviceId),
      ])

      if (cancelled) {
        return
      }

      if (monitorResult.status === 'fulfilled') {
        setMonitorEvents(monitorResult.value)
      } else {
        setMonitorEvents([])
        setMonitorError(
          monitorResult.reason instanceof Error
            ? monitorResult.reason.message
            : 'Unable to load monitor audit.'
        )
      }

      if (serviceResult.status === 'fulfilled') {
        setServiceEvents(serviceResult.value)
      } else {
        setServiceEvents([])
        setServiceError(
          serviceResult.reason instanceof Error
            ? serviceResult.reason.message
            : 'Unable to load service audit.'
        )
      }

      setLoading(false)
    }

    void loadAudit()

    return () => {
      cancelled = true
    }
  }, [monitorId, serviceId])

  const events = [...monitorEvents, ...serviceEvents].sort((a, b) =>
    a.occurredAt.localeCompare(b.occurredAt)
  )

  return (
    <Card>
      <CardHeader>
        <CardTitle>Audit</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {loading ? <p className="text-sm text-muted-foreground">Loading audit trail…</p> : null}
        {monitorError ? (
          <div className="rounded-md border border-status-down/30 bg-status-down/10 px-3 py-2 text-sm text-status-down">
            Monitor audit unavailable: {monitorError}
          </div>
        ) : null}
        {serviceError ? (
          <div className="rounded-md border border-status-down/30 bg-status-down/10 px-3 py-2 text-sm text-status-down">
            Service audit unavailable: {serviceError}
          </div>
        ) : null}
        {!loading && events.length === 0 ? (
          <EmptyState
            description="No audit events recorded for this incident scope."
            title="No audit events"
          />
        ) : null}
        {!loading && events.length > 0 ? (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>When</TableHead>
                  <TableHead>Event</TableHead>
                  <TableHead>Actor</TableHead>
                  <TableHead>Origin</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {events.map((event) => (
                  <TableRow key={event.auditId}>
                    <TableCell className="font-mono text-xs">
                      {formatDateTime(event.occurredAt)}
                    </TableCell>
                    <TableCell className="font-medium">{event.eventType}</TableCell>
                    <TableCell>{event.actor ?? '—'}</TableCell>
                    <TableCell>{event.origin ?? '—'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ) : null}
      </CardContent>
    </Card>
  )
}
