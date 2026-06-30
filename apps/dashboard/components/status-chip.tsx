import { cn, formatOutcome } from '@/lib/utils'

const statusStyles: Record<string, string> = {
  UP: 'border-status-up/30 bg-status-up/15 text-status-up',
  SUCCESS: 'border-status-up/30 bg-status-up/15 text-status-up',
  DOWN: 'border-status-down/30 bg-status-down/15 text-status-down',
  FAILED: 'border-status-down/30 bg-status-down/15 text-status-down',
  DEGRADED: 'border-status-warn/30 bg-status-warn/15 text-status-warn',
  RECOVERING: 'border-status-warn/30 bg-status-warn/15 text-status-warn',
  MAINTENANCE: 'border-status-unknown/30 bg-status-unknown/15 text-status-unknown',
  UNKNOWN: 'border-status-unknown/30 bg-status-unknown/15 text-status-unknown',
}

export function StatusChip({ status }: { status?: string }) {
  const label = formatOutcome(status)
  return (
    <span
      className={cn(
        'inline-flex items-center gap-2 rounded-full border px-2.5 py-1 text-[11px] font-bold uppercase tracking-[0.2em]',
        statusStyles[label] ?? statusStyles.UNKNOWN
      )}
    >
      <span
        aria-label={`Status indicator for ${label}`}
        className="h-1.5 w-1.5 rounded-full bg-current"
        role="img"
      />
      {label}
    </span>
  )
}
