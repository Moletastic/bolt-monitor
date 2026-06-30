'use client'

import { Suspense, useEffect, useRef } from 'react'
import { useSearchParams } from 'next/navigation'
import { toast } from '@/hooks/use-toast'

function ToastWatcherInner() {
  const searchParams = useSearchParams()
  const hasShownToast = useRef(false)

  useEffect(() => {
    if (hasShownToast.current) return
    if (typeof window === 'undefined') return

    const created = searchParams.get('created')
    const updated = searchParams.get('updated')
    const error = searchParams.get('error')
    const run = searchParams.get('run')

    if (created) {
      toast({ title: 'Created successfully', variant: 'default' })
      hasShownToast.current = true
    } else if (updated) {
      toast({ title: 'Updated successfully', variant: 'default' })
      hasShownToast.current = true
    } else if (error) {
      toast({ title: error, variant: 'destructive' })
      hasShownToast.current = true
    } else if (run) {
      toast({ title: 'Manual run triggered', variant: 'default' })
      hasShownToast.current = true
    }
  }, [searchParams])

  return null
}

export function ToastWatcher() {
  return (
    <Suspense fallback={null}>
      <ToastWatcherInner />
    </Suspense>
  )
}
