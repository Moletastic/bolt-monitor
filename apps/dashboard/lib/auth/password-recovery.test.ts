import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))

import type { AuthTransactionReference } from './contracts'
import { beginPasswordRecovery, confirmPasswordRecovery } from './password-recovery'

describe('password recovery', () => {
  it('acknowledges recovery requests regardless of Cognito outcome', async () => {
    const transactionStore = { create: vi.fn() }

    await expect(
      beginPasswordRecovery({
        username: 'operator@example.com',
        provider: { beginPasswordRecovery: vi.fn().mockResolvedValue({ ok: false }) },
        transactionStore,
        transactionExpiresAt: 1_784_117_700,
      })
    ).resolves.toEqual({ kind: 'acknowledged' })
    expect(transactionStore.create).not.toHaveBeenCalled()
  })

  it('stores the email only as encrypted server-side transaction state', async () => {
    const transactionStore = {
      create: vi.fn().mockResolvedValue({ ok: true, value: 'opaque-recovery-reference' }),
    }

    await expect(
      beginPasswordRecovery({
        username: 'operator@example.com',
        provider: {
          beginPasswordRecovery: vi.fn().mockResolvedValue({ ok: true, value: undefined }),
        },
        transactionStore,
        transactionExpiresAt: 1_784_117_700,
      })
    ).resolves.toEqual({ kind: 'acknowledged', transactionReference: 'opaque-recovery-reference' })
    expect(transactionStore.create).toHaveBeenCalledWith({
      flow: 'password-recovery',
      challenge: 'password-recovery',
      providerState: { username: 'operator@example.com' },
      attempts: 0,
      expiresAt: 1_784_117_700,
    })
  })

  it('uses the opaque transaction to confirm recovery and exposes no submitted secrets', async () => {
    const provider = {
      confirmPasswordRecovery: vi.fn().mockResolvedValue({ ok: true, value: undefined }),
    }
    const transactionStore = {
      read: vi.fn().mockResolvedValue({
        ok: true,
        value: {
          challenge: 'password-recovery',
          providerState: { username: 'operator@example.com' },
        },
      }),
      consume: vi.fn().mockResolvedValue({ ok: true, value: undefined }),
      invalidate: vi.fn().mockResolvedValue({ ok: true, value: undefined }),
    }

    await expect(
      confirmPasswordRecovery({
        reference: 'opaque-recovery-reference' as AuthTransactionReference,
        code: '123456',
        newPassword: 'new-password',
        provider,
        transactionStore,
      })
    ).resolves.toEqual({ kind: 'completed' })
    expect(provider.confirmPasswordRecovery).toHaveBeenCalledWith({
      username: 'operator@example.com',
      code: '123456',
      newPassword: 'new-password',
    })
  })

  it.each(['transaction-expired', 'transaction-invalid'] as const)(
    'rejects a %s recovery transaction without confirming a password',
    async (failure) => {
      const provider = { confirmPasswordRecovery: vi.fn() }
      const transactionStore = {
        read: vi.fn().mockResolvedValue({ ok: false, error: { kind: failure } }),
        consume: vi.fn(),
        invalidate: vi.fn(),
      }

      await expect(
        confirmPasswordRecovery({
          reference: 'wrong-flow-reference' as AuthTransactionReference,
          code: '123456',
          newPassword: 'new-password',
          provider,
          transactionStore,
        })
      ).resolves.toEqual({ kind: 'failed', failure: { kind: failure } })

      expect(provider.confirmPasswordRecovery).not.toHaveBeenCalled()
      expect(transactionStore.consume).not.toHaveBeenCalled()
    }
  )
})
