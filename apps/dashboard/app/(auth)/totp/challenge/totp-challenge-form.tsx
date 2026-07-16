'use client'

import { useActionState } from 'react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

import { completeTotpChallengeAction, type TotpChallengeFormState } from './actions'

const initialState: TotpChallengeFormState = { message: null }

export function TotpChallengeForm() {
  const [state, formAction, pending] = useActionState(completeTotpChallengeAction, initialState)
  return (
    <form action={formAction} className="grid gap-5">
      <div className="grid gap-2">
        <label className="text-sm font-medium text-foreground" htmlFor="code">
          Authentication code
        </label>
        <Input autoComplete="one-time-code" id="code" inputMode="numeric" name="code" required />
      </div>
      {state.message ? (
        <p aria-live="polite" className="text-sm text-destructive">
          {state.message}
        </p>
      ) : null}
      <Button disabled={pending} type="submit">
        {pending ? 'Verifying...' : 'Continue'}
      </Button>
    </form>
  )
}
