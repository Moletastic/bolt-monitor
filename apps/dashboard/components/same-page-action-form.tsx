'use client'

import { useActionState, useEffect, useState, type ReactNode } from 'react'
import { Power, PowerOff } from 'lucide-react'

import { samePageActionStartEvent } from '@/components/query-feedback-banner'
import { Button } from '@/components/ui/button'
import type { ButtonProps } from '@/components/ui/button'
import { idleActionState, actionErrorMessage, type ActionState } from '@/lib/action-state'
import { cn } from '@/lib/utils'

const icons = {
  power: Power,
  powerOff: PowerOff,
}

type SamePageActionFormProps = {
  action: (_previousState: ActionState, formData: FormData) => Promise<ActionState>
  buttonLabel: string
  children: ReactNode
  buttonClassName?: string
  className?: string
  compactLabel?: string
  disabled?: boolean
  iconName?: keyof typeof icons
  iconOnlyBelow?: 'lg'
  pendingLabel?: string
  size?: ButtonProps['size']
  variant?: ButtonProps['variant']
}

export function SamePageActionForm({
  action,
  buttonClassName,
  buttonLabel,
  children,
  className,
  compactLabel,
  disabled,
  iconName,
  iconOnlyBelow,
  pendingLabel = 'Working...',
  size,
  variant,
}: SamePageActionFormProps) {
  const [state, formAction, pending] = useActionState(action, idleActionState)
  const [locallyPending, setLocallyPending] = useState(false)
  const isPending = pending || locallyPending
  const Icon = iconName ? icons[iconName] : undefined

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
      <Button
        aria-label={buttonLabel}
        className={buttonClassName}
        disabled={isPending || disabled}
        size={size}
        type="submit"
        variant={variant}
      >
        {Icon && <Icon aria-hidden="true" className="h-4 w-4" />}
        {isPending ? (
          pendingLabel
        ) : iconOnlyBelow === 'lg' ? (
          <span className="hidden lg:inline">{compactLabel ?? buttonLabel}</span>
        ) : (
          buttonLabel
        )}
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
