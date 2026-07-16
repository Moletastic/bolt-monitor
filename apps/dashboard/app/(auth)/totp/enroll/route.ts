import { getUnixTime } from 'date-fns'
import { cookies } from 'next/headers'
import { NextResponse } from 'next/server'

import type { AuthTransactionReference } from '@/lib/auth/contracts'
import { beginTotpEnrollment, completeTotpChallenge } from '@/lib/auth/totp'
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

// This route deliberately bypasses RSC: the TOTP secret exists only in this immediate HTML response.
export async function GET(request: Request) {
  const cookieStore = await cookies()
  const reference = cookieStore.get(AUTH_TRANSACTION_COOKIE.name)?.value
  if (!reference) return failure(request)
  const enrollment = await beginTotpEnrollment({
    reference: reference as AuthTransactionReference,
    transactionExpiresAt: getUnixTime(now()) + AUTH_TRANSACTION_LIFETIME_SECONDS,
    provider: createCognitoIdentityProviderFromEnv(),
    transactionStore: createDynamoAuthTransactionStoreFromEnv(),
  })
  if (enrollment.kind !== 'enrollment-ready') return failure(request)

  const response = new NextResponse(enrollmentHtml(enrollment.enrollment), {
    headers: { 'Cache-Control': 'no-store', 'Content-Type': 'text/html; charset=utf-8' },
  })
  response.cookies.set(AUTH_TRANSACTION_COOKIE.name, enrollment.transactionReference, {
    ...AUTH_TRANSACTION_COOKIE,
    maxAge: AUTH_TRANSACTION_LIFETIME_SECONDS,
  })
  return response
}

export async function POST(request: Request) {
  const cookieStore = await cookies()
  const reference = cookieStore.get(AUTH_TRANSACTION_COOKIE.name)?.value
  if (!reference) return failure(request)
  const formData = await request.formData()
  const completed = await completeTotpChallenge({
    reference: reference as AuthTransactionReference,
    code: String(formData.get('code') ?? ''),
    sessionExpiresAt: getUnixTime(now()) + DASHBOARD_SESSION_LIFETIME_SECONDS,
    provider: createCognitoIdentityProviderFromEnv(),
    transactionStore: createDynamoAuthTransactionStoreFromEnv(),
    sessionStore: createDynamoDashboardSessionStoreFromEnv(),
  })
  if (completed.kind !== 'authenticated') return failure(request)
  const response = NextResponse.redirect(new URL('/', request.url), 303)
  response.cookies.set(
    DASHBOARD_SESSION_COOKIE.name,
    completed.sessionReference,
    DASHBOARD_SESSION_COOKIE
  )
  response.cookies.delete(AUTH_TRANSACTION_COOKIE.name)
  return response
}

function failure(request: Request) {
  return NextResponse.redirect(new URL('/login', request.url), 303)
}

function enrollmentHtml(enrollment: {
  readonly secret: string
  readonly issuer: string
  readonly accountName: string
}) {
  return `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="robots" content="noindex"><title>Set up authenticator</title></head><body><main><h1>Set up your authenticator</h1><p>Add this key to your authenticator app now. It will not be shown again.</p><p><code>${escapeHtml(enrollment.secret)}</code></p><p>Account: ${escapeHtml(enrollment.accountName)} (${escapeHtml(enrollment.issuer)})</p><form method="post"><label for="code">Authentication code</label><input id="code" name="code" autocomplete="one-time-code" inputmode="numeric" required><button type="submit">Verify and continue</button></form></main></body></html>`
}

function escapeHtml(value: string) {
  return value.replace(
    /[&<>'"]/g,
    (character) =>
      ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' })[character] ?? ''
  )
}
