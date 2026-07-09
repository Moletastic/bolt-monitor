import { cn } from '@/lib/utils'
import type { ServiceCategory } from '@/lib/types'
import { TechIcon } from '@/components/tech-icon'

export type ServiceIconTone = 'default' | 'up' | 'down' | 'degraded' | 'unknown'

interface ServiceIconProps {
  serviceCategory?: ServiceCategory
  className?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  tone?: ServiceIconTone
}

const sizeClass = {
  sm: 'h-9 w-9',
  md: 'h-11 w-11',
  lg: 'h-14 w-14',
  xl: 'h-16 w-16',
}

const iconSizeClass = {
  sm: 'h-5 w-5',
  md: 'h-5 w-5',
  lg: 'h-6 w-6',
  xl: 'h-16 w-16',
}

const toneClass: Record<ServiceIconTone, string> = {
  default: 'bg-surface-high text-foreground',
  up: 'bg-status-up/15 text-status-up',
  down: 'bg-status-down/15 text-status-down',
  degraded: 'bg-status-warn/15 text-status-warn',
  unknown: 'bg-status-unknown/15 text-status-unknown',
}

export function ServiceIcon({
  serviceCategory,
  className,
  size = 'md',
  tone = 'default',
}: ServiceIconProps) {
  return (
    <span
      className={cn(
        'inline-flex shrink-0 items-center justify-center rounded-xl',
        toneClass[tone],
        sizeClass[size],
        className
      )}
    >
      <TechIcon category={serviceCategory} className={iconSizeClass[size]} />
    </span>
  )
}
