import 'server-only'

import type {
  AuthResult,
  AuthTransactionReference,
  AuthTransactionStore,
  DashboardSessionReference,
  DashboardSessionStore,
  NewDashboardSession,
} from '@/lib/auth/contracts'

/**
 * Converts a successful provider authentication into exactly one fresh dashboard
 * session. References from the preceding browser state are removed before the
 * new identifier is created, so they cannot be retained for session fixation.
 */
export async function establishDashboardSession(input: {
  readonly session: NewDashboardSession
  readonly sessionStore: Pick<DashboardSessionStore, 'create' | 'invalidate'>
  readonly transactionStore: Pick<AuthTransactionStore, 'invalidate'>
  readonly priorSession?: DashboardSessionReference
  readonly priorTransaction?: AuthTransactionReference
}): Promise<AuthResult<DashboardSessionReference>> {
  if (input.priorTransaction) {
    const invalidated = await input.transactionStore.invalidate(input.priorTransaction)
    if (!invalidated.ok) return invalidated
  }
  if (input.priorSession) {
    const invalidated = await input.sessionStore.invalidate(input.priorSession)
    if (!invalidated.ok) return invalidated
  }
  return input.sessionStore.create(input.session)
}
