'use client'

import { useEffect, useState } from 'react'

export const samePageActionStartEvent = 'dashboard:same-page-action-start'

export function QueryFeedbackBanner({
  message,
  tone,
}: {
  message: string
  tone: 'error' | 'success'
}) {
  const [visible, setVisible] = useState(true)

  useEffect(() => {
    const hide = () => setVisible(false)
    window.addEventListener(samePageActionStartEvent, hide)
    return () => window.removeEventListener(samePageActionStartEvent, hide)
  }, [])

  if (!visible) return null

  return (
    <p
      className={`rounded-md border px-3 py-2 text-sm ${tone === 'error' ? 'border-status-down/30 bg-status-down/10 text-status-down' : 'border-status-up/30 bg-status-up/10 text-status-up'}`}
      role={tone === 'error' ? 'alert' : 'status'}
    >
      {message}
    </p>
  )
}
