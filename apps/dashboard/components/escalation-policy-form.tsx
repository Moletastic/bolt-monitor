'use client'

import { useState } from 'react'
import { Plus, Trash2 } from 'lucide-react'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { SubmitButton } from '@/components/submit-button'
import type {
  BusinessHoursConfig,
  EscalationPath,
  EscalationStep,
  NotificationChannel,
} from '@/lib/types'
import { createEscalationPolicyAction, updateEscalationPolicyAction } from '@/lib/actions'

const DAY_LABELS = [
  { value: 1, label: 'Mon' },
  { value: 2, label: 'Tue' },
  { value: 3, label: 'Wed' },
  { value: 4, label: 'Thu' },
  { value: 5, label: 'Fri' },
  { value: 6, label: 'Sat' },
  { value: 7, label: 'Sun' },
]

const TIMEZONE_OPTIONS = [
  'UTC',
  'America/New_York',
  'America/Los_Angeles',
  'America/Chicago',
  'Europe/London',
  'Europe/Berlin',
  'Europe/Paris',
  'Asia/Tokyo',
  'Asia/Singapore',
  'Australia/Sydney',
]

function defaultBusinessHours(): BusinessHoursConfig {
  return { timezone: 'UTC', startHour: 9, endHour: 17, daysOfWeek: [1, 2, 3, 4, 5] }
}

function defaultPath(channelId = ''): EscalationPath {
  return { steps: [{ channelId, delayMinutes: 0 }] }
}

function clonePath(path: EscalationPath): EscalationPath {
  return {
    steps: path.steps.map((step) => ({
      channelId: step.channelId,
      delayMinutes: step.delayMinutes,
    })),
  }
}

interface EscalationPolicyFormProps {
  mode: 'create' | 'edit'
  policyId?: string
  initialName?: string
  initialDescription?: string
  initialBusinessHoursPath?: EscalationPath
  initialOffHoursPath?: EscalationPath
  initialBusinessHours?: BusinessHoursConfig
  availableChannels: NotificationChannel[]
  returnTo: string
  errorHref: string
  searchParams: Promise<{ error?: string }>
}

