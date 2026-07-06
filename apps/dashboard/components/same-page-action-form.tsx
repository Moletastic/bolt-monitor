'use client'

import { useActionState, useEffect, useState, type ReactNode } from 'react'

import { samePageActionStartEvent } from '@/components/query-feedback-banner'
import { Button } from '@/components/ui/button'
import type { ButtonProps } from '@/components/ui/button'
import { idleActionState, actionErrorMessage, type ActionState } from '@/lib/action-state'
import { cn } from '@/lib/utils'

type SamePageActionFormProps = {
  action: (_previousState: ActionState, formData: FormData) => Promise<ActionState>
  buttonLabel: string
  children: ReactNode
  className?: string
  disabled?: boolean
  pendingLabel?: string
  size?: ButtonProps['size']
  variant?: ButtonProps['variant']
}

export function SamePageActionForm({
  action,
  buttonLabel,
  children,
  className,
  disabled,
  pendingLabel = 'Working...',
  size,
  variant,
}: SamePageActionFormProps) {
  const [state, formAction, pending] = useActionState(action, idleActionState)
  const [locallyPending, setLocallyPending] = useState(false)
  const isPending = pending || locallyPending

  useEffect(() => {
    if (state.status !== 'idle') setLocallyPending(false)
  }, [state.status])

  return (
    <form
      action={formAction}
      className={cn('space-y-2', className)}
      onSubmit={() => {
        setLocallyPending(true)
        window.dispatchEvent(new Event(samePageActionStartEvent))
      }}
    >
      {children}
      <Button disabled={isPending || disabled} size={size} type="submit" variant={variant}>
        {isPending ? pendingLabel : buttonLabel}
      </Button>
      {state.status === 'success' && state.message ? (
        <p
          aria-live="polite"
          className="rounded-md border border-status-up/30 bg-status-up/10 px-3 py-2 text-sm text-status-up"
          role="status"
        >
          {state.message}
        </p>
      ) : null}
      {state.status === 'error' ? (
        <p
          className="rounded-md border border-status-down/30 bg-status-down/10 px-3 py-2 text-sm text-status-down"
          role="alert"
        >
          {actionErrorMessage(state.error)}
        </p>
      ) : null}
    </form>
  )
}
