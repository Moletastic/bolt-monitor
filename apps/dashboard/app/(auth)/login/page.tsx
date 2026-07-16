import { Activity } from 'lucide-react'
import Link from 'next/link'

import { SignInForm } from './sign-in-form'

export default function SignInPage() {
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
              Operator sign in
            </h1>
          </div>
        </div>
        <p className="mt-6 text-sm leading-6 text-muted-foreground">
          Sign in with the credentials issued to your operator account.
        </p>
        <div className="mt-6 border-t border-border pt-6">
          <SignInForm />
          <Link
            className="mt-4 inline-block text-sm text-primary hover:underline"
            href="/forgot-password"
          >
            Forgot your password?
          </Link>
        </div>
      </section>
    </main>
  )
}
