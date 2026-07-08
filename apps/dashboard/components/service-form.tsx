'use client'

import { useState, type ReactNode } from 'react'
import Link from 'next/link'

import { createServiceAction, updateServiceAction } from '@/lib/actions'
import {
  SERVICE_CATEGORIES,
  type Service,
  type EscalationPolicy,
  type BusinessHoursConfig,
  type ServiceCategory,
} from '@/lib/types'

import { SubmitButton } from '@/components/submit-button'
import { TechIcon } from '@/components/tech-icon'
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

function formatCategoryLabel(category: ServiceCategory) {
  return category === 'http' ? 'HTTP' : category.charAt(0).toUpperCase() + category.slice(1)
}

function SectionIcon({ children }: { children: ReactNode }) {
  return (
    <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary">
      {children}
    </span>
  )
}

function BellIcon({ className }: { className: string }) {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.6"
      viewBox="0 0 24 24"
    >
      <path d="M18 8a6 6 0 0 0-12 0c0 7-3 8-3 8h18s-3-1-3-8" />
      <path d="M13.7 21a2 2 0 0 1-3.4 0" />
    </svg>
  )
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
  const [selectedCategory, setSelectedCategory] = useState<ServiceCategory | ''>(
    service ? (service.serviceCategory ?? '') : 'server'
  )
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
      {service ? (
        <CardHeader>
          <CardTitle>Edit service</CardTitle>
        </CardHeader>
      ) : null}
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
          <div className="grid gap-5 xl:grid-cols-2">
            <section className="grid content-start gap-5 rounded-lg border border-border bg-surface-low p-4">
              <header className="flex items-center gap-3">
                <SectionIcon>
                  <TechIcon className="h-5 w-5" />
                </SectionIcon>
                <div>
                  <h3 className="text-sm font-semibold text-foreground">Service identity</h3>
                  <p className="text-xs text-muted-foreground">
                    Choose the service icon and name operators will scan first.
                  </p>
                </div>
              </header>
              <div className="grid gap-3 md:grid-cols-[12rem_1fr]">
                <fieldset className="grid gap-2 text-sm text-muted-foreground">
                  <legend className="sr-only">Service icon</legend>
                  <div className="relative h-10 rounded-md border border-border bg-background">
                    <span className="pointer-events-none absolute left-2 top-1/2 inline-flex h-6 w-6 -translate-y-1/2 items-center justify-center rounded-md bg-surface-high text-foreground">
                      <TechIcon category={selectedCategory || undefined} className="h-5 w-5" />
                    </span>
                    <Select
                      aria-label="Service icon"
                      className="h-full w-full cursor-pointer appearance-none bg-transparent pl-10 pr-2 text-foreground"
                      name="serviceCategory"
                      onChange={(event) =>
                        setSelectedCategory(event.target.value as ServiceCategory | '')
                      }
                      value={selectedCategory}
                    >
                      <option className="bg-background text-foreground" value="">
                        None
                      </option>
                      {SERVICE_CATEGORIES.map((category) => (
                        <option
                          className="bg-background text-foreground"
                          key={category}
                          value={category}
                        >
                          {formatCategoryLabel(category)}
                        </option>
                      ))}
                    </Select>
                  </div>
                </fieldset>
                <label className="grid text-sm text-muted-foreground">
                  <span className="sr-only">Name</span>
                  <Input
                    aria-label="Name"
                    defaultValue={service?.name ?? ''}
                    name="name"
                    placeholder="Payments API"
                    required
                  />
                </label>
                <label className="grid text-sm text-muted-foreground md:col-span-2">
                  <span className="sr-only">Description</span>
                  <textarea
                    aria-label="Description"
                    className="min-h-24 rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    defaultValue={service?.description ?? ''}
                    name="description"
                    placeholder="Optional operator-facing notes for this service"
                  />
                </label>
              </div>
            </section>

            <section className="grid content-start gap-5 rounded-lg border border-border bg-surface-low p-4">
              <header className="flex items-center gap-3">
                <SectionIcon>
                  <BellIcon className="h-5 w-5" />
                </SectionIcon>
                <div>
                  <h3 className="text-sm font-semibold text-foreground">Notifications</h3>
                  <p className="text-xs text-muted-foreground">
                    Route incidents and optionally limit business-hours handling.
                  </p>
                </div>
              </header>
              <label className="grid text-sm text-muted-foreground">
                <span className="sr-only">Notification route</span>
                <Select
                  aria-label="Notification route"
                  onChange={(event) => setSelectedPolicyId(event.target.value)}
                  value={selectedPolicyId}
                >
                  <option className="bg-background text-foreground" value="">
                    None
                  </option>
                  {policies.map((policy) => (
                    <option
                      className="bg-background text-foreground"
                      key={policy.policyId}
                      value={policy.policyId}
                    >
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
              <div className="grid gap-4 rounded-lg border border-border bg-background p-4">
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <h4 className="text-sm font-semibold text-foreground">Business hours</h4>
                    <p className="text-xs text-muted-foreground">
                      Enable when notification routing should distinguish business hours from
                      off-hours paths.
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
                </div>
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
              </div>
            </section>
          </div>

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