export function EscalationPolicyForm({
  mode,
  policyId,
  initialName = '',
  initialDescription = '',
  initialBusinessHoursPath,
  initialOffHoursPath,
  initialBusinessHours,
  availableChannels,
  returnTo,
  errorHref,
  searchParams,
}: EscalationPolicyFormProps) {
  const firstChannelId = availableChannels[0]?.channelId ?? ''
  const [name, setName] = useState(initialName)
  const [description, setDescription] = useState(initialDescription)
  const [businessHoursPath, setBusinessHoursPath] = useState<EscalationPath>(
    initialBusinessHoursPath ?? defaultPath(firstChannelId)
  )
  const [offHoursPath, setOffHoursPath] = useState<EscalationPath>(
    initialOffHoursPath ?? clonePath(initialBusinessHoursPath ?? defaultPath(firstChannelId))
  )
  const [businessHours, setBusinessHours] = useState<BusinessHoursConfig>(
    initialBusinessHours ?? defaultBusinessHours()
  )
  const [businessHoursEnabled, setBusinessHoursEnabled] = useState(Boolean(initialBusinessHours))
  const [inlineError, setInlineError] = useState<string | null>(null)

  return (
    <form
      action={mode === 'edit' ? updateEscalationPolicyAction : createEscalationPolicyAction}
      className="grid gap-6"
      onSubmit={(event) => {
        const paths = [businessHoursPath, businessHoursEnabled ? offHoursPath : businessHoursPath]
        for (const path of paths) {
          const missingIndex = path.steps.findIndex((step) => !step.channelId)
          if (missingIndex >= 0) {
            event.preventDefault()
            setInlineError(`Pick a channel for step ${missingIndex + 1}`)
            return
          }
        }
        const form = event.currentTarget
        const businessHoursPayload = form.elements.namedItem('businessHoursPathPayload')
        const offHoursPayload = form.elements.namedItem('offHoursPathPayload')
        if (!(businessHoursPayload instanceof HTMLInputElement)) {
          event.preventDefault()
          setInlineError('Business hours payload field was not found.')
          return
        }
        if (!(offHoursPayload instanceof HTMLInputElement)) {
          event.preventDefault()
          setInlineError('Off-hours payload field was not found.')
          return
        }
        businessHoursPayload.value = JSON.stringify(businessHoursPath)
        offHoursPayload.value = JSON.stringify(
          businessHoursEnabled ? offHoursPath : clonePath(businessHoursPath)
        )
      }}
    >
      {policyId ? <input name="policyId" type="hidden" value={policyId} /> : null}
      <input name="mode" type="hidden" value={mode} />
      <input name="returnTo" type="hidden" value={returnTo} />
      <input name="errorHref" type="hidden" value={errorHref} />
      <input name="businessHoursPathPayload" type="hidden" defaultValue="" />
      <input name="offHoursPathPayload" type="hidden" defaultValue="" />

      <ErrorBanner errorHref={errorHref} searchParams={searchParams} />
      {inlineError ? (
        <p className="rounded-md border border-status-down/30 bg-status-down/10 px-3 py-2 text-sm text-status-down">
          {inlineError}
        </p>
      ) : null}
      {availableChannels.length === 0 ? (
        <p className="rounded-md border border-status-warn/30 bg-status-warn/10 px-3 py-2 text-sm text-status-warn">
          Create a notification channel before saving a route.
        </p>
      ) : null}

      <div className="grid gap-4 lg:grid-cols-2">
        <label className="grid gap-2 text-sm text-muted-foreground">
          <span className="font-semibold text-foreground">Name</span>
          <Input
            name="name"
            onChange={(event) => setName(event.target.value)}
            required
            value={name}
          />
        </label>
        <label className="grid gap-2 text-sm text-muted-foreground">
          <span className="font-semibold text-foreground">Description</span>
          <Input
            name="description"
            onChange={(event) => setDescription(event.target.value)}
            value={description}
          />
        </label>
      </div>

      <BusinessHoursCard
        businessHours={businessHours}
        enabled={businessHoursEnabled}
        onChange={setBusinessHours}
        onToggle={setBusinessHoursEnabled}
      />

      <div className="grid gap-6 xl:grid-cols-2">
        <PathEditor
          channels={availableChannels}
          label="Business hours path"
          path={businessHoursPath}
          onChange={setBusinessHoursPath}
        />
        {businessHoursEnabled ? (
          <PathEditor
            channels={availableChannels}
            label="Off-hours path"
            path={offHoursPath}
            onChange={setOffHoursPath}
          />
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Off-hours path</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">
                Off-hours path mirrors business hours until you enable business hours branching.
              </p>
            </CardContent>
          </Card>
        )}
      </div>

      <div className="flex items-center justify-end gap-3">
        <SubmitButton>{mode === 'create' ? 'Create route' : 'Save changes'}</SubmitButton>
      </div>
    </form>
  )
}

function ErrorBanner({
  errorHref,
  searchParams,
}: {
  errorHref: string
  searchParams: Promise<{ error?: string }>
}) {
  const [error, setError] = useState<string | null>(null)
  searchParams.then((query) => {
    if (query.error) setError(query.error)
  })
  if (!error) return null
  return (
    <p className="rounded-md border border-status-down/30 bg-status-down/10 px-3 py-2 text-sm text-status-down">
      {error}{' '}
      <a className="underline" href={errorHref}>
        clear
      </a>
    </p>
  )
}

