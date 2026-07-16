'use server'

import { getUnixTime } from 'date-fns'
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

import { signInWithPassword } from '@/lib/auth/sign-in'
import { feedbackForAuthFailure, type AuthFeedback } from '@/lib/auth/feedback'
import { redirectIfDashboardSession } from '@/lib/auth/session-guard'
import { sanitizeReturnTarget } from '@/lib/auth/return-target'
import { requireDashboardCsrf } from '@/lib/auth/csrf'
import { now } from '@/lib/clock'
import { createCognitoIdentityProviderFromEnv } from '@/lib/io/auth/cognito'
import {
  AUTH_TRANSACTION_COOKIE,
  AUTH_TRANSACTION_LIFETIME_SECONDS,
  createDynamoAuthTransactionStoreFromEnv,
} from '@/lib/io/auth/transactions'
import {
  DASHBOARD_SESSION_COOKIE,
  DASHBOARD_SESSION_LIFETIME_SECONDS,
  createDynamoDashboardSessionStoreFromEnv,
} from '@/lib/io/auth/sessions'

export type SignInFormState = { readonly feedback: AuthFeedback | null }

export async function signInAction(
  _previousState: SignInFormState,
  formData: FormData
): Promise<SignInFormState> {
  await requireDashboardCsrf()
  await redirectIfDashboardSession()
  const returnTarget = sanitizeReturnTarget(formData.get('returnTo'))
  const outcome = await signInWithPassword({
    username: String(formData.get('email') ?? '').trim(),
    password: String(formData.get('password') ?? ''),
    sessionExpiresAt: getUnixTime(now()) + DASHBOARD_SESSION_LIFETIME_SECONDS,
    transactionExpiresAt: getUnixTime(now()) + AUTH_TRANSACTION_LIFETIME_SECONDS,
    provider: createCognitoIdentityProviderFromEnv(),
    sessionStore: createDynamoDashboardSessionStoreFromEnv(),
    transactionStore: createDynamoAuthTransactionStoreFromEnv(),
  })

  if (outcome.kind === 'authenticated') {
    const cookieStore = await cookies()
    cookieStore.set(DASHBOARD_SESSION_COOKIE.name, outcome.sessionReference, {
      httpOnly: DASHBOARD_SESSION_COOKIE.httpOnly,
      secure: DASHBOARD_SESSION_COOKIE.secure,
      sameSite: DASHBOARD_SESSION_COOKIE.sameSite,
      path: DASHBOARD_SESSION_COOKIE.path,
    })
    redirect(returnTarget)
  }

  if (outcome.kind === 'challenge-required') {
    const cookieStore = await cookies()
    cookieStore.set(AUTH_TRANSACTION_COOKIE.name, outcome.transactionReference, {
      httpOnly: AUTH_TRANSACTION_COOKIE.httpOnly,
      secure: AUTH_TRANSACTION_COOKIE.secure,
      sameSite: AUTH_TRANSACTION_COOKIE.sameSite,
      path: AUTH_TRANSACTION_COOKIE.path,
      maxAge: AUTH_TRANSACTION_LIFETIME_SECONDS,
    })
    redirect(`${challengePath(outcome.challenge)}?returnTo=${encodeURIComponent(returnTarget)}`)
  }

  return {
    feedback: feedbackForAuthFailure(outcome.failure, 'sign-in'),
  }
}

function challengePath(
  challenge: 'new-password-required' | 'software-token-mfa' | 'software-token-setup'
) {
  if (challenge === 'new-password-required') return '/activate'
  return challenge === 'software-token-mfa' ? '/totp/challenge' : '/totp/enroll'
}
