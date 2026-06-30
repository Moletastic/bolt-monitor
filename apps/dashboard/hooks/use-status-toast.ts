'use client'

import { useEffect, useRef } from 'react'
import { toast } from '@/hooks/use-toast'

interface UseStatusToastOptions {
  serviceName?: string
}

export function useStatusToast(
  currentStatus: string | null | undefined,
  options: UseStatusToastOptions = {}
) {
  const prevStatusRef = useRef<string | null | undefined>(currentStatus)

  useEffect(() => {
    const prevStatus = prevStatusRef.current

    if (prevStatus && prevStatus !== currentStatus) {
      if (currentStatus === 'DOWN') {
        toast({
          title: options.serviceName ? `${options.serviceName} is DOWN` : 'Service is DOWN',
          description: options.serviceName,
          variant: 'destructive',
        })
      } else if (currentStatus === 'UP' && prevStatus === 'DOWN') {
        toast({
          title: options.serviceName ? `${options.serviceName} is UP again` : 'Service is UP again',
          variant: 'default',
        })
      }
    }

    prevStatusRef.current = currentStatus
  }, [currentStatus, options.serviceName])
}
