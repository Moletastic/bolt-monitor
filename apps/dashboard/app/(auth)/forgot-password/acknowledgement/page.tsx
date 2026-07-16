import Link from 'next/link'

import { buttonVariants } from '@/components/ui/button'

export default function PasswordRecoveryAcknowledgementPage() {
  return (
    <main className="flex min-h-screen items-center justify-center px-5 py-10" id="main-content">
      <section className="w-full max-w-md rounded-xl border border-border bg-surface p-6 shadow-2xl shadow-surface-lowest/40 sm:p-8">
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">Check your email</h1>
        <p className="mt-4 text-sm leading-6 text-muted-foreground">
          If recovery is available for that address, you will receive instructions shortly.
        </p>
        <div className="mt-6">
          <Link className={buttonVariants()} href="/reset-password">
            Enter recovery code
          </Link>
        </div>
      </section>
    </main>
  )
}
