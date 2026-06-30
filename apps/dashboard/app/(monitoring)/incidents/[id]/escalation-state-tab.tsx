import { ApiError, getIncidentEscalationState } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { EmptyState } from '@/components/empty-state'
import { formatDateTime } from '@/lib/utils'

export async function EscalationStateTab({ incidentId }: { incidentId: string }) {
  let state
  try {
    state = await getIncidentEscalationState(incidentId)
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      return (
        <Card>
          <CardContent className="pt-6">
            <EmptyState description="Incident not found." title="Escalation unavailable" />
          </CardContent>
        </Card>
      )
    }
    const message = error instanceof Error ? error.message : 'Unable to load escalation state.'
    return (
      <Card>
        <CardContent className="pt-6">
          <EmptyState description={message} title="Escalation unavailable" />
        </CardContent>
      </Card>
    )
  }

  if (!state.exists) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Escalation state</CardTitle>
        </CardHeader>
        <CardContent>
          <EmptyState
            description="No escalation has been initiated for this incident."
            title="No escalation active"
          />
        </CardContent>
      </Card>
    )
  }

  const statusLabel = (state.status ?? '').toLowerCase()
  const scheduledLabel = state.scheduledFor ? formatDateTime(state.scheduledFor) : null

  return (
    <Card>
      <CardHeader>
        <CardTitle>Escalation state</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-4 md:grid-cols-2">
          <div className="rounded-lg border border-border bg-surface-low p-4">
            <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
              Status
            </p>
            <p className="mt-2 text-xl font-semibold text-foreground capitalize">
              {statusLabel || 'unknown'}
            </p>
          </div>
          <div className="rounded-lg border border-border bg-surface-low p-4">
            <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
              Current step
            </p>
            <p className="mt-2 text-xl font-semibold text-foreground">
              {state.currentStep !== undefined ? `Step ${state.currentStep + 1}` : '—'}
            </p>
          </div>
          <div className="rounded-lg border border-border bg-surface-low p-4">
            <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
              Selected path
            </p>
            <p className="mt-2 text-sm font-medium text-foreground">{state.selectedPath ?? '—'}</p>
          </div>
          <div className="rounded-lg border border-border bg-surface-low p-4">
            <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
              Next step ETA
            </p>
            <p className="mt-2 text-sm font-medium text-foreground">{scheduledLabel ?? '—'}</p>
          </div>
        </div>

        <div>
          <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
            Steps fired
          </p>
          {state.stepsFired && state.stepsFired.length > 0 ? (
            <ol className="mt-3 space-y-2">
              {state.stepsFired.map((step) => (
                <li
                  key={step}
                  className="flex items-center gap-3 rounded-md border border-border bg-surface-low px-3 py-2 text-sm"
                >
                  <span className="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                    Step {step + 1}
                  </span>
                  <span className="text-muted-foreground">Fired</span>
                </li>
              ))}
            </ol>
          ) : (
            <p className="mt-3 text-sm text-muted-foreground">No steps have fired yet.</p>
          )}
        </div>

        {state.policyId ? (
          <div className="rounded-lg border border-border bg-surface-low p-4 text-sm">
            <span className="font-medium">Policy:</span>{' '}
            <span className="font-mono">{state.policyId}</span>
          </div>
        ) : null}
      </CardContent>
    </Card>
  )
}
