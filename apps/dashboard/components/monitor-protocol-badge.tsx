import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import type { MonitorType } from '@/lib/types'

interface MonitorProtocolBadgeProps {
  type: MonitorType
  className?: string
  index?: number
  status?: string
}

const PROTOCOL_LABELS: Record<MonitorType, string> = {
  http: 'HTTP',
  tcp: 'TCP',
  grpc: 'gRPC',
  dns: 'DNS',
}

const STATUS_TONE: Record<string, string> = {
  UP: 'text-status-up',
  SUCCESS: 'text-status-up',
  DOWN: 'text-status-down',
  FAILED: 'text-status-down',
  DEGRADED: 'text-status-warn',
  RECOVERING: 'text-status-warn',
  MAINTENANCE: 'text-muted-foreground',
  UNKNOWN: 'text-muted-foreground',
}

function badgeTone(status?: string) {
  if (!status) {
    return 'text-foreground/80'
  }
  return STATUS_TONE[status.toUpperCase()] ?? 'text-foreground/80'
}

export function MonitorProtocolBadge({
  type,
  className,
  index,
  status,
}: MonitorProtocolBadgeProps) {
  const label = index === undefined ? PROTOCOL_LABELS[type] : `M${index} · ${PROTOCOL_LABELS[type]}`
  return (
    <span
      className={twMerge(
        clsx(
          'inline-flex items-center rounded-md bg-surface-high px-2 py-0.5 text-xs font-medium',
          badgeTone(status),
          className
        )
      )}
    >
      {label}
    </span>
  )
}
