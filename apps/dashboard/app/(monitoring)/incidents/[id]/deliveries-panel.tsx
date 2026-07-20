import { ApiError, listIncidentDeliveries } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { EmptyState } from '@/components/empty-state'
import { formatDateTime } from '@/lib/utils'
import { type Delivery, type DeliveryState, isReplayable } from '@/lib/types'

import { __DELIVERY_STATE_LABELS } from '@/lib/deliveries-panel-helpers'

import { DeliveryReplayForm } from './delivery-replay-form'

const STATE_LABEL: Record<DeliveryState, string> = __DELIVERY_STATE_LABELS

function stateDescription(state: DeliveryState): string {
  switch (state) {
    case 'pending':
      return 'Durable work exists and no provider attempt is active.'
    case 'in_flight':
      return 'One fenced provider attempt is active until lease expiry.'
    case 'retryable_failed':
      return 'Confirmed transient failure; eligible for automatic retry within budget.'
    case 'ambiguous':
      return 'Provider may have accepted the request; duplicate-side-effect risk disclosed.'
    case 'delivered':
      return 'Provider accepted the notification. The system does not guarantee human receipt.'
    case 'terminal_failed':
      return 'Terminal rejection, configuration failure, or exhausted retry budget. Eligible for replay.'
  }
}

function stateClass(state: DeliveryState): string {
  switch (state) {
    case 'delivered':
      return 'border-emerald-300 bg-emerald-50 text-emerald-900'
    case 'in_flight':
    case 'pending':
      return 'border-blue-300 bg-blue-50 text-blue-900'
    case 'ambiguous':
    case 'retryable_failed':
      return 'border-amber-300 bg-amber-50 text-amber-900'
    case 'terminal_failed':
      return 'border-red-300 bg-red-50 text-red-900'
  }
}

export async function IncidentDeliveriesPanel({ incidentId }: { incidentId: string }) {
  let deliveries: Delivery[] = []
  let loadError: string | null = null

  try {
    deliveries = await listIncidentDeliveries(incidentId)
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      return (
        <Card>
          <CardContent className="pt-6">
            <EmptyState description="Incident not found." title="Deliveries unavailable" />
          </CardContent>
        </Card>
      )
    }
    loadError = error instanceof Error ? error.message : 'Unable to load deliveries.'
  }

  if (loadError) {
    return (
      <Card>
        <CardContent className="pt-6">
          <EmptyState description={loadError} title="Deliveries unavailable" />
        </CardContent>
      </Card>
    )
  }

  if (deliveries.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Notification deliveries</CardTitle>
        </CardHeader>
        <CardContent>
          <EmptyState
            description="No notification deliveries have been recorded for this incident yet."
            title="No deliveries yet"
          />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Notification deliveries</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-sm text-muted-foreground">
          Each card shows the durable state of one provider attempt for a single step and channel.
          Delivery means the provider accepted the request; the system does not claim that a human
          read or acted on it.
        </p>
        <ul className="space-y-3">
          {deliveries.map((delivery) => (
            <li
              key={delivery.deliveryId}
              data-testid={`delivery-${delivery.deliveryId}`}
              className={`rounded-lg border p-4 ${stateClass(delivery.state)}`}
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p className="text-xs font-bold uppercase tracking-[0.18em] opacity-80">
                    Step {delivery.stepNumber} · {delivery.channelType}
                  </p>
                  <p className="mt-1 font-mono text-sm text-foreground">{delivery.channelId}</p>
                </div>
                <DeliveryStateBadge state={delivery.state} />
              </div>
              <p className="mt-2 text-sm">{stateDescription(delivery.state)}</p>
              <dl className="mt-3 grid gap-2 text-xs text-muted-foreground sm:grid-cols-2">
                <div>
                  <dt className="font-bold uppercase tracking-[0.18em]">Attempts</dt>
                  <dd>{delivery.attemptCount}</dd>
                </div>
                <div>
                  <dt className="font-bold uppercase tracking-[0.18em]">Last attempt</dt>
                  <dd>{delivery.lastAttemptAt ? formatDateTime(delivery.lastAttemptAt) : '—'}</dd>
                </div>
                {delivery.lastOutcomeClass ? (
                  <div>
                    <dt className="font-bold uppercase tracking-[0.18em]">Outcome class</dt>
                    <dd>{delivery.lastOutcomeClass}</dd>
                  </div>
                ) : null}
                {delivery.providerRequestId ? (
                  <div>
                    <dt className="font-bold uppercase tracking-[0.18em]">Provider request id</dt>
                    <dd className="font-mono">{delivery.providerRequestId}</dd>
                  </div>
                ) : null}
              </dl>
              {isReplayable(delivery) ? (
                <div className="mt-3">
                  <DeliveryReplayForm
                    incidentId={incidentId}
                    deliveryId={delivery.deliveryId}
                    state={delivery.state}
                  />
                </div>
              ) : null}
            </li>
          ))}
        </ul>
        <p className="text-xs text-muted-foreground">
          Recovery suppression is reported as escalation eligibility, not as a delivery state.
        </p>
      </CardContent>
    </Card>
  )
}

function DeliveryStateBadge({ state }: { state: DeliveryState }) {
  return (
    <span
      className={`rounded-full border px-2 py-0.5 text-xs font-bold uppercase tracking-[0.18em] ${stateClass(state)}`}
      aria-label={`delivery state ${state}`}
    >
      {STATE_LABEL[state]}
    </span>
  )
}

export const __DELIVERY_STATE_LABELS = STATE_LABEL
