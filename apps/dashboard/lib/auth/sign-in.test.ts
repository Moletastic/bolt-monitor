import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))

import type { AuthTransactionReference, DashboardSessionReference } from './contracts'
import { completeNewPasswordChallenge, signInWithPassword } from './sign-in'

const tokens = {
  accessToken: 'access-token',
  idToken: 'id-token',
  refreshToken: 'refresh-token',
  accessTokenExpiresAt: 1_784_117_100,
}

describe('signInWithPassword', () => {
  it('creates an opaque dashboard session only after established authentication', async () => {
    const provider = {
      beginSignIn: vi.fn().mockResolvedValue({
        ok: true,
        value: { kind: 'authenticated', subject: 'subject-1', tokens },
      }),
    }
    const sessionStore = {
      create: vi.fn().mockResolvedValue({ ok: true, value: 'opaque-session-reference' }),
    }
    const transactionStore = {
      create: vi.fn().mockResolvedValue({ ok: true, value: 'opaque-transaction-reference' }),
    }

    await expect(
      signInWithPassword({
        username: 'operator@example.com',
        password: 'password',
        sessionExpiresAt: 1_784_160_000,
        provider,
        sessionStore,
        transactionStore,
        transactionExpiresAt: 1_784_117_700,
      })
    ).resolves.toEqual({ kind: 'authenticated', sessionReference: 'opaque-session-reference' })

    expect(sessionStore.create).toHaveBeenCalledWith({
      subject: 'subject-1',
      tokens,
      expiresAt: 1_784_160_000,
    })
  })

  it('does not create a session for rejected credentials or an incomplete challenge', async () => {
    const sessionStore = { create: vi.fn() }
    const transactionStore = {
      create: vi.fn().mockResolvedValue({ ok: true, value: 'opaque-transaction-reference' }),
    }
    const rejectedProvider = {
      beginSignIn: vi
        .fn()
        .mockResolvedValue({ ok: false, error: { kind: 'authentication-failed' } }),
    }
    const challengeProvider = {
      beginSignIn: vi.fn().mockResolvedValue({
        ok: true,
        value: {
          kind: 'challenge',
          challenge: { kind: 'new-password-required', continuation: {} },
        },
      }),
    }

    await expect(
      signInWithPassword({
        username: 'operator@example.com',
        password: 'password',
        sessionExpiresAt: 1_784_160_000,
        provider: rejectedProvider,
        sessionStore,
        transactionStore,
        transactionExpiresAt: 1_784_117_700,
      })
    ).resolves.toEqual({ kind: 'failed', failure: { kind: 'authentication-failed' } })
    await expect(
      signInWithPassword({
        username: 'operator@example.com',
        password: 'password',
        sessionExpiresAt: 1_784_160_000,
        provider: challengeProvider,
        sessionStore,
        transactionStore,
        transactionExpiresAt: 1_784_117_700,
      })
    ).resolves.toEqual({
      kind: 'challenge-required',
      challenge: 'new-password-required',
      transactionReference: 'opaque-transaction-reference',
    })

    expect(sessionStore.create).not.toHaveBeenCalled()
  })

  it('stores only a server-held challenge continuation before activation', async () => {
    const continuation = { username: 'operator@example.com', session: 'provider-session' }
    const transactionStore = {
      create: vi.fn().mockResolvedValue({ ok: true, value: 'opaque-transaction-reference' }),
    }

    await expect(
      signInWithPassword({
        username: 'operator@example.com',
        password: 'temporary-password',
        sessionExpiresAt: 1_784_160_000,
        transactionExpiresAt: 1_784_117_700,
        provider: {
          beginSignIn: vi.fn().mockResolvedValue({
            ok: true,
            value: {
              kind: 'challenge',
              challenge: { kind: 'new-password-required', continuation },
            },
          }),
        },
        sessionStore: { create: vi.fn() },
        transactionStore,
      })
    ).resolves.toEqual({
      kind: 'challenge-required',
      challenge: 'new-password-required',
      transactionReference: 'opaque-transaction-reference',
    })

    expect(transactionStore.create).toHaveBeenCalledWith({
      flow: 'sign-in',
      challenge: 'new-password-required',
      providerState: continuation,
      attempts: 0,
      expiresAt: 1_784_117_700,
    })
  })

  it('consumes the invitation transaction and creates a fresh dashboard session', async () => {
    const provider = {
      answerNewPassword: vi.fn().mockResolvedValue({
        ok: true,
        value: { kind: 'authenticated', subject: 'subject-1', tokens },
      }),
    }
    const transactionStore = {
      read: vi.fn().mockResolvedValue({
        ok: true,
        value: {
          challenge: 'new-password-required',
          providerState: { session: 'provider-session' },
        },
      }),
      consume: vi.fn().mockResolvedValue({ ok: true, value: undefined }),
      invalidate: vi.fn().mockResolvedValue({ ok: true, value: undefined }),
    }
    const sessionStore = {
      create: vi.fn().mockResolvedValue({ ok: true, value: 'fresh-session-reference' }),
      invalidate: vi.fn().mockResolvedValue({ ok: true, value: undefined }),
    }

    await expect(
      completeNewPasswordChallenge({
        reference: 'opaque-transaction-reference' as AuthTransactionReference,
        newPassword: 'new-password',
        sessionExpiresAt: 1_784_160_000,
        provider,
        transactionStore,
        sessionStore,
        priorSession: 'old-session-reference' as DashboardSessionReference,
      })
    ).resolves.toEqual({ kind: 'authenticated', sessionReference: 'fresh-session-reference' })

    expect(transactionStore.consume).toHaveBeenCalledWith('opaque-transaction-reference', 'sign-in')
    expect(transactionStore.invalidate).toHaveBeenCalledWith('opaque-transaction-reference')
    expect(sessionStore.invalidate).toHaveBeenCalledWith('old-session-reference')
    expect(sessionStore.create).toHaveBeenCalledWith({
      subject: 'subject-1',
      tokens,
      expiresAt: 1_784_160_000,
    })
  })

  it('does not establish a session for an invalid or reused transaction', async () => {
    const sessionStore = { create: vi.fn(), invalidate: vi.fn() }
    const transactionStore = {
      read: vi.fn().mockResolvedValue({ ok: false, error: { kind: 'transaction-consumed' } }),
      consume: vi.fn(),
      invalidate: vi.fn(),
    }

    await expect(
      completeNewPasswordChallenge({
        reference: 'used-transaction-reference' as AuthTransactionReference,
        newPassword: 'new-password',
        sessionExpiresAt: 1_784_160_000,
        provider: { answerNewPassword: vi.fn() },
        transactionStore,
        sessionStore,
      })
    ).resolves.toEqual({ kind: 'failed', failure: { kind: 'transaction-consumed' } })

    expect(sessionStore.create).not.toHaveBeenCalled()
  })

  it.each(['transaction-expired', 'transaction-invalid'] as const)(
    'does not establish a session for a %s activation transaction',
    async (failure) => {
      const sessionStore = { create: vi.fn(), invalidate: vi.fn() }
      const transactionStore = {
        read: vi.fn().mockResolvedValue({ ok: false, error: { kind: failure } }),
        consume: vi.fn(),
        invalidate: vi.fn(),
      }

      await expect(
        completeNewPasswordChallenge({
          reference: 'invalid-transaction-reference' as AuthTransactionReference,
          newPassword: 'new-password',
          sessionExpiresAt: 1_784_160_000,
          provider: { answerNewPassword: vi.fn() },
          transactionStore,
          sessionStore,
        })
      ).resolves.toEqual({ kind: 'failed', failure: { kind: failure } })

      expect(sessionStore.create).not.toHaveBeenCalled()
      expect(transactionStore.consume).not.toHaveBeenCalled()
    }
  )
})
