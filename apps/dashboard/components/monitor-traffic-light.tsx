import { cn } from '@/lib/utils'
import type { Monitor } from '@/lib/types'

const statusClass: Record<string, string> = {
  UP: 'bg-status-up',
  DEGRADED: 'bg-status-warn',
  DOWN: 'bg-status-down',
  MAINTENANCE: 'bg-muted-foreground',
}

const unknownClass = 'bg-muted-foreground/50'

export function MonitorTrafficLights({
  monitors,
  maxVisible = 12,
}: {
  monitors: Pick<Monitor, 'monitorId' | 'name' | 'status'>[]
  maxVisible?: number
}) {
  if (monitors.length === 0) {
    return null
  }

  const visible = monitors.slice(0, maxVisible)
  const overflow = monitors.length - visible.length

  return (
    <div
      aria-label={`${monitors.length} child monitors`}
      className="flex flex-wrap items-center gap-1"
      role="img"
    >
      {visible.map((monitor) => {
        const status = monitor.status?.currentStatus ?? ''
        const tone = statusClass[status] ?? unknownClass
        const label = status ? `${monitor.name}: ${status}` : monitor.name
        return (
          <span
            aria-label={label}
            className={cn('h-2 w-2 rounded-full', tone)}
            key={monitor.monitorId}
            title={label}
          />
        )
      })}
      {overflow > 0 ? (
        <span
          aria-label={`${overflow} more monitors`}
          className="font-mono text-[10px] font-semibold text-muted-foreground"
        >
          +{overflow}
        </span>
      ) : null}
    </div>
  )
}
