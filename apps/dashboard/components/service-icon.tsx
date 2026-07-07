import { cn } from '@/lib/utils'
import type { ServiceCategory } from '@/lib/types'
import { TechIcon } from '@/components/tech-icon'

interface ServiceIconProps {
  serviceCategory?: ServiceCategory
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

const sizeClass = {
  sm: 'h-9 w-9',
  md: 'h-11 w-11',
  lg: 'h-14 w-14',
}

export function ServiceIcon({ serviceCategory, className, size = 'md' }: ServiceIconProps) {
  return (
    <span
      className={cn(
        'inline-flex shrink-0 items-center justify-center rounded-xl bg-surface-high',
        sizeClass[size],
        className
      )}
    >
      <TechIcon category={serviceCategory} className="h-5 w-5" />
    </span>
  )
}
