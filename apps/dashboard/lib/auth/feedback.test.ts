import { describe, expect, it } from 'vitest'

import { feedbackForAuthFailure, messageForAuthFeedback } from './feedback'

describe('operator-safe auth feedback', () => {
  it.each([
    ['sign-in', { kind: 'authentication-failed' }, 'sign-in-failed'],
    ['sign-in', { kind: 'provider-unavailable' }, 'sign-in-failed'],
    ['activation', { kind: 'transaction-expired' }, 'activation-failed'],
    ['activation', { kind: 'challenge-failed' }, 'activation-failed'],
    ['password-reset', { kind: 'recovery-failed' }, 'password-reset-failed'],
    ['totp', { kind: 'totp-failed' }, 'totp-verification-failed'],
    ['totp', 'validation-failed', 'totp-verification-failed'],
  ] as const)('maps %s failures to typed public feedback', (flow, failure, expected) => {
    expect(feedbackForAuthFailure(failure, flow)).toBe(expected)
  })

  it('does not disclose provider exception or account-state details', () => {
    const messages = [
      messageForAuthFeedback(feedbackForAuthFailure({ kind: 'authentication-failed' }, 'sign-in')),
      messageForAuthFeedback(feedbackForAuthFailure({ kind: 'membership-unavailable' }, 'sign-in')),
      messageForAuthFeedback(feedbackForAuthFailure({ kind: 'transaction-expired' }, 'activation')),
    ].join(' ')

    expect(messages).not.toMatch(/Cognito|Exception|disabled|inactive|unknown account|membership/i)
  })
})
