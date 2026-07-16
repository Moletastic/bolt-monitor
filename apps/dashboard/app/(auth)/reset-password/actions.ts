'use server'

import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

import { confirmPasswordRecovery } from '@/lib/auth/password-recovery'
import type { AuthTransactionReference } from '@/lib/auth/contracts'
import { createCognitoIdentityProviderFromEnv } from '@/lib/io/auth/cognito'
import {
  AUTH_TRANSACTION_COOKIE,
  AUTH_TRANSACTION_EXPIRY_COOKIE,
  createDynamoAuthTransactionStoreFromEnv,
} from '@/lib/io/auth/transactions'

export type ResetPasswordFormState = { readonly message: string | null }

const RESET_FAILED_MESSAGE =
  'Unable to reset your password. Request new recovery instructions and try again.'

export async function resetPasswordAction(
  _previousState: ResetPasswordFormState,
  formData: FormData
): Promise<ResetPasswordFormState> {
  const cookieStore = await cookies()
  const reference = cookieStore.get(AUTH_TRANSACTION_COOKIE.name)?.value
  if (!reference) return { message: RESET_FAILED_MESSAGE }

  const outcome = await confirmPasswordRecovery({
    reference: reference as AuthTransactionReference,
    code: String(formData.get('code') ?? ''),
    newPassword: String(formData.get('newPassword') ?? ''),
    provider: createCognitoIdentityProviderFromEnv(),
    transactionStore: createDynamoAuthTransactionStoreFromEnv(),
  })
  if (outcome.kind !== 'completed') return { message: RESET_FAILED_MESSAGE }

  cookieStore.set(AUTH_TRANSACTION_EXPIRY_COOKIE.name, '', {
    httpOnly: AUTH_TRANSACTION_EXPIRY_COOKIE.httpOnly,
    secure: AUTH_TRANSACTION_EXPIRY_COOKIE.secure,
    sameSite: AUTH_TRANSACTION_EXPIRY_COOKIE.sameSite,
    path: AUTH_TRANSACTION_EXPIRY_COOKIE.path,
    maxAge: AUTH_TRANSACTION_EXPIRY_COOKIE.maxAge,
  })
  redirect('/login')
}
