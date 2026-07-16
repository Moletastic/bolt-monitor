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
    })

    expect(output).toHaveBeenCalledOnce()
    expect(JSON.parse(String(output.mock.calls[0][0]))).toMatchObject({
      event: 'auth.sign_in.failed',
      outcome: 'failure',
      component: 'dashboard-auth',
      subject: 'subject-1',
    })
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
      ])
    )
  })
})
