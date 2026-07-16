import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))

import type { AuthTransactionReference } from './contracts'
import { beginTotpEnrollment, completeTotpChallenge } from './totp'

const tokens = {
  accessToken: 'access',
  idToken: 'id',
  refreshToken: 'refresh',
  accessTokenExpiresAt: 100,
}
const reference = 'old-transaction' as AuthTransactionReference

describe('TOTP authentication', () => {
  it('keeps the enrollment secret out of the replacement transaction', async () => {
    const transactionStore = {
      read: vi.fn().mockResolvedValue({
        ok: true,
        value: { challenge: 'software-token-setup', providerState: { session: 'old' } },
      }),
      create: vi.fn().mockResolvedValue({ ok: true, value: 'new-transaction' }),
      invalidate: vi.fn().mockResolvedValue({ ok: true, value: undefined }),
    }
    const result = await beginTotpEnrollment({
      reference,
      transactionExpiresAt: 700,
      provider: {
        associateTotp: vi.fn().mockResolvedValue({
          ok: true,
          value: {
            enrollment: {
              secret: 'totp-secret',
              issuer: 'Bolt Monitor',
              accountName: 'operator@example.com',
            },
            continuation: { session: 'associated' },
          },
        }),
      },
      transactionStore,
    })

    expect(result).toEqual({
      kind: 'enrollment-ready',
      enrollment: {
        secret: 'totp-secret',
        issuer: 'Bolt Monitor',
        accountName: 'operator@example.com',
      },
      transactionReference: 'new-transaction',
    })
    expect(transactionStore.create).toHaveBeenCalledWith(
      expect.objectContaining({ providerState: { session: 'associated' } })
    )
    expect(JSON.stringify(transactionStore.create.mock.calls)).not.toContain('totp-secret')
    expect(transactionStore.invalidate).toHaveBeenCalledWith(reference)
  })

  it('verifies enrollment before creating a dashboard session', async () => {
    const transactionStore = {
      read: vi.fn().mockResolvedValue({
        ok: true,
        value: { challenge: 'software-token-setup', providerState: { session: 'associated' } },
      }),
      consume: vi.fn().mockResolvedValue({ ok: true, value: undefined }),
      invalidate: vi.fn().mockResolvedValue({ ok: true, value: undefined }),
    }
    const sessionStore = {
      create: vi.fn().mockResolvedValue({ ok: true, value: 'new-session' }),
      invalidate: vi.fn(),
    }
    const result = await completeTotpChallenge({
      reference,
      code: '123456',
      sessionExpiresAt: 900,
      provider: {
        answerTotpChallenge: vi.fn(),
        verifyTotpEnrollment: vi.fn().mockResolvedValue({
          ok: true,
          value: { kind: 'authenticated', subject: 'subject', tokens },
        }),
      },
      transactionStore,
      sessionStore,
    })

    expect(result).toEqual({ kind: 'authenticated', sessionReference: 'new-session' })
    expect(transactionStore.consume).toHaveBeenCalledWith(reference, 'sign-in')
    expect(sessionStore.create).toHaveBeenCalledWith({ subject: 'subject', tokens, expiresAt: 900 })
  })

  it('does not establish a session when an MFA code is rejected', async () => {
    const sessionStore = { create: vi.fn(), invalidate: vi.fn() }
    const result = await completeTotpChallenge({
      reference,
      code: 'bad-code',
      sessionExpiresAt: 900,
      provider: {
        answerTotpChallenge: vi
          .fn()
          .mockResolvedValue({ ok: false, error: { kind: 'totp-failed' } }),
        verifyTotpEnrollment: vi.fn(),
      },
      transactionStore: {
        read: vi.fn().mockResolvedValue({
          ok: true,
          value: { challenge: 'software-token-mfa', providerState: {} },
        }),
        consume: vi.fn(),
        invalidate: vi.fn(),
      },
      sessionStore,
    })

    expect(result).toEqual({ kind: 'failed' })
    expect(sessionStore.create).not.toHaveBeenCalled()
  })
})
