'use client'

import { useState } from 'react'
import Link from 'next/link'

import { createServiceAction, updateServiceAction } from '@/lib/actions'
import {
  TECHNOLOGY_KEYS,
  type Service,
  type EscalationPolicy,
  type BusinessHoursConfig,
} from '@/lib/types'

import { SubmitButton } from '@/components/submit-button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'

const DAYS_OF_WEEK = [
  { value: 1, label: 'Mon' },
  { value: 2, label: 'Tue' },
  { value: 3, label: 'Wed' },
  { value: 4, label: 'Thu' },
  { value: 5, label: 'Fri' },
  { value: 6, label: 'Sat' },
  { value: 0, label: 'Sun' },
]

const COMMON_TIMEZONES = [
  'UTC',
  'America/Los_Angeles',
  'America/Denver',
  'America/Chicago',
  'America/New_York',
  'Europe/London',
  'Europe/Berlin',
  'Asia/Tokyo',
  'Australia/Sydney',
]

function buildBusinessHoursPayload(state: {
  enabled: boolean
  timezone: string
  startHour: number
  endHour: number
  daysOfWeek: number[]
}): BusinessHoursConfig | null {
  if (!state.enabled) {
    return null
  }
  return {
    timezone: state.timezone,
    startHour: state.startHour,
    endHour: state.endHour,
    daysOfWeek: state.daysOfWeek,
  }
}

function toggleDay(days: number[], day: number): number[] {
  return days.includes(day) ? days.filter((d) => d !== day) : [...days, day].sort((a, b) => a - b)
}

