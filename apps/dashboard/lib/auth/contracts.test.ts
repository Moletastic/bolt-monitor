import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))

import { err, isErr, isOk, ok } from '@/lib/result'

import type {
  AuthError,
  AuthResult,
  AuthTransactionDraft,
  IdentityProvider,
  Membership,
  TokenBundle,
} from './contracts'

const tokens: TokenBundle = {
  accessToken: 'access',
  idToken: 'id',
  refreshToken: 'refresh',
  accessTokenExpiresAt: 1_700_000_000,
}

const provider: IdentityProvider = {
  beginSignIn: async () => ok({ kind: 'authenticated', subject: 'subject', tokens }),
  answerNewPassword: async () => ok({ kind: 'authenticated', subject: 'subject', tokens }),
  answerTotpChallenge: async () => ok({ kind: 'authenticated', subject: 'subject', tokens }),
  associateTotp: async () =>
    ok({
      enrollment: { secret: 'secret', issuer: 'Bolt Monitor', accountName: 'operator' },
      continuation: { session: 'opaque' },
    }),
  verifyTotpEnrollment: async () => ok({ kind: 'authenticated', subject: 'subject', tokens }),
  beginPasswordRecovery: async () => ok(undefined),
  confirmPasswordRecovery: async () => ok(undefined),
  refresh: async () => ok(tokens),
  revoke: async () => ok(undefined),
}

describe('auth contracts', () => {
  it('keeps identity-provider outcomes provider-neutral and discriminated', async () => {
    const result = await provider.beginSignIn({
      username: 'operator@example.com',
      password: 'password',
    })

    expect(isOk(result)).toBe(true)
    if (isOk(result)) {
      expect(result.value.kind).toBe('authenticated')
      if (result.value.kind === 'authenticated') {
        expect(result.value.tokens).toBe(tokens)
      }
    }
  })

  it('uses safe discriminated errors for recoverable provider and storage failures', () => {
    const result: AuthResult<void> = err<AuthError>({ kind: 'refresh-failed', retryable: true })

    expect(isErr(result)).toBe(true)
    if (isErr(result)) {
      expect(result.error).toEqual({ kind: 'refresh-failed', retryable: true })
    }
  })

  it('models opaque transaction state for server-side storage only', () => {
    const draft: AuthTransactionDraft = {
      flow: 'sign-in',
      challenge: 'software-token-mfa',
      providerState: { adapterState: 'opaque' },
      attempts: 0,
      expiresAt: 1_700_000_600,
    }
    const membership: Membership = {
      membershipId: 'membership',
      subject: 'subject',
      tenantId: 'DEFAULT',
      status: 'ACTIVE',
      role: 'ADMIN',
      authValidAfter: 1_700_000_000,
      version: 1,
      createdAt: 1_700_000_000,
      updatedAt: 1_700_000_000,
    }

    expect(draft.challenge).toBe('software-token-mfa')
    expect(membership.role).toBe('ADMIN')
  })
})
