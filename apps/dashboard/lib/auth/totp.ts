import 'server-only'

import type {
  AuthTransactionReference,
  AuthTransactionStore,
  DashboardSessionReference,
  DashboardSessionStore,
  IdentityProvider,
  NewDashboardSession,
  TotpEnrollment,
} from '@/lib/auth/contracts'
import { establishDashboardSession } from '@/lib/io/auth/authentication-state'

export type TotpEnrollmentResult =
  | {
      readonly kind: 'enrollment-ready'
      readonly enrollment: TotpEnrollment
      readonly transactionReference: AuthTransactionReference
    }
  | { readonly kind: 'failed' }

export type TotpChallengeResult =
  | { readonly kind: 'authenticated'; readonly sessionReference: string }
  | { readonly kind: 'failed' }

/**
 * The secret is returned only to the immediate non-RSC enrollment response.
 * The replacement transaction contains only Cognito's advanced continuation.
 */
export async function beginTotpEnrollment(input: {
  readonly reference: AuthTransactionReference
  readonly transactionExpiresAt: number
  readonly provider: Pick<IdentityProvider, 'associateTotp'>
  readonly transactionStore: Pick<AuthTransactionStore, 'read' | 'create' | 'invalidate'>
}): Promise<TotpEnrollmentResult> {
  const transaction = await input.transactionStore.read(input.reference, 'sign-in')
  if (
    !transaction.ok ||
    !transaction.value ||
    transaction.value.challenge !== 'software-token-setup'
  )
    return { kind: 'failed' }

  const association = await input.provider.associateTotp({
    continuation: transaction.value.providerState,
  })
  if (!association.ok) return { kind: 'failed' }

  const replacement = await input.transactionStore.create({
    flow: 'sign-in',
    challenge: 'software-token-setup',
    providerState: association.value.continuation,
    attempts: 0,
    expiresAt: input.transactionExpiresAt,
  })
  if (!replacement.ok) return { kind: 'failed' }
  const invalidated = await input.transactionStore.invalidate(input.reference)
  if (!invalidated.ok) return { kind: 'failed' }

  return {
    kind: 'enrollment-ready',
    enrollment: association.value.enrollment,
    transactionReference: replacement.value,
  }
}

/** Completes either a required TOTP challenge or a just-associated software token. */
export async function completeTotpChallenge(input: {
  readonly reference: AuthTransactionReference
  readonly code: string
  readonly sessionExpiresAt: number
  readonly provider: Pick<IdentityProvider, 'answerTotpChallenge' | 'verifyTotpEnrollment'>
  readonly transactionStore: Pick<AuthTransactionStore, 'read' | 'consume' | 'invalidate'>
  readonly sessionStore: Pick<DashboardSessionStore, 'create' | 'invalidate'>
  readonly priorSession?: DashboardSessionReference
}): Promise<TotpChallengeResult> {
  if (!input.code) return { kind: 'failed' }
  const transaction = await input.transactionStore.read(input.reference, 'sign-in')
  if (
    !transaction.ok ||
    !transaction.value ||
    (transaction.value.challenge !== 'software-token-mfa' &&
      transaction.value.challenge !== 'software-token-setup')
  )
    return { kind: 'failed' }

  const completed =
    transaction.value.challenge === 'software-token-mfa'
      ? await input.provider.answerTotpChallenge({
          continuation: transaction.value.providerState,
          code: input.code,
        })
      : await input.provider.verifyTotpEnrollment({
          continuation: transaction.value.providerState,
          code: input.code,
        })
  if (!completed.ok || completed.value.kind !== 'authenticated') return { kind: 'failed' }

  const consumed = await input.transactionStore.consume(input.reference, 'sign-in')
  if (!consumed.ok) return { kind: 'failed' }
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
    : { kind: 'failed' }
}
