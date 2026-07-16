import { Activity } from 'lucide-react'

import { ActivationForm } from './activation-form'
import { sanitizeReturnTarget } from '@/lib/auth/return-target'

export default async function ActivateInvitationPage({
  searchParams,
}: {
  searchParams: Promise<{ returnTo?: string | string[] }>
}) {
  const { returnTo } = await searchParams

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
              Activate your account
            </h1>
          </div>
        </div>
        <p className="mt-6 text-sm leading-6 text-muted-foreground">
          Choose a new password to activate your operator account.
        </p>
        <div className="mt-6 border-t border-border pt-6">
          <ActivationForm returnTarget={sanitizeReturnTarget(returnTo)} />
        </div>
      </section>
    </main>
  )
}