export function ServiceForm({
  service,
  policies,
  error,
}: {
  service?: Service
  policies: EscalationPolicy[]
  error?: string
}) {
  const action = service ? updateServiceAction : createServiceAction
  const initialPolicy = service?.escalationPolicyId ?? ''
  const initialBusinessHours = service?.businessHours

  const [selectedPolicyId, setSelectedPolicyId] = useState(initialPolicy)
  const [businessHours, setBusinessHours] = useState({
    enabled: Boolean(initialBusinessHours),
    timezone: initialBusinessHours?.timezone ?? 'UTC',
    startHour: initialBusinessHours?.startHour ?? 9,
    endHour: initialBusinessHours?.endHour ?? 17,
    daysOfWeek: initialBusinessHours?.daysOfWeek ?? [1, 2, 3, 4, 5],
  })

  const businessHoursPayload = buildBusinessHoursPayload(businessHours)

  return (
    <Card>
      <CardHeader>
        <CardTitle>{service ? 'Edit service' : 'Create service'}</CardTitle>
      </CardHeader>
      <CardContent>
        <form action={action} className="grid gap-5">
          {service ? <input name="serviceId" type="hidden" value={service.serviceId} /> : null}
          <input
            name="returnTo"
            type="hidden"
            value={service ? `/services/${service.serviceId}` : '/services/new'}
          />
          <input name="escalationPolicyId" type="hidden" value={selectedPolicyId} />
          <input
            name="businessHoursPayload"
            type="hidden"
            value={businessHoursPayload ? JSON.stringify(businessHoursPayload) : ''}
          />
          <div className="grid gap-5 lg:grid-cols-2">
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Name</span>
              <Input
                defaultValue={service?.name ?? ''}
                name="name"
                placeholder="Payments API"
                required
              />
            </label>
            <label className="grid gap-2 text-sm text-muted-foreground lg:col-span-2">
              <span className="font-semibold text-foreground">Description</span>
              <textarea
                className="min-h-24 rounded-md border border-border bg-surface-low px-3 py-2 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                defaultValue={service?.description ?? ''}
                name="description"
                placeholder="Optional operator-facing notes for this service"
              />
            </label>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Primary technology</span>
              <Select defaultValue={service?.technologyKey ?? ''} name="technologyKey">
                <option value="">None</option>
                {TECHNOLOGY_KEYS.map((technologyKey) => (
                  <option key={technologyKey} value={technologyKey}>
                    {technologyKey}
                  </option>
                ))}
              </Select>
            </label>
          </div>

          <section className="grid gap-4 rounded-lg border border-border bg-surface-low p-4">
            <header className="flex items-center justify-between">
              <div>
                <h3 className="text-sm font-semibold text-foreground">Notification route</h3>
                <p className="text-xs text-muted-foreground">
                  Routes incidents through the selected on-call path.
                </p>
              </div>
            </header>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-medium text-foreground">Notification route</span>
              <Select
                onChange={(event) => setSelectedPolicyId(event.target.value)}
                value={selectedPolicyId}
              >
                <option value="">None</option>
                {policies.map((policy) => (
                  <option key={policy.policyId} value={policy.policyId}>
                    {policy.name}
                  </option>
                ))}
              </Select>
              {policies.length === 0 ? (
                <p className="text-xs text-muted-foreground">
                  No notification routes exist yet.{' '}
                  <Link className="text-primary hover:underline" href="/policies/new">
                    Create one
                  </Link>{' '}
                  before assigning one to a service.
                </p>
              ) : null}
            </label>
          </section>

          <section className="grid gap-4 rounded-lg border border-border bg-surface-low p-4">
            <header className="flex items-center justify-between">
              <div>
                <h3 className="text-sm font-semibold text-foreground">Business hours</h3>
                <p className="text-xs text-muted-foreground">
                  Define when the on-call path follows businessHours vs offHoursPath.
                </p>
              </div>
              <label className="inline-flex items-center gap-2 text-sm">
                <input
                  checked={businessHours.enabled}
                  onChange={(event) =>
                    setBusinessHours((prev) => ({ ...prev, enabled: event.target.checked }))
                  }
                  type="checkbox"
                />
                <span>Enabled</span>
              </label>
            </header>
            {businessHours.enabled ? (
              <div className="grid gap-4 md:grid-cols-2">
                <label className="grid gap-2 text-sm text-muted-foreground">
                  <span className="font-medium text-foreground">Timezone</span>
                  <Select
                    onChange={(event) =>
                      setBusinessHours((prev) => ({ ...prev, timezone: event.target.value }))
                    }
                    value={businessHours.timezone}
                  >
                    {COMMON_TIMEZONES.map((timezone) => (
                      <option key={timezone} value={timezone}>
                        {timezone}
                      </option>
                    ))}
                  </Select>
                </label>
                <label className="grid gap-2 text-sm text-muted-foreground">
                  <span className="font-medium text-foreground">Window</span>
                  <div className="flex items-center gap-2">
                    <Input
                      max={24}
                      min={0}
                      onChange={(event) =>
                        setBusinessHours((prev) => ({
                          ...prev,
                          startHour: Number(event.target.value || 0),
                        }))
                      }
                      type="number"
                      value={businessHours.startHour}
                    />
                    <span className="text-xs text-muted-foreground">to</span>
                    <Input
                      max={24}
                      min={0}
                      onChange={(event) =>
                        setBusinessHours((prev) => ({
                          ...prev,
                          endHour: Number(event.target.value || 0),
                        }))
                      }
                      type="number"
                      value={businessHours.endHour}
                    />
                  </div>
                </label>
                <div className="md:col-span-2">
                  <p className="mb-2 text-sm font-medium text-foreground">Days of week</p>
                  <div className="flex flex-wrap gap-2">
                    {DAYS_OF_WEEK.map((day) => {
                      const active = businessHours.daysOfWeek.includes(day.value)
                      return (
                        <button
                          key={day.value}
                          type="button"
                          onClick={() =>
                            setBusinessHours((prev) => ({
                              ...prev,
                              daysOfWeek: toggleDay(prev.daysOfWeek, day.value),
                            }))
                          }
                          className={`rounded-md border px-3 py-1.5 text-xs font-medium transition-colors ${
                            active
                              ? 'border-primary/40 bg-primary/10 text-primary'
                              : 'border-border bg-transparent text-muted-foreground hover:bg-surface-low hover:text-foreground'
                          }`}
                        >
                          {day.label}
                        </button>
                      )
                    })}
                  </div>
                </div>
              </div>
            ) : null}
          </section>

          {error ? (
            <p className="rounded-md border border-status-down/30 bg-status-down/10 px-3 py-2 text-sm text-status-down">
              {error}
            </p>
          ) : null}
          <div className="flex items-center justify-end gap-3">
            <SubmitButton>{service ? 'Save changes' : 'Create service'}</SubmitButton>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
