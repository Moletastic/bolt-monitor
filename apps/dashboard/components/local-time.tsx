'use client'

import { format, isValid, parseISO } from 'date-fns'
import { useEffect, useState } from 'react'

export function LocalTime({ iso }: { iso: string }) {
  const [time, setTime] = useState<string | null>(null)

  useEffect(() => {
    const date = parseISO(iso)
    if (isValid(date)) {
      setTime(format(date, 'p'))
    }
  }, [iso])

  return <span className="font-mono">{time ?? '--:--'}</span>
}
