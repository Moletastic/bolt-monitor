import { afterEach, describe, expect, it, vi } from 'vitest'

import { emitSecurityEvent, securityEvents } from './security-events'

describe('security events', () => {
  afterEach(() => vi.restoreAllMocks())

  it('emits a fixed, secret-safe schema', () => {
    const output = vi.spyOn(console, 'info').mockImplementation(() => undefined)

    emitSecurityEvent({
      event: securityEvents.signInFailed,
      outcome: 'failure',
      subject: 'subject-1',
      correlationId: 'correlation-1',
    })

    expect(output).toHaveBeenCalledOnce()
    expect(JSON.parse(String(output.mock.calls[0][0]))).toMatchObject({
      event: 'auth.sign_in.failed',
      outcome: 'failure',
      component: 'dashboard-auth',
      subject: 'subject-1',
      correlationId: 'correlation-1',
    })
  })

  it('excludes sensitive values even when an unsafe caller supplies them', () => {
    const output = vi.spyOn(console, 'info').mockImplementation(() => undefined)
    const sensitive = {
      password: 'password-value',
      code: 'recovery-code-value',
      totpSecret: 'totp-secret-value',
      transactionID: 'transaction-id-value',
      sessionHash: 'session-hash-value',
      jwt: 'eyJhbGciOiJIUzI1NiJ9.payload.signature',
      accessToken: 'access-token-value',
      refreshToken: 'refresh-token-value',
      cookie: 'cookie-value',
      encryptionKey: 'encryption-key-value',
      requestBody: 'request-body-value',
      providerPayload: { Session: 'provider-session-value', ChallengeParameters: 'unsafe' },
    }

    emitSecurityEvent({
      event: securityEvents.signInFailed,
      outcome: 'failure',
      subject: 'subject-1',
      correlationId: 'correlation-1',
      ...sensitive,
    } as never)

    const serialized = String(output.mock.calls[0][0])
    for (const value of [
      sensitive.password,
      sensitive.code,
      sensitive.totpSecret,
      sensitive.transactionID,
      sensitive.sessionHash,
      sensitive.jwt,
      sensitive.accessToken,
      sensitive.refreshToken,
      sensitive.cookie,
      sensitive.encryptionKey,
      sensitive.requestBody,
      sensitive.providerPayload.Session,
      sensitive.providerPayload.ChallengeParameters,
    ]) {
      expect(serialized).not.toContain(value)
    }
  })

  it('defines every dashboard-auth event used by the handlers', () => {
    expect(Object.values(securityEvents)).toEqual(
      expect.arrayContaining([
        'auth.sign_in.succeeded',
        'auth.sign_in.failed',
        'auth.recovery.requested',
        'auth.recovery.completed',
        'auth.totp_enrollment.succeeded',
        'auth.totp_enrollment.failed',
        'auth.totp_challenge.succeeded',
        'auth.totp_challenge.failed',
        'auth.session.created',
        'auth.session.terminated',
        'auth.refresh.failed',
        'auth.storage.failed',
        'auth.key_loading.failed',
      ])
    )
  })

  it.each([
    [securityEvents.signInFailed, 'sign_in'],
    [securityEvents.recoveryRequested, 'recovery'],
    [securityEvents.refreshFailed, 'refresh'],
    [securityEvents.storageFailed, 'storage'],
    [securityEvents.keyLoadingFailed, 'key_loading'],
  ] as const)('emits bounded CloudWatch EMF for %s', (event, operation) => {
    const output = vi.spyOn(console, 'info').mockImplementation(() => undefined)

    emitSecurityEvent({ event, outcome: 'failure', subject: 'operator-1' })

    const record = JSON.parse(String(output.mock.calls[0][0])) as Record<string, unknown>
    expect(record).toMatchObject({
      stage: expect.any(String),
      component: 'dashboard-auth',
      operation,
      outcome: 'failure',
      AuthenticationEvents: 1,
      _aws: {
        CloudWatchMetrics: [
          {
            Namespace: 'BoltMonitor/Auth',
            Dimensions: [['stage', 'component', 'operation', 'outcome']],
            Metrics: [{ Name: 'AuthenticationEvents', Unit: 'Count' }],
          },
        ],
      },
    })
    expect(JSON.stringify(record._aws)).not.toContain('operator-1')
  })
})
