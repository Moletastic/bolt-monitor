'use server'

import { getUnixTime } from 'date-fns'
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

import { beginPasswordRecovery } from '@/lib/auth/password-recovery'
import { redirectIfDashboardSession } from '@/lib/auth/session-guard'
import { requireDashboardCsrf } from '@/lib/auth/csrf'
import { now } from '@/lib/clock'
import { createCognitoIdentityProviderFromEnv } from '@/lib/io/auth/cognito'
import {
  AUTH_TRANSACTION_COOKIE,
  AUTH_TRANSACTION_EXPIRY_COOKIE,
  AUTH_TRANSACTION_LIFETIME_SECONDS,
  createDynamoAuthTransactionStoreFromEnv,
} from '@/lib/io/auth/transactions'

export async function beginPasswordRecoveryAction(formData: FormData): Promise<void> {
  await requireDashboardCsrf()
  await redirectIfDashboardSession()
  const outcome = await beginPasswordRecovery({
    username: String(formData.get('email') ?? '').trim(),
    provider: createCognitoIdentityProviderFromEnv(),
    transactionStore: createDynamoAuthTransactionStoreFromEnv(),
    transactionExpiresAt: getUnixTime(now()) + AUTH_TRANSACTION_LIFETIME_SECONDS,
  })
  const cookieStore = await cookies()

  if (outcome.transactionReference) {
    cookieStore.set(AUTH_TRANSACTION_COOKIE.name, outcome.transactionReference, {
      httpOnly: AUTH_TRANSACTION_COOKIE.httpOnly,
      secure: AUTH_TRANSACTION_COOKIE.secure,
      sameSite: AUTH_TRANSACTION_COOKIE.sameSite,
      path: AUTH_TRANSACTION_COOKIE.path,
      maxAge: AUTH_TRANSACTION_LIFETIME_SECONDS,
    })
  } else {
    cookieStore.set(AUTH_TRANSACTION_EXPIRY_COOKIE.name, '', {
      httpOnly: AUTH_TRANSACTION_EXPIRY_COOKIE.httpOnly,
      secure: AUTH_TRANSACTION_EXPIRY_COOKIE.secure,
      sameSite: AUTH_TRANSACTION_EXPIRY_COOKIE.sameSite,
      path: AUTH_TRANSACTION_EXPIRY_COOKIE.path,
      maxAge: AUTH_TRANSACTION_EXPIRY_COOKIE.maxAge,
    })
  }

  redirect('/forgot-password/acknowledgement')
}
