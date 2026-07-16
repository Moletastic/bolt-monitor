import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))

import { establishDashboardSession } from './authentication-state'

const session = {
  subject: 'cognito-subject',
  tokens: {
    accessToken: 'access-token',
    idToken: 'id-token',
    refreshToken: 'refresh-token',
    accessTokenExpiresAt: 1_784_117_100,
  },
  expiresAt: 1_784_160_000,
}

describe('establishDashboardSession', () => {
  it('removes prior transaction and session references before issuing a fresh session', async () => {
    const transactionStore = {
      invalidate: vi.fn().mockResolvedValue({ ok: true, value: undefined }),
    }
    const sessionStore = {
      invalidate: vi.fn().mockResolvedValue({ ok: true, value: undefined }),
      create: vi.fn().mockResolvedValue({ ok: true, value: 'fresh-session' }),
    }

    await expect(
      establishDashboardSession({
        session,
        transactionStore,
        sessionStore,
        priorTransaction: 'prior-transaction' as never,
        priorSession: 'prior-session' as never,
      })
    ).resolves.toEqual({ ok: true, value: 'fresh-session' })

    expect(transactionStore.invalidate).toHaveBeenCalledWith('prior-transaction')
    expect(sessionStore.invalidate).toHaveBeenCalledWith('prior-session')
    expect(sessionStore.create).toHaveBeenCalledWith(session)
    expect(
      transactionStore.invalidate.mock.invocationCallOrder[0] <
        sessionStore.invalidate.mock.invocationCallOrder[0]
    ).toBe(true)
    expect(
      sessionStore.invalidate.mock.invocationCallOrder[0] <
        sessionStore.create.mock.invocationCallOrder[0]
    ).toBe(true)
  })

  it('does not issue a new session when prior state invalidation fails', async () => {
    const transactionStore = {
      invalidate: vi.fn().mockResolvedValue({ ok: false, error: { kind: 'storage-unavailable' } }),
    }
    const sessionStore = {
      invalidate: vi.fn(),
      create: vi.fn(),
    }

    await expect(
      establishDashboardSession({
        session,
        transactionStore,
        sessionStore,
        priorTransaction: 'prior-transaction' as never,
      })
    ).resolves.toEqual({ ok: false, error: { kind: 'storage-unavailable' } })

    expect(sessionStore.invalidate).not.toHaveBeenCalled()
    expect(sessionStore.create).not.toHaveBeenCalled()
  })
})
