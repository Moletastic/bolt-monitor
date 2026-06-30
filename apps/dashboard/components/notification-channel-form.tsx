'use client'

import { useState } from 'react'

import { SubmitButton } from '@/components/submit-button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { createNotificationChannelAction, updateNotificationChannelAction } from '@/lib/actions'
import type { EscalationChannelType, NotificationChannel } from '@/lib/types'

const CHANNEL_TYPES: {
  value: EscalationChannelType
  label: string
  targetLabel: string
  fields: { key: string; label: string; type?: string }[]
}[] = [
  {
    value: 'telegram',
    label: 'Telegram',
    targetLabel: 'Chat ID',
    fields: [{ key: 'botToken', label: 'Bot token', type: 'password' }],
  },
  {
    value: 'email',
    label: 'Email',
    targetLabel: 'Recipient email',
    fields: [
      { key: 'apiKey', label: 'Provider API key', type: 'password' },
      { key: 'fromEmail', label: 'From address' },
      { key: 'apiBaseUrl', label: 'API base URL' },
    ],
  },
  {
    value: 'sms',
    label: 'SMS (Twilio)',
    targetLabel: 'Destination number',
    fields: [
      { key: 'accountSid', label: 'Account SID' },
      { key: 'authToken', label: 'Auth token', type: 'password' },
      { key: 'fromNumber', label: 'From number' },
      { key: 'apiBaseUrl', label: 'API base URL' },
    ],
  },
  { value: 'webhook', label: 'Webhook', targetLabel: 'Webhook URL', fields: [] },
  {
    value: 'pagerduty',
    label: 'PagerDuty',
    targetLabel: 'Routing key',
    fields: [{ key: 'routingKey', label: 'Routing key' }],
  },
]

export function NotificationChannelForm({ channel }: { channel?: NotificationChannel }) {
  const [type, setType] = useState<EscalationChannelType>(channel?.type ?? 'telegram')
  const metadata = CHANNEL_TYPES.find((entry) => entry.value === type) ?? CHANNEL_TYPES[0]
  const action = channel ? updateNotificationChannelAction : createNotificationChannelAction

  return (
    <form action={action} className="grid gap-5">
      {channel ? <input name="channelId" type="hidden" value={channel.channelId} /> : null}
      <input
        name="returnTo"
        type="hidden"
        value={
          channel ? `/integrations/channels/${channel.channelId}` : '/integrations/channels/new'
        }
      />
      <label className="grid gap-2 text-sm text-muted-foreground">
        <span className="font-semibold text-foreground">Name</span>
        <Input
          defaultValue={channel?.name ?? ''}
          name="name"
          placeholder="Payments on-call"
          required
        />
        <span className="text-xs">What you will recognize it as</span>
      </label>
      <label className="grid gap-2 text-sm text-muted-foreground">
        <span className="font-semibold text-foreground">Type</span>
        <Select
          name="type"
          onChange={(event) => setType(event.target.value as EscalationChannelType)}
          value={type}
        >
          {CHANNEL_TYPES.map((entry) => (
            <option key={entry.value} value={entry.value}>
              {entry.label}
            </option>
          ))}
        </Select>
      </label>
      <label className="grid gap-2 text-sm text-muted-foreground">
        <span className="font-semibold text-foreground">Target</span>
        <Input
          defaultValue={channel?.target ?? ''}
          name="target"
          placeholder={metadata.targetLabel}
          required
        />
        <span className="text-xs">Where this channel delivers to</span>
      </label>
      {metadata.fields.map((field) => {
        const raw = channel?.config?.[field.key]
        const value =
          typeof raw === 'string' && raw === '***REDACTED***'
            ? '••••••'
            : typeof raw === 'string'
              ? raw
              : ''
        return (
          <label className="grid gap-2 text-sm text-muted-foreground" key={field.key}>
            <span className="font-semibold text-foreground">{field.label}</span>
            <Input defaultValue={value} name={`config.${field.key}`} type={field.type ?? 'text'} />
          </label>
        )
      })}
      <div className="flex justify-end">
        <SubmitButton>{channel ? 'Save changes' : 'Create channel'}</SubmitButton>
      </div>
    </form>
  )
}
