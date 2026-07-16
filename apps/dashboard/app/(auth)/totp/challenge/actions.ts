'use server'

import { getUnixTime } from 'date-fns'
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

import type { AuthTransactionReference, DashboardSessionReference } from '@/lib/auth/contracts'
import { completeTotpChallenge } from '@/lib/auth/totp'
import { feedbackForAuthFailure, type AuthFeedback } from '@/lib/auth/feedback'
import { redirectIfDashboardSession } from '@/lib/auth/session-guard'
import { requireDashboardCsrf } from '@/lib/auth/csrf'
import { now } from '@/lib/clock'
import { createCognitoIdentityProviderFromEnv } from '@/lib/io/auth/cognito'
import {
  AUTH_TRANSACTION_COOKIE,
  createDynamoAuthTransactionStoreFromEnv,
} from '@/lib/io/auth/transactions'
import {
  DASHBOARD_SESSION_COOKIE,
  DASHBOARD_SESSION_LIFETIME_SECONDS,
  createDynamoDashboardSessionStoreFromEnv,
} from '@/lib/io/auth/sessions'

export type TotpChallengeFormState = { readonly feedback: AuthFeedback | null }

export async function completeTotpChallengeAction(
  _previousState: TotpChallengeFormState,
  formData: FormData
): Promise<TotpChallengeFormState> {
  await requireDashboardCsrf()
  await redirectIfDashboardSession()
  const cookieStore = await cookies()
  const reference = cookieStore.get(AUTH_TRANSACTION_COOKIE.name)?.value
  if (!reference)
    return { feedback: feedbackForAuthFailure({ kind: 'transaction-invalid' }, 'totp') }

  const outcome = await completeTotpChallenge({
    reference: reference as AuthTransactionReference,
    code: String(formData.get('code') ?? ''),
    sessionExpiresAt: getUnixTime(now()) + DASHBOARD_SESSION_LIFETIME_SECONDS,
    provider: createCognitoIdentityProviderFromEnv(),
    transactionStore: createDynamoAuthTransactionStoreFromEnv(),
    sessionStore: createDynamoDashboardSessionStoreFromEnv(),
    priorSession: cookieStore.get(DASHBOARD_SESSION_COOKIE.name)?.value as
      | DashboardSessionReference
      | undefined,
  })
  if (outcome.kind !== 'authenticated')
    return { feedback: feedbackForAuthFailure(outcome.failure, 'totp') }

  cookieStore.set(DASHBOARD_SESSION_COOKIE.name, outcome.sessionReference, DASHBOARD_SESSION_COOKIE)
  cookieStore.delete(AUTH_TRANSACTION_COOKIE.name)
  redirect('/')
}
