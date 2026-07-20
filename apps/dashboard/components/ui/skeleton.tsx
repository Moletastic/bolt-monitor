import * as React from 'react'

import { cn } from '@/lib/utils'

export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  height?: string | number
  width?: string | number
}

export function Skeleton({ className, height, width, style, ...props }: SkeletonProps) {
  return (
    <div
      aria-hidden="true"
      className={cn('animate-pulse rounded bg-surface-high/70', className)}
      style={{ height, width, ...style }}
      {...props}
    />
  )
}
