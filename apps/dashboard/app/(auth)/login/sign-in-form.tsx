'use client'

import { useActionState } from 'react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

import { signInAction, type SignInFormState } from './actions'

const initialState: SignInFormState = { message: null }

export function SignInForm() {
  const [state, formAction, pending] = useActionState(signInAction, initialState)

  return (
    <form action={formAction} className="grid gap-5">
      <div className="grid gap-2">
        <label className="text-sm font-medium text-foreground" htmlFor="email">
          Email
        </label>
        <Input autoComplete="email" id="email" name="email" required type="email" />
      </div>
      <div className="grid gap-2">
        <label className="text-sm font-medium text-foreground" htmlFor="password">
          Password
        </label>
        <Input
          autoComplete="current-password"
          id="password"
          name="password"
          required
          type="password"
        />
      </div>
      {state.message ? (
        <p
          aria-live="polite"
          className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive-foreground"
        >
          {state.message}
        </p>
      ) : null}
      <Button disabled={pending} type="submit">
        {pending ? 'Signing in...' : 'Sign in'}
      </Button>
    </form>
  )
}
