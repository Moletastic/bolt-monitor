'use client'

import { useState, type ReactNode } from 'react'
import { CheckCircle2, Fingerprint, Plus, Send, Trash2, X, type LucideIcon } from 'lucide-react'

import { createMonitorAction, updateMonitorAction } from '@/lib/actions'
import type { Monitor } from '@/lib/types'
import { cn, monitorCadenceOptions } from '@/lib/utils'

import { SubmitButton } from '@/components/submit-button'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'

const commonStatusCodes = [
  200, 201, 202, 204, 301, 302, 400, 401, 403, 404, 409, 422, 500, 502, 503,
]
const defaultHeaderRows = [{ key: 'Content-Type', value: 'application/json' }]

interface HeaderRow {
  key: string
  value: string
}

function headersToRows(headers?: Record<string, string>): HeaderRow[] {
  if (!headers || Object.keys(headers).length === 0) {
    return defaultHeaderRows
  }
  return Object.entries(headers).map(([key, value]) => ({ key, value }))
}

function serializeHeaders(rows: HeaderRow[]) {
  return rows
    .filter((row) => row.key.trim() && row.value.trim())
    .map((row) => `${row.key.trim()}: ${row.value.trim()}`)
    .join('\n')
}

function FormSection({
  icon: Icon,
  title,
  description,
  children,
}: {
  icon: LucideIcon
  title: string
  description: string
  children: ReactNode
}) {
  return (
    <section className="rounded-xl border border-border bg-surface-low p-4 md:p-5">
      <div className="mb-4 space-y-1">
        <h3 className="flex items-center gap-2 text-base font-semibold text-foreground">
          <Icon aria-hidden="true" className="h-4 w-4 text-primary" />
          {title}
        </h3>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
      {children}
    </section>
  )
}

