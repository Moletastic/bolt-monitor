import { Mail, MessageSquare, Send, Siren, Webhook } from 'lucide-react'

import type { EscalationChannelType } from '@/lib/types'

const iconMap = {
  telegram: Send,
  email: Mail,
  sms: MessageSquare,
  webhook: Webhook,
  pagerduty: Siren,
} as const

export function ChannelTypeIcon({ type }: { type: EscalationChannelType }) {
  const Icon = iconMap[type] ?? Webhook
  return <Icon aria-hidden="true" className="h-4 w-4 text-muted-foreground" />
}
