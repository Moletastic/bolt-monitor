import { formatISO } from 'date-fns'

import { now } from '@/lib/clock'

export const securityEvents = {
  signInSucceeded: 'auth.sign_in.succeeded',
  signInFailed: 'auth.sign_in.failed',
  recoveryRequested: 'auth.recovery.requested',
  recoveryCompleted: 'auth.recovery.completed',
  totpEnrollmentSucceeded: 'auth.totp_enrollment.succeeded',
  totpEnrollmentFailed: 'auth.totp_enrollment.failed',
  totpChallengeSucceeded: 'auth.totp_challenge.succeeded',
  totpChallengeFailed: 'auth.totp_challenge.failed',
  sessionCreated: 'auth.session.created',
  sessionTerminated: 'auth.session.terminated',
  refreshFailed: 'auth.refresh.failed',
} as const

type SecurityEvent = (typeof securityEvents)[keyof typeof securityEvents]

/** Emits only fixed, secret-safe audit fields. Never pass request or provider data here. */
export function emitSecurityEvent(input: {
  readonly event: SecurityEvent
  readonly outcome: 'success' | 'failure'
  readonly subject?: string
}) {
  console.info(
    JSON.stringify({
      timestamp: formatISO(now()),
      event: input.event,
      outcome: input.outcome,
      stage: process.env.SST_STAGE ?? 'unknown',
      component: 'dashboard-auth',
      ...(input.subject ? { subject: input.subject } : {}),
    })
  )
}
