'use server'

import { getUnixTime } from 'date-fns'
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

import type { AuthTransactionReference, DashboardSessionReference } from '@/lib/auth/contracts'
import { completeTotpChallenge } from '@/lib/auth/totp'
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

export type TotpChallengeFormState = { readonly message: string | null }

const TOTP_FAILED_MESSAGE = 'Unable to verify that code. Start again from sign in.'

export async function completeTotpChallengeAction(
  _previousState: TotpChallengeFormState,
  formData: FormData
): Promise<TotpChallengeFormState> {
  const cookieStore = await cookies()
  const reference = cookieStore.get(AUTH_TRANSACTION_COOKIE.name)?.value
  if (!reference) return { message: TOTP_FAILED_MESSAGE }

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
  if (outcome.kind !== 'authenticated') return { message: TOTP_FAILED_MESSAGE }

  cookieStore.set(DASHBOARD_SESSION_COOKIE.name, outcome.sessionReference, DASHBOARD_SESSION_COOKIE)
  cookieStore.delete(AUTH_TRANSACTION_COOKIE.name)
  redirect('/')
}
