'use server'

import { getUnixTime } from 'date-fns'
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

import { completeNewPasswordChallenge } from '@/lib/auth/sign-in'
import { feedbackForAuthFailure, type AuthFeedback } from '@/lib/auth/feedback'
import { redirectIfDashboardSession } from '@/lib/auth/session-guard'
import type { AuthTransactionReference, DashboardSessionReference } from '@/lib/auth/contracts'
import { now } from '@/lib/clock'
import { createCognitoIdentityProviderFromEnv } from '@/lib/io/auth/cognito'
import {
  AUTH_TRANSACTION_COOKIE,
  AUTH_TRANSACTION_EXPIRY_COOKIE,
  createDynamoAuthTransactionStoreFromEnv,
} from '@/lib/io/auth/transactions'
import {
  DASHBOARD_SESSION_COOKIE,
  DASHBOARD_SESSION_LIFETIME_SECONDS,
  createDynamoDashboardSessionStoreFromEnv,
} from '@/lib/io/auth/sessions'

export type ActivateFormState = { readonly feedback: AuthFeedback | null }

export async function activateInvitationAction(
  _previousState: ActivateFormState,
  formData: FormData
): Promise<ActivateFormState> {
  await redirectIfDashboardSession()
  const cookieStore = await cookies()
  const reference = cookieStore.get(AUTH_TRANSACTION_COOKIE.name)?.value
  if (!reference)
    return { feedback: feedbackForAuthFailure({ kind: 'transaction-invalid' }, 'activation') }

  const outcome = await completeNewPasswordChallenge({
    reference: reference as AuthTransactionReference,
    newPassword: String(formData.get('newPassword') ?? ''),
    sessionExpiresAt: getUnixTime(now()) + DASHBOARD_SESSION_LIFETIME_SECONDS,
    provider: createCognitoIdentityProviderFromEnv(),
    transactionStore: createDynamoAuthTransactionStoreFromEnv(),
    sessionStore: createDynamoDashboardSessionStoreFromEnv(),
    priorSession: cookieStore.get(DASHBOARD_SESSION_COOKIE.name)?.value as
      | DashboardSessionReference
      | undefined,
  })

  if (outcome.kind === 'authenticated') {
    cookieStore.set(DASHBOARD_SESSION_COOKIE.name, outcome.sessionReference, {
      httpOnly: DASHBOARD_SESSION_COOKIE.httpOnly,
      secure: DASHBOARD_SESSION_COOKIE.secure,
      sameSite: DASHBOARD_SESSION_COOKIE.sameSite,
      path: DASHBOARD_SESSION_COOKIE.path,
    })
    cookieStore.set(AUTH_TRANSACTION_EXPIRY_COOKIE.name, '', {
      httpOnly: AUTH_TRANSACTION_EXPIRY_COOKIE.httpOnly,
      secure: AUTH_TRANSACTION_EXPIRY_COOKIE.secure,
      sameSite: AUTH_TRANSACTION_EXPIRY_COOKIE.sameSite,
      path: AUTH_TRANSACTION_EXPIRY_COOKIE.path,
      maxAge: AUTH_TRANSACTION_EXPIRY_COOKIE.maxAge,
    })
    redirect('/')
  }

  return { feedback: feedbackForAuthFailure(outcome.failure, 'activation') }
}
