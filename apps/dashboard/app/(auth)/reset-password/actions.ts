'use server'

import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

import { confirmPasswordRecovery } from '@/lib/auth/password-recovery'
import { feedbackForAuthFailure, type AuthFeedback } from '@/lib/auth/feedback'
import { redirectIfDashboardSession } from '@/lib/auth/session-guard'
import { requireDashboardCsrf } from '@/lib/auth/csrf'
import { emitSecurityEvent, securityEvents } from '@/lib/auth/security-events'
import type { AuthTransactionReference } from '@/lib/auth/contracts'
import { createCognitoIdentityProviderFromEnv } from '@/lib/io/auth/cognito'
import {
  AUTH_TRANSACTION_COOKIE,
  AUTH_TRANSACTION_EXPIRY_COOKIE,
  createDynamoAuthTransactionStoreFromEnv,
} from '@/lib/io/auth/transactions'

export type ResetPasswordFormState = { readonly feedback: AuthFeedback | null }

export async function resetPasswordAction(
  _previousState: ResetPasswordFormState,
  formData: FormData
): Promise<ResetPasswordFormState> {
  await requireDashboardCsrf()
  await redirectIfDashboardSession()
  const cookieStore = await cookies()
  const reference = cookieStore.get(AUTH_TRANSACTION_COOKIE.name)?.value
  if (!reference) {
    emitSecurityEvent({ event: securityEvents.recoveryCompleted, outcome: 'failure' })
    return { feedback: feedbackForAuthFailure({ kind: 'transaction-invalid' }, 'password-reset') }
  }

  const outcome = await confirmPasswordRecovery({
    reference: reference as AuthTransactionReference,
    code: String(formData.get('code') ?? ''),
    newPassword: String(formData.get('newPassword') ?? ''),
    provider: createCognitoIdentityProviderFromEnv(),
    transactionStore: createDynamoAuthTransactionStoreFromEnv(),
  })
  if (outcome.kind !== 'completed') {
    emitSecurityEvent({ event: securityEvents.recoveryCompleted, outcome: 'failure' })
    return { feedback: feedbackForAuthFailure(outcome.failure, 'password-reset') }
  }

  emitSecurityEvent({ event: securityEvents.recoveryCompleted, outcome: 'success' })
  cookieStore.set(AUTH_TRANSACTION_EXPIRY_COOKIE.name, '', {
    httpOnly: AUTH_TRANSACTION_EXPIRY_COOKIE.httpOnly,
    secure: AUTH_TRANSACTION_EXPIRY_COOKIE.secure,
    sameSite: AUTH_TRANSACTION_EXPIRY_COOKIE.sameSite,
    path: AUTH_TRANSACTION_EXPIRY_COOKIE.path,
    maxAge: AUTH_TRANSACTION_EXPIRY_COOKIE.maxAge,
  })
  redirect('/login')
}
