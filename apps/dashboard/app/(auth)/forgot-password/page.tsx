import { Activity } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

import { beginPasswordRecoveryAction } from './actions'

export default function ForgotPasswordPage() {
  return (
    <main className="flex min-h-screen items-center justify-center px-5 py-10" id="main-content">
      <section className="w-full max-w-md rounded-xl border border-border bg-surface p-6 shadow-2xl shadow-surface-lowest/40 sm:p-8">
        <div className="flex items-center gap-3">
          <span className="grid size-10 place-items-center rounded-lg bg-primary/15 text-primary">
            <Activity aria-hidden="true" className="size-5" />
          </span>
          <div>
            <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-primary">
              Bolt Monitor
            </p>
            <h1 className="mt-1 text-2xl font-semibold tracking-tight text-foreground">
              Reset your password
            </h1>
          </div>
        </div>
        <p className="mt-6 text-sm leading-6 text-muted-foreground">
          Enter your email address and we will send recovery instructions if available.
        </p>
        <form
          action={beginPasswordRecoveryAction}
          className="mt-6 grid gap-5 border-t border-border pt-6"
        >
          <div className="grid gap-2">
            <label className="text-sm font-medium text-foreground" htmlFor="email">
              Email
            </label>
            <Input autoComplete="email" id="email" name="email" required type="email" />
          </div>
          <Button type="submit">Send recovery instructions</Button>
        </form>
      </section>
    </main>
  )
}
