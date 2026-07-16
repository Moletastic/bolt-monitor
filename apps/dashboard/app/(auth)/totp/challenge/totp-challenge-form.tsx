'use client'

import { useActionState } from 'react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { messageForAuthFeedback } from '@/lib/auth/feedback'

import { completeTotpChallengeAction, type TotpChallengeFormState } from './actions'

const initialState: TotpChallengeFormState = { feedback: null }

export function TotpChallengeForm({ returnTarget }: { returnTarget: string }) {
  const [state, formAction, pending] = useActionState(completeTotpChallengeAction, initialState)
  return (
    <form action={formAction} className="grid gap-5">
      <input name="returnTo" type="hidden" value={returnTarget} />
      <div className="grid gap-2">
        <label className="text-sm font-medium text-foreground" htmlFor="code">
          Authentication code
        </label>
        <Input autoComplete="one-time-code" id="code" inputMode="numeric" name="code" required />
      </div>
      {state.feedback ? (
        <p aria-live="polite" className="text-sm text-destructive">
          {messageForAuthFeedback(state.feedback)}
        </p>
      ) : null}
      <Button disabled={pending} type="submit">
        {pending ? 'Verifying...' : 'Continue'}
      </Button>
    </form>
  )
}
