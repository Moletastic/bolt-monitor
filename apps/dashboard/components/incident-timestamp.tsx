'use client'

import { useEffect, useState } from 'react'
import { isValid, parseISO } from 'date-fns'

export function IncidentTimestamp({ iso }: { iso: string }) {
  const [text, setText] = useState<string>('--')
  useEffect(() => {
    const date = parseISO(iso)
    if (!isValid(date)) {
      setText(iso)
      return
    }
    setText(
      new Intl.DateTimeFormat('en-US', {
        dateStyle: 'medium',
        timeStyle: 'short',
      }).format(date)
    )
  }, [iso])
  return <span className="font-mono">{text}</span>
}