function BusinessHoursCard({
  businessHours,
  enabled,
  onChange,
  onToggle,
}: {
  businessHours: BusinessHoursConfig
  enabled: boolean
  onChange: (value: BusinessHoursConfig) => void
  onToggle: (value: boolean) => void
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Business hours</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <label className="flex items-center gap-2 text-sm">
          <input
            checked={enabled}
            name="businessHoursEnabled"
            onChange={(event) => onToggle(event.target.checked)}
            type="checkbox"
          />
          <span className="font-semibold text-foreground">
            Use different steps for business hours and off hours
          </span>
        </label>
        {enabled ? (
          <div className="grid gap-4 lg:grid-cols-2">
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Timezone</span>
              <Select
                onChange={(event) => onChange({ ...businessHours, timezone: event.target.value })}
                value={businessHours.timezone}
              >
                {TIMEZONE_OPTIONS.map((tz) => (
                  <option key={tz} value={tz}>
                    {tz}
                  </option>
                ))}
              </Select>
            </label>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Days of week</span>
              <div className="flex flex-wrap gap-2">
                {DAY_LABELS.map((day) => {
                  const active = businessHours.daysOfWeek.includes(day.value)
                  return (
                    <label
                      className={`cursor-pointer rounded-md border px-2.5 py-1 text-xs font-semibold ${active ? 'border-primary/40 bg-primary/10 text-primary' : 'border-border bg-surface-low text-muted-foreground'}`}
                      key={day.value}
                    >
                      <input
                        checked={active}
                        className="sr-only"
                        onChange={(event) =>
                          onChange({
                            ...businessHours,
                            daysOfWeek: event.target.checked
                              ? Array.from(new Set([...businessHours.daysOfWeek, day.value])).sort()
                              : businessHours.daysOfWeek.filter((value) => value !== day.value),
                          })
                        }
                        type="checkbox"
                      />
                      {day.label}
                    </label>
                  )
                })}
              </div>
            </label>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Start hour</span>
              <Input
                max={23}
                min={0}
                onChange={(event) =>
                  onChange({ ...businessHours, startHour: Number(event.target.value) })
                }
                type="number"
                value={businessHours.startHour}
              />
            </label>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">End hour</span>
              <Input
                max={24}
                min={0}
                onChange={(event) =>
                  onChange({ ...businessHours, endHour: Number(event.target.value) })
                }
                type="number"
                value={businessHours.endHour}
              />
            </label>
          </div>
        ) : null}
      </CardContent>
    </Card>
  )
}

function PathEditor({
  label,
  path,
  channels,
  onChange,
}: {
  label: string
  path: EscalationPath
  channels: NotificationChannel[]
  onChange: (path: EscalationPath) => void
}) {
  function updateStep(index: number, next: EscalationStep) {
    onChange({ steps: path.steps.map((step, idx) => (idx === index ? next : step)) })
  }
  return (
    <Card>
      <CardHeader>
        <CardTitle>{label}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {path.steps.map((step, index) => (
          <div className="space-y-3 rounded-lg border border-border bg-surface-low p-4" key={index}>
            <div className="flex items-center justify-between">
              <p className="text-sm font-semibold text-foreground">Step {index + 1}</p>
              {path.steps.length > 1 ? (
                <button
                  className="text-xs font-semibold text-status-down hover:underline"
                  onClick={() => onChange({ steps: path.steps.filter((_, idx) => idx !== index) })}
                  type="button"
                >
                  <Trash2 className="mr-1 inline h-3.5 w-3.5" /> Remove step
                </button>
              ) : null}
            </div>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Channel</span>
              <Select
                onChange={(event) => updateStep(index, { ...step, channelId: event.target.value })}
                required
                value={step.channelId}
              >
                <option value="">Pick a channel</option>
                {channels.map((channel) => (
                  <option key={channel.channelId} value={channel.channelId}>
                    {channel.name} · {channel.type} · {channel.target}
                  </option>
                ))}
              </Select>
            </label>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Delay before firing</span>
              <Input
                max={1440}
                min={0}
                onChange={(event) =>
                  updateStep(index, { ...step, delayMinutes: Number(event.target.value) })
                }
                type="number"
                value={step.delayMinutes}
              />
            </label>
          </div>
        ))}
        <button
          className="inline-flex items-center gap-1 text-sm font-semibold text-primary hover:underline"
          onClick={() =>
            onChange({
              steps: [...path.steps, { channelId: channels[0]?.channelId ?? '', delayMinutes: 0 }],
            })
          }
          type="button"
        >
          <Plus className="h-4 w-4" /> Add step
        </button>
      </CardContent>
    </Card>
  )
}
