import { Activity } from 'lucide-react'

import { TotpChallengeForm } from './totp-challenge-form'

export default function TotpChallengePage() {
  return (
    <main className="flex min-h-screen items-center justify-center px-5 py-10" id="main-content">
      <section className="w-full max-w-md rounded-xl border border-border bg-surface p-6 shadow-2xl shadow-surface-lowest/40 sm:p-8">
        <Activity aria-hidden="true" className="size-5 text-primary" />
        <h1 className="mt-4 text-2xl font-semibold tracking-tight text-foreground">
          Verify your code
        </h1>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">
          Enter the code from your authenticator app to continue.
        </p>
        <div className="mt-6 border-t border-border pt-6">
          <TotpChallengeForm />
        </div>
      </section>
    </main>
  )
}
