import Link from 'next/link'

import { StatusChip } from '@/components/status-chip'
import { TechIcon } from '@/components/tech-icon'
import { Card, CardContent } from '@/components/ui/card'
import {
  formatMetricDuration,
  formatRecentUptime,
  serviceMetricStateLabel,
} from '@/lib/service-card-metrics'
import type { Service, ServiceCardTrendPoint } from '@/lib/types'
import { LocalTime } from '@/components/local-time'
import { cn } from '@/lib/utils'

function normalizeState(status?: string) {
  return status?.toUpperCase() ?? ''
}

function cardTone(status?: string) {
  const normalized = normalizeState(status)
  if (normalized === 'DOWN') {
    return 'border-status-down/40 shadow-[0_0_0_1px_hsl(var(--status-down)/0.18)]'
  }
  if (normalized === 'DEGRADED') {
    return 'border-status-warn/40 shadow-[0_0_0_1px_hsl(var(--status-warn)/0.16)]'
  }
  if (normalized === 'UP') {
    return 'border-status-up/25'
  }
  return 'border-border/80'
}

function iconTileTone(status?: string) {
  const normalized = normalizeState(status)
  if (normalized === 'DOWN') {
    return 'bg-status-down/15 text-status-down ring-1 ring-status-down/30'
  }
  if (normalized === 'DEGRADED') {
    return 'bg-status-warn/15 text-status-warn ring-1 ring-status-warn/30'
  }
  if (normalized === 'UP') {
    return 'bg-status-up/15 text-status-up ring-1 ring-status-up/30'
  }
  return 'bg-surface-high text-muted-foreground ring-1 ring-border/80'
}

function TrendSparkline({ points = [] }: { points?: ServiceCardTrendPoint[] }) {
  const visible = points.slice(-14)
  if (visible.length < 2) {
    return (
      <div className="flex h-16 items-center justify-center rounded-md border border-dashed border-border/80 bg-background/30 text-xs italic text-muted-foreground">
        Not enough recent data
      </div>
    )
  }

  const maxDuration = Math.max(...visible.map((point) => point.durationMs), 1)
  const coordinates = visible.map((point, index) => {
    const x = (index / (visible.length - 1)) * 100
    const y = 42 - (point.durationMs / maxDuration) * 34
    return `${x.toFixed(2)},${y.toFixed(2)}`
  })
  const successful = visible.every((point) => point.success)

  return (
    <svg
      aria-label="Recent service latency trend"
      className="h-16 w-full overflow-visible"
      preserveAspectRatio="none"
      role="img"
      viewBox="0 0 100 48"
    >
      <polyline
        className={successful ? 'stroke-status-up' : 'stroke-status-warn'}
        fill="none"
        points={coordinates.join(' ')}
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="3"
        vectorEffect="non-scaling-stroke"
      />
      {visible.map((point, index) => {
        if (point.success) {
          return null
        }
        const [cx, cy] = coordinates[index].split(',')
        return (
          <circle
            className="fill-status-down"
            cx={cx}
            cy={cy}
            key={`${point.monitorId}-${point.startedAt}`}
            r="2.3"
          />
        )
      })}
    </svg>
  )
}

export function ServiceOverviewCard({ service }: { service: Service }) {
  const metrics = service.cardMetrics
  const hasMetrics = metrics?.state === 'ready'
  const stateLabel = serviceMetricStateLabel(metrics, service)
  const monitors = service.monitors ?? []

  const lastCheckedAt = monitors
    .map((m) => m.status?.lastCheckedAt)
    .filter((v): v is string => Boolean(v))
    .sort()
    .pop()

  return (
    <Card
      className={cn(
        'overflow-hidden transition-colors hover:bg-surface-low/70',
        cardTone(service.rollupStatus)
      )}
    >
      <Link
        aria-label={`Open ${service.name}`}
        className="block focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
        href={`/services/${service.serviceId}`}
      >
        <CardContent className="space-y-4 pt-5">
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <span
                aria-hidden="true"
                className={cn(
                  'inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ring-1',
                  iconTileTone(service.rollupStatus)
                )}
              >
                <TechIcon category={service.serviceCategory} />
              </span>
              <p className="truncate text-lg font-semibold text-foreground">
                {service.name}
              </p>
            </div>
            <div className="flex shrink-0 flex-col items-end gap-2 pt-1">
              <StatusChip status={service.rollupStatus ?? service.lifecycleState} />
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex h-1.5 w-full items-center gap-1 overflow-hidden rounded-full bg-surface-low/50 p-0.5">
              {monitors.length > 0 ? (
                monitors.map((monitor) => {
                  const status = monitor.status?.currentStatus?.toUpperCase()
                  let color = 'bg-muted-foreground'
                  if (status === 'UP') color = 'bg-status-up'
                  else if (status === 'DEGRADED') color = 'bg-status-warn'
                  else if (status === 'DOWN') color = 'bg-status-down'

                  return (
                    <div
                      key={monitor.monitorId}
                      className={cn('h-full flex-1 rounded-sm transition-colors', color)}
                      title={`${monitor.name}: ${monitor.status?.currentStatus ?? 'Unknown'}`}
                    />
                  )
                })
              ) : (
                <div className="h-full w-full rounded-sm bg-muted-foreground/30" />
              )}
            </div>
            {stateLabel && stateLabel !== 'Pending config' ? (
              <p className="text-xs font-semibold text-muted-foreground">{stateLabel}</p>
            ) : null}
          </div>

          <dl className="grid grid-cols-3 gap-3 border-y border-border/80 py-3 text-sm">
            <div>
              <dt className="text-muted-foreground">Avg latency</dt>
              <dd className="mt-1 font-mono font-semibold text-foreground">
                {formatMetricDuration(metrics?.avgLatencyMs)}
              </dd>
            </div>
            <div className="text-center">
              <dt className="text-muted-foreground">Agg. P99</dt>
              <dd className="mt-1 font-mono font-semibold text-foreground">
                {formatMetricDuration(metrics?.p99LatencyMs)}
              </dd>
            </div>
            <div className="text-right">
              <dt className="text-muted-foreground">Recent uptime</dt>
              <dd className="mt-1 font-mono font-semibold text-foreground">
                {formatRecentUptime(metrics?.recentUptimePct)}
              </dd>
            </div>
          </dl>

          {hasMetrics ? (
            <TrendSparkline points={metrics.trend} />
          ) : (
            <div className="flex h-16 items-center justify-center rounded-md border border-dashed border-border/80 bg-background/30 text-xs italic text-muted-foreground">
              {stateLabel}
            </div>
          )}

          <div className="flex justify-end text-xs text-muted-foreground">
            {lastCheckedAt ? (
              <>
                <span className="font-medium">Last checked at</span>
                <span className="ml-1 font-mono">
                  <LocalTime iso={lastCheckedAt} />
                </span>
              </>
            ) : (
              <span>Never checked</span>
            )}
          </div>
        </CardContent>
      </Link>
    </Card>
  )
}
