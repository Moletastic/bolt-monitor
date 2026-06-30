import { createMonitorAction, updateMonitorAction } from '@/lib/actions'
import type { Monitor, ProbeLocation } from '@/lib/types'
import { monitorCadenceOptions } from '@/lib/utils'
import { getMonitorLocationField } from '@/lib/probe-locations'

import { SubmitButton } from '@/components/submit-button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'

const defaultHeaders = 'Accept: text/html'

function stringHeaders(headers?: Record<string, string>) {
  if (!headers || Object.keys(headers).length === 0) {
    return defaultHeaders
  }
  return Object.entries(headers)
    .map(([key, value]) => `${key}: ${value}`)
    .join('\n')
}

export function MonitorForm({
  monitor,
  serviceId,
  locations,
  error,
}: {
  monitor?: Monitor
  serviceId: string
  locations: ProbeLocation[]
  error?: string
}) {
  const action = monitor ? updateMonitorAction : createMonitorAction
  const locationField = getMonitorLocationField(locations)
  const currentLocation = monitor?.probeLocations?.[0]

  return (
    <Card>
      <CardHeader>
        <CardTitle>{monitor ? 'Edit monitor' : 'Create monitor'}</CardTitle>
      </CardHeader>
      <CardContent>
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
          <div className="grid gap-5 lg:grid-cols-2">
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Name</span>
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
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">HTTP target</span>
              <Input
                defaultValue={monitor?.http?.target ?? 'https://example.com'}
                name="target"
                required
                type="url"
              />
            </label>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Method</span>
              <Select defaultValue={monitor?.http?.method ?? 'GET'} name="method">
                <option value="GET">GET</option>
                <option value="HEAD">HEAD</option>
                <option value="POST">POST</option>
              </Select>
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
            <div className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Probe location</span>
              {locationField.kind === 'single-fixed' ? (
                <>
                  <input
                    name="probeLocation"
                    type="hidden"
                    value={locationField.location.locationId}
                  />
                  <div
                    aria-label="Probe location"
                    className="flex items-center gap-2 rounded-md border border-border bg-surface-low px-3 py-2 text-sm text-foreground"
                  >
                    <span aria-hidden="true" className="font-semibold">
                      {locationField.location.locationId.toUpperCase()}
                    </span>
                    <span aria-hidden="true" className="text-muted-foreground">
                      ·
                    </span>
                    <span>{locationField.location.displayName}</span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Single-region preview. Once multi-region probes are enabled, operators will pick
                    regions here.{' '}
                    <a className="text-primary hover:underline" href="/locations">
                      View probe locations
                    </a>
                    .
                  </p>
                </>
              ) : (
                <Select
                  defaultValue={currentLocation ?? locationField.locations[0]?.locationId ?? ''}
                  name="probeLocation"
                >
                  {locationField.locations.map((loc) => (
                    <option key={loc.locationId} value={loc.locationId}>
                      {loc.locationId.toUpperCase()} · {loc.displayName}
                    </option>
                  ))}
                </Select>
              )}
            </div>
            <label className="grid gap-2 text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">Expected status codes</span>
              <Input
                defaultValue={monitor?.http?.expectedStatusCodes?.join(', ') ?? '200'}
                name="expectedStatusCodes"
                placeholder="200, 204"
              />
            </label>
            <label className="grid gap-2 text-sm text-muted-foreground lg:col-span-2">
              <span className="font-semibold text-foreground">Expected body contains</span>
              <Input
                defaultValue={monitor?.http?.expectedBodyContains ?? ''}
                name="expectedBodyContains"
                placeholder="Optional response text"
              />
            </label>
            <label className="grid gap-2 text-sm text-muted-foreground lg:col-span-2">
              <span className="font-semibold text-foreground">Headers</span>
              <textarea
                className="min-h-32 rounded-md border border-border bg-surface-low px-3 py-2 font-mono text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                defaultValue={stringHeaders(monitor?.http?.headers)}
                name="headers"
                placeholder="Header-Name: value"
              />
            </label>
          </div>
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
      </CardContent>
    </Card>
  )
}
