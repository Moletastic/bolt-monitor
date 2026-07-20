import * as React from 'react'

import { cn } from '@/lib/utils'

const Select = React.forwardRef<HTMLSelectElement, React.SelectHTMLAttributes<HTMLSelectElement>>(
  ({ className, 'aria-invalid': ariaInvalid, ...props }, ref) => {
    return (
      <select
        className={cn(
          'flex h-10 w-full rounded border border-border bg-surface-low px-3 py-2 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50 aria-[invalid=true]:border-dashed aria-[invalid=true]:border-destructive aria-[invalid=true]:bg-destructive/5',
          className
        )}
        aria-invalid={ariaInvalid}
        ref={ref}
        {...props}
      />
    )
  }
)
Select.displayName = 'Select'

export { Select }
