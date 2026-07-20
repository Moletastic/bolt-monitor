'use client'

import { useEffect, useState } from 'react'

import { Feedback } from '@/components/ui/feedback'

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

  return <Feedback tone={tone}>{message}</Feedback>
}
