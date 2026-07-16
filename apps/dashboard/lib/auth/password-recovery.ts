import 'server-only'

import type {
  AuthTransactionReference,
  AuthTransactionStore,
  IdentityProvider,
} from '@/lib/auth/contracts'
import type { AuthFailure } from '@/lib/auth/feedback'

export type PasswordRecoveryResult = { readonly kind: 'acknowledged' }
export type PasswordResetResult =
  | { readonly kind: 'completed' }
  | { readonly kind: 'failed'; readonly failure: AuthFailure }

type RecoveryState = { readonly username: string }

/**
 * Cognito's recovery result is deliberately collapsed to one acknowledgement so
 * callers cannot distinguish an unknown, disabled, or otherwise unusable account.
 */
export async function beginPasswordRecovery(input: {
  readonly username: string
  readonly provider: Pick<IdentityProvider, 'beginPasswordRecovery'>
  readonly transactionStore: Pick<AuthTransactionStore, 'create'>
  readonly transactionExpiresAt: number
}): Promise<PasswordRecoveryResult & { readonly transactionReference?: AuthTransactionReference }> {
  if (!input.username) return { kind: 'acknowledged' }

  const recovery = await input.provider.beginPasswordRecovery({ username: input.username })
  if (!recovery.ok) return { kind: 'acknowledged' }

  const transaction = await input.transactionStore.create({
    flow: 'password-recovery',
    challenge: 'password-recovery',
    providerState: { username: input.username } satisfies RecoveryState,
    attempts: 0,
    expiresAt: input.transactionExpiresAt,
  })
  return transaction.ok
    ? { kind: 'acknowledged', transactionReference: transaction.value }
    : { kind: 'acknowledged' }
}

/** Confirms a recovery code using the email held only in the encrypted transaction. */
export async function confirmPasswordRecovery(input: {
  readonly reference: AuthTransactionReference
  readonly code: string
  readonly newPassword: string
  readonly provider: Pick<IdentityProvider, 'confirmPasswordRecovery'>
  readonly transactionStore: Pick<AuthTransactionStore, 'read' | 'consume' | 'invalidate'>
}): Promise<PasswordResetResult> {
  if (!input.code || !input.newPassword) return { kind: 'failed', failure: 'validation-failed' }

  const transaction = await input.transactionStore.read(input.reference, 'password-recovery')
  if (
    !transaction.ok ||
    !transaction.value ||
    transaction.value.challenge !== 'password-recovery' ||
    !isRecoveryState(transaction.value.providerState)
  )
    return {
      kind: 'failed',
      failure: transaction.ok ? { kind: 'transaction-invalid' } : transaction.error,
    }

  const confirmed = await input.provider.confirmPasswordRecovery({
    username: transaction.value.providerState.username,
    code: input.code,
    newPassword: input.newPassword,
  })
  if (!confirmed.ok) return { kind: 'failed', failure: confirmed.error }

  const consumed = await input.transactionStore.consume(input.reference, 'password-recovery')
  if (!consumed.ok) return { kind: 'failed', failure: consumed.error }
  await input.transactionStore.invalidate(input.reference)
  return { kind: 'completed' }
}

function isRecoveryState(value: unknown): value is RecoveryState {
  if (!value || typeof value !== 'object') return false
  return typeof (value as Record<string, unknown>).username === 'string'
}
