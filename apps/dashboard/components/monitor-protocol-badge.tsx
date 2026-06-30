import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import type { MonitorType } from '@/lib/types'

interface MonitorProtocolBadgeProps {
  type: MonitorType
  className?: string
}

const PROTOCOL_LABELS: Record<MonitorType, string> = {
  http: 'HTTP',
  tcp: 'TCP',
  grpc: 'gRPC',
  dns: 'DNS',
}

export function MonitorProtocolBadge({ type, className }: MonitorProtocolBadgeProps) {
  return (
    <span
      className={twMerge(
        clsx(
          'inline-flex items-center rounded-md bg-muted px-2 py-0.5 text-xs font-medium text-foreground',
          className
        )
      )}
    >
      {PROTOCOL_LABELS[type]}
    </span>
  )
}
