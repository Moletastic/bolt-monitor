import 'server-only'

import type {
  AuthTransactionReference,
  AuthTransactionStore,
  DashboardSessionStore,
  IdentityProvider,
  NewDashboardSession,
} from '@/lib/auth/contracts'
import type { AuthFailure } from '@/lib/auth/feedback'
import { establishDashboardSession } from '@/lib/io/auth/authentication-state'

export type PasswordSignInResult =
  | { readonly kind: 'authenticated'; readonly sessionReference: string }
  | {
      readonly kind: 'challenge-required'
      readonly challenge: 'new-password-required' | 'software-token-mfa' | 'software-token-setup'
      readonly transactionReference: AuthTransactionReference
    }
  | { readonly kind: 'failed'; readonly failure: AuthFailure }

export type NewPasswordResult =
  | { readonly kind: 'authenticated'; readonly sessionReference: string }
  | { readonly kind: 'failed'; readonly failure: AuthFailure }

/**
 * Keeps password submission at the server boundary and turns only completed
 * Cognito authentication into an opaque dashboard session.
 */
export async function signInWithPassword(input: {
  readonly username: string
  readonly password: string
  readonly sessionExpiresAt: number
  readonly provider: Pick<IdentityProvider, 'beginSignIn'>
  readonly sessionStore: Pick<DashboardSessionStore, 'create'>
  readonly transactionStore: Pick<AuthTransactionStore, 'create'>
  readonly transactionExpiresAt: number
}): Promise<PasswordSignInResult> {
  if (!input.username || !input.password) return { kind: 'failed', failure: 'validation-failed' }

  const signIn = await input.provider.beginSignIn({
    username: input.username,
    password: input.password,
  })
  if (!signIn.ok) return { kind: 'failed', failure: signIn.error }
  if (signIn.value.kind === 'challenge') {
    if (signIn.value.challenge.kind === 'password-recovery')
      return { kind: 'failed', failure: { kind: 'challenge-failed' } }
    const transaction = await input.transactionStore.create({
      flow: 'sign-in',
      challenge: signIn.value.challenge.kind,
      providerState: signIn.value.challenge.continuation,
      attempts: 0,
      expiresAt: input.transactionExpiresAt,
    })
    return transaction.ok
      ? {
          kind: 'challenge-required',
          challenge: signIn.value.challenge.kind,
          transactionReference: transaction.value,
        }
      : { kind: 'failed', failure: transaction.error }
  }

  const session: NewDashboardSession = {
    subject: signIn.value.subject,
    tokens: signIn.value.tokens,
    expiresAt: input.sessionExpiresAt,
  }
  const created = await input.sessionStore.create(session)
  return created.ok
    ? { kind: 'authenticated', sessionReference: created.value }
    : { kind: 'failed', failure: created.error }
}

/** Completes an invitation's server-held Cognito challenge into a new session. */
export async function completeNewPasswordChallenge(input: {
  readonly reference: AuthTransactionReference
  readonly newPassword: string
  readonly sessionExpiresAt: number
  readonly provider: Pick<IdentityProvider, 'answerNewPassword'>
  readonly transactionStore: Pick<AuthTransactionStore, 'read' | 'consume' | 'invalidate'>
  readonly sessionStore: Pick<DashboardSessionStore, 'create' | 'invalidate'>
  readonly priorSession?: import('@/lib/auth/contracts').DashboardSessionReference
}): Promise<NewPasswordResult> {
  if (!input.newPassword) return { kind: 'failed', failure: 'validation-failed' }

  const transaction = await input.transactionStore.read(input.reference, 'sign-in')
  if (
    !transaction.ok ||
    !transaction.value ||
    transaction.value.challenge !== 'new-password-required'
  )
    return {
      kind: 'failed',
      failure: transaction.ok ? { kind: 'transaction-invalid' } : transaction.error,
    }

  const completed = await input.provider.answerNewPassword({
    continuation: transaction.value.providerState,
    newPassword: input.newPassword,
  })
  if (!completed.ok) return { kind: 'failed', failure: completed.error }
  if (completed.value.kind !== 'authenticated')
    return { kind: 'failed', failure: { kind: 'challenge-failed' } }

  const consumed = await input.transactionStore.consume(input.reference, 'sign-in')
  if (!consumed.ok) return { kind: 'failed', failure: consumed.error }

  const session: NewDashboardSession = {
    subject: completed.value.subject,
    tokens: completed.value.tokens,
    expiresAt: input.sessionExpiresAt,
  }
  const established = await establishDashboardSession({
    session,
    sessionStore: input.sessionStore,
    transactionStore: input.transactionStore,
    priorTransaction: input.reference,
    priorSession: input.priorSession,
  })
  return established.ok
    ? { kind: 'authenticated', sessionReference: established.value }
    : { kind: 'failed', failure: established.error }
}
