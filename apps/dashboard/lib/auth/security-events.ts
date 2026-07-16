import { formatISO, getTime } from 'date-fns'

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
  storageFailed: 'auth.storage.failed',
  keyLoadingFailed: 'auth.key_loading.failed',
} as const

type SecurityEvent = (typeof securityEvents)[keyof typeof securityEvents]

const metricOperations: Partial<Record<SecurityEvent, string>> = {
  [securityEvents.signInFailed]: 'sign_in',
  [securityEvents.recoveryRequested]: 'recovery',
  [securityEvents.refreshFailed]: 'refresh',
  [securityEvents.storageFailed]: 'storage',
  [securityEvents.keyLoadingFailed]: 'key_loading',
}

/** Emits only fixed, secret-safe audit fields. Never pass request or provider data here. */
export function emitSecurityEvent(input: {
  readonly event: SecurityEvent
  readonly outcome: 'success' | 'failure'
  readonly subject?: string
  readonly correlationId?: string
}) {
  const timestamp = now()
  const stage = process.env.AUTH_STAGE ?? process.env.SST_STAGE ?? 'unknown'
  const operation = metricOperations[input.event]
  console.info(
    JSON.stringify({
      timestamp: formatISO(timestamp),
      event: input.event,
      outcome: input.outcome,
      stage,
      component: 'dashboard-auth',
      ...(input.subject ? { subject: input.subject } : {}),
      ...(input.correlationId ? { correlationId: input.correlationId } : {}),
      ...(operation
        ? {
            operation,
            AuthenticationEvents: 1,
            _aws: {
              Timestamp: getTime(timestamp),
              CloudWatchMetrics: [
                {
                  Namespace: 'BoltMonitor/Auth',
                  Dimensions: [['stage', 'component', 'operation', 'outcome']],
                  Metrics: [{ Name: 'AuthenticationEvents', Unit: 'Count' }],
                },
              ],
            },
          }
        : {}),
    })
  )
}
