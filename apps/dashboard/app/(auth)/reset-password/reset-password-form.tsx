'use client'

import { useActionState } from 'react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { messageForAuthFeedback } from '@/lib/auth/feedback'

import { resetPasswordAction, type ResetPasswordFormState } from './actions'

const initialState: ResetPasswordFormState = { feedback: null }

export function ResetPasswordForm() {
  const [state, formAction, pending] = useActionState(resetPasswordAction, initialState)

  return (
    <form action={formAction} className="grid gap-5">
      <div className="grid gap-2">
        <label className="text-sm font-medium text-foreground" htmlFor="code">
          Recovery code
        </label>
        <Input autoComplete="one-time-code" id="code" name="code" required type="text" />
      </div>
      <div className="grid gap-2">
        <label className="text-sm font-medium text-foreground" htmlFor="newPassword">
          New password
        </label>
        <Input
          autoComplete="new-password"
          id="newPassword"
          name="newPassword"
          required
          type="password"
        />
      </div>
      {state.feedback ? (
        <p
          aria-live="polite"
          className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive-foreground"
        >
          {messageForAuthFeedback(state.feedback)}
        </p>
      ) : null}
      <Button disabled={pending} type="submit">
        {pending ? 'Resetting password...' : 'Reset password'}
      </Button>
    </form>
  )
}