export function MonitorForm({
  monitor,
  serviceId,
  error,
}: {
  monitor?: Monitor
  serviceId: string
  error?: string
}) {
  const action = monitor ? updateMonitorAction : createMonitorAction
  const [headers, setHeaders] = useState<HeaderRow[]>(headersToRows(monitor?.http?.headers))
  const [statusCodes, setStatusCodes] = useState<number[]>(
    monitor?.http?.expectedStatusCodes?.length ? monitor.http.expectedStatusCodes : [200]
  )

  function updateHeader(index: number, next: HeaderRow) {
    setHeaders((current) => current.map((row, rowIndex) => (rowIndex === index ? next : row)))
  }

  function removeHeader(index: number) {
    setHeaders((current) => current.filter((_, rowIndex) => rowIndex !== index))
  }

  function addStatusCode(rawValue: string) {
    if (!rawValue) {
      return
    }
    const nextCode = Number(rawValue)
    setStatusCodes((current) =>
      current.includes(nextCode)
        ? current
        : [...current, nextCode].sort((left, right) => left - right)
    )
  }

  return (
    <div className="grid gap-5">
      <div>
        <h2 className="text-xl font-semibold tracking-tight text-foreground">
          {monitor ? 'Edit monitor' : 'Create monitor'}
        </h2>
      </div>
      <form action={action} className="grid gap-5">
        <input name="serviceId" type="hidden" value={serviceId} />
        {monitor ? <input name="monitorId" type="hidden" value={monitor.monitorId} /> : null}
        <input
          name="returnTo"
          type="hidden"
          value={
            monitor
              ? `/services/${serviceId}/monitors/${monitor.monitorId}`
              : `/services/${serviceId}/monitors/new`
          }
        />
        <input name="headers" type="hidden" value={serializeHeaders(headers)} />
        <input name="expectedStatusCodes" type="hidden" value={statusCodes.join(', ')} />
        <input
          name="expectedBodyContains"
          type="hidden"
          value={monitor?.http?.expectedBodyContains ?? ''}
        />
        <FormSection
          description="Name the monitor and choose how often it should run."
          icon={Fingerprint}
          title="Identity"
        >
          <div className="grid gap-5 lg:grid-cols-2">
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Monitor name</span>
              <Input
                defaultValue={monitor?.name ?? ''}
                name="name"
                placeholder="Homepage availability"
                required
              />
            </label>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Check frequency</span>
              <Select defaultValue={String(monitor?.intervalSeconds ?? 60)} name="intervalSeconds">
                {monitorCadenceOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </Select>
            </label>
          </div>
        </FormSection>
        <div aria-label="Monitor protocol" className="grid gap-2" role="tablist">
          <p className="text-sm font-semibold text-foreground">Protocol</p>
          <div className="grid gap-2 sm:grid-cols-3">
            {[
              { label: 'HTTP', disabled: false, hint: 'Selected' },
              { label: 'TCP', disabled: true, hint: 'Coming soon' },
              { label: 'gRPC', disabled: true, hint: 'Coming soon' },
            ].map((protocol) => (
              <button
                aria-disabled={protocol.disabled}
                aria-selected={!protocol.disabled}
                className={cn(
                  'rounded-lg border px-4 py-3 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
                  protocol.disabled
                    ? 'cursor-not-allowed border-border bg-surface-low text-muted-foreground opacity-70'
                    : 'border-primary bg-primary/10 text-foreground'
                )}
                disabled={protocol.disabled}
                key={protocol.label}
                role="tab"
                title={protocol.disabled ? `${protocol.label} coming soon` : 'HTTP selected'}
                type="button"
              >
                <span className="block text-sm font-semibold">{protocol.label}</span>
                <span className="mt-1 block text-xs text-muted-foreground">{protocol.hint}</span>
              </button>
            ))}
          </div>
        </div>
        <FormSection
          description="Configure the HTTP request sent by each monitor check."
          icon={Send}
          title="Request"
        >
          <div className="grid gap-5 lg:grid-cols-[0.7fr_1.6fr_0.7fr]">
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Method</span>
              <Select defaultValue={monitor?.http?.method ?? 'GET'} name="method">
                <option value="GET">GET</option>
                <option value="HEAD">HEAD</option>
                <option value="POST">POST</option>
              </Select>
            </label>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">HTTP target URL</span>
              <Input
                defaultValue={monitor?.http?.target ?? 'https://example.com'}
                name="target"
                required
                type="url"
              />
            </label>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Timeout ms</span>
              <Input
                defaultValue={String(monitor?.http?.timeoutMs ?? 5000)}
                min={100}
                name="timeoutMs"
                required
                type="number"
              />
            </label>
          </div>
          <div className="mt-5 grid gap-3">
            <div>
              <p className="text-sm font-semibold text-foreground">Headers</p>
              <p className="text-sm text-muted-foreground">Add HTTP headers as key/value pairs.</p>
            </div>
            {headers.map((header, index) => (
              <div className="grid gap-2 md:grid-cols-[1fr_1fr_auto]" key={index}>
                <Input
                  aria-label={`Header ${index + 1} name`}
                  onChange={(event) => updateHeader(index, { ...header, key: event.target.value })}
                  placeholder="Header name"
                  value={header.key}
                />
                <Input
                  aria-label={`Header ${index + 1} value`}
                  onChange={(event) =>
                    updateHeader(index, { ...header, value: event.target.value })
                  }
                  placeholder="Header value"
                  value={header.value}
                />
                <Button
                  aria-label={`Remove header ${index + 1}`}
                  onClick={() => removeHeader(index)}
                  size="default"
                  type="button"
                  variant="outline"
                >
                  <Trash2 aria-hidden="true" className="h-4 w-4" />
                </Button>
              </div>
            ))}
            <Button
              className="w-full gap-2"
              onClick={() => setHeaders((current) => [...current, { key: '', value: '' }])}
              type="button"
              variant="outline"
            >
              <Plus aria-hidden="true" className="h-4 w-4" /> Add header
            </Button>
          </div>
        </FormSection>
        <FormSection
          description="Define which HTTP responses count as successful checks."
          icon={CheckCircle2}
          title="Validation"
        >
          <div className="grid gap-3">
            <div className="flex flex-wrap gap-2">
              {statusCodes.map((statusCode) => (
                <span
                  className="inline-flex items-center gap-1 rounded-full bg-surface-high px-3 py-1 text-sm font-semibold text-foreground"
                  key={statusCode}
                >
                  {statusCode}
                  <button
                    aria-label={`Remove ${statusCode}`}
                    className="rounded-full text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    onClick={() =>
                      setStatusCodes((current) => current.filter((code) => code !== statusCode))
                    }
                    type="button"
                  >
                    <X aria-hidden="true" className="h-3.5 w-3.5" />
                  </button>
                </span>
              ))}
            </div>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Add expected status code</span>
              <Select
                aria-label="Add expected status code"
                onChange={(event) => {
                  addStatusCode(event.target.value)
                  event.target.value = ''
                }}
                value=""
              >
                <option value="">Choose common status code</option>
                {commonStatusCodes
                  .filter((statusCode) => !statusCodes.includes(statusCode))
                  .map((statusCode) => (
                    <option key={statusCode} value={statusCode}>
                      {statusCode}
                    </option>
                  ))}
              </Select>
            </label>
            <p className="text-sm text-muted-foreground">
              Body content validation is coming soon for HTTP monitors.
            </p>
          </div>
        </FormSection>
        {!monitor ? (
          <label className="flex items-center gap-3 rounded-md border border-border bg-surface-low px-4 py-3 text-sm text-foreground">
            <input
              className="h-4 w-4 accent-cyan-400"
              defaultChecked
              name="enabled"
              type="checkbox"
            />
            Start enabled
          </label>
        ) : null}
        {error ? (
          <p className="rounded-md border border-status-down/30 bg-status-down/10 px-3 py-2 text-sm text-status-down">
            {error}
          </p>
        ) : null}
        <div className="flex items-center justify-end gap-3">
          <SubmitButton>{monitor ? 'Save changes' : 'Create monitor'}</SubmitButton>
        </div>
      </form>
    </div>
  )
}
