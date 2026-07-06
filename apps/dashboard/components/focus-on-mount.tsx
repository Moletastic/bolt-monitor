'use client'

import { useEffect, useRef, type ReactNode } from 'react'

type FocusOnMountProps = {
  active: boolean
  children: ReactNode
  className?: string
}

export function FocusOnMount({ active, children, className }: FocusOnMountProps) {
  const ref = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!active) return
    ref.current?.focus()
  }, [active])

  return (
    <div className={className} ref={ref} tabIndex={active ? -1 : undefined}>
      {children}
    </div>
  )
}
