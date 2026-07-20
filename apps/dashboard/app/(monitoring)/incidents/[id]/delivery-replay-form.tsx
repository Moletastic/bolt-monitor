'use client'

import { useActionState } from 'react'

import { replayDeliveryStateAction } from '@/lib/actions'
import { idleActionState, actionErrorMessage, type ActionState } from '@/lib/action-state'
import { SamePageActionForm } from '@/components/same-page-action-form'
import { useCallback, useEffect, useState } from 'react'

export function DeliveryReplayForm({
  incidentId,
  deliveryId,
  state,
}: {
  incidentId: string
  deliveryId: string
  state: string
}) {
  const [actionState, dispatch, pending] = useActionState(
    replayDeliveryStateAction,
    idleActionState
  )
  const [showSuccess, setShowSuccess] = useState(false)
  useEffect(() => {
    if (actionState.status === 'success') setShowSuccess(true)
    if (actionState.status === 'error') setShowSuccess(false)
  }, [actionState.status])

  const action = useCallback(
    async (_previous: ActionState, formData: FormData) => {
      dispatch(formData)
      return idleActionState
    },
    [dispatch]
  )

  return (
    <div className="flex flex-col gap-2" data-testid={`replay-form-${deliveryId}`}>
      <SamePageActionForm
        action={action}
        buttonLabel="Replay delivery"
        pendingLabel={pending ? 'Replaying…' : 'Replay delivery'}
        size="sm"
        variant="secondary"
      >
        <input type="hidden" name="incidentId" value={incidentId} />
        <input type="hidden" name="deliveryId" value={deliveryId} />
        <input type="hidden" name="state" value={state} />
      </SamePageActionForm>
      {actionState.status === 'error' ? (
        <span className="text-xs text-red-700" role="alert">
          {actionErrorMessage(actionState.error)}
        </span>
      ) : null}
      {showSuccess && actionState.status === 'success' ? (
        <span className="text-xs text-emerald-700" role="status">
          {actionState.message ?? 'Replay queued.'}
        </span>
      ) : null}
    </div>
  )
}
