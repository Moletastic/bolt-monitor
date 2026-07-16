import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))
vi.mock('@/lib/clock', () => ({ now: () => new Date('2026-07-15T12:00:00Z') }))

import type {
  AuthFlow,
  AuthTransaction,
  AuthTransactionDraft,
  AuthTransactionReference,
  DashboardSession,
  DashboardSessionReference,
  IdentityProvider,
  NewDashboardSession,
  TokenBundle,
} from './contracts'
import { beginPasswordRecovery, confirmPasswordRecovery } from './password-recovery'
import { completeNewPasswordChallenge, signInWithPassword } from './sign-in'
import { beginTotpEnrollment, completeTotpChallenge } from './totp'
import { ok, type Result } from '@/lib/result'
import { createAuthenticatedMonitorApiClient } from '@/lib/io/monitor-api'
import { expireDashboardSessionCookie } from '@/lib/io/auth/sessions'

const sessionExpiry = 1_784_160_000
const transactionExpiry = 1_784_117_700

describe('custom dashboard authentication integration', () => {
  it('signs in and forwards the server-held access token to the authenticated API', async () => {
    const sessions = new FakeSessionStore()
    const transactions = new FakeTransactionStore()
    const result = await signInWithPassword({
      username: 'operator@example.com',
      password: 'correct-password',
      sessionExpiresAt: sessionExpiry,
      transactionExpiresAt: transactionExpiry,
      provider: {
        beginSignIn: vi.fn().mockResolvedValue(ok(authenticated('established-access-token'))),
      },
      sessionStore: sessions,
      transactionStore: transactions,
    })

    expect(result.kind).toBe('authenticated')
    if (result.kind !== 'authenticated') throw new Error('expected authenticated sign-in')

    const fetch = vi
      .fn<(input: string, init?: RequestInit) => Promise<Response>>()
      .mockResolvedValue(successResponse())
    const api = apiClient(sessions, result.sessionReference as DashboardSessionReference, fetch)

    await expect(api.request<{ services: unknown[] }>('/api/v1/services')).resolves.toEqual(
      ok({ services: [] })
    )
    const firstRequest = fetch.mock.calls[0]
    if (!firstRequest) throw new Error('expected API request')
    expect(new Headers(firstRequest[1]?.headers).get('Authorization')).toBe(
      'Bearer established-access-token'
    )
  })

  it('activates an invitation, completes recovery, and establishes a session after optional TOTP setup', async () => {
    const sessions = new FakeSessionStore()
    const transactions = new FakeTransactionStore()
    const invite = await signInWithPassword({
      username: 'invited@example.com',
      password: 'temporary-password',
      sessionExpiresAt: sessionExpiry,
      transactionExpiresAt: transactionExpiry,
      provider: {
        beginSignIn: vi.fn().mockResolvedValue(
          ok({
            kind: 'challenge',
            challenge: { kind: 'new-password-required', continuation: { session: 'invite' } },
          })
        ),
      },
      sessionStore: sessions,
      transactionStore: transactions,
    })
    expect(invite.kind).toBe('challenge-required')
    if (invite.kind !== 'challenge-required') throw new Error('expected invitation challenge')

    await expect(
      completeNewPasswordChallenge({
        reference: invite.transactionReference,
        newPassword: 'replacement-password',
        sessionExpiresAt: sessionExpiry,
        provider: {
          answerNewPassword: vi.fn().mockResolvedValue(ok(authenticated('activated-access-token'))),
        },
        transactionStore: transactions,
        sessionStore: sessions,
      })
    ).resolves.toMatchObject({ kind: 'authenticated', subject: 'operator-subject' })

    const recovery = await beginPasswordRecovery({
      username: 'operator@example.com',
      transactionExpiresAt: transactionExpiry,
      provider: { beginPasswordRecovery: vi.fn().mockResolvedValue(ok(undefined)) },
      transactionStore: transactions,
    })
    expect(recovery.transactionReference).toBeDefined()
    if (!recovery.transactionReference) throw new Error('expected recovery transaction')
    const confirmPasswordRecoveryProvider = {
      confirmPasswordRecovery: vi.fn().mockResolvedValue(ok(undefined)),
    }
    await expect(
      confirmPasswordRecovery({
        reference: recovery.transactionReference,
        code: '123456',
        newPassword: 'recovered-password',
        provider: confirmPasswordRecoveryProvider,
        transactionStore: transactions,
      })
    ).resolves.toEqual({ kind: 'completed' })
    expect(confirmPasswordRecoveryProvider.confirmPasswordRecovery).toHaveBeenCalledWith({
      username: 'operator@example.com',
      code: '123456',
      newPassword: 'recovered-password',
    })

    const setup = await transactions.create({
      flow: 'sign-in',
      challenge: 'software-token-setup',
      providerState: { session: 'totp-start' },
      attempts: 0,
      expiresAt: transactionExpiry,
    })
    if (!setup.ok) throw new Error('expected TOTP transaction')
    const enrollment = await beginTotpEnrollment({
      reference: setup.value,
      transactionExpiresAt: transactionExpiry,
      provider: {
        associateTotp: vi.fn().mockResolvedValue(
          ok({
            enrollment: {
              secret: 'immediate-setup-secret',
              issuer: 'Bolt Monitor',
              accountName: 'operator@example.com',
            },
            continuation: { session: 'totp-associated' },
          })
        ),
      },
      transactionStore: transactions,
    })
    expect(enrollment.kind).toBe('enrollment-ready')
    if (enrollment.kind !== 'enrollment-ready') throw new Error('expected TOTP enrollment')
    expect(transactions.persistedState()).not.toContain('immediate-setup-secret')

    await expect(
      completeTotpChallenge({
        reference: enrollment.transactionReference,
        code: '123456',
        sessionExpiresAt: sessionExpiry,
        provider: {
          answerTotpChallenge: vi.fn(),
          verifyTotpEnrollment: vi.fn().mockResolvedValue(ok(authenticated('totp-access-token'))),
        },
        transactionStore: transactions,
        sessionStore: sessions,
      })
    ).resolves.toMatchObject({ kind: 'authenticated', subject: 'operator-subject' })
  })

  it('expires the browser cookie and denies dashboard API access after logout or explicit session expiry', async () => {
    const sessions = new FakeSessionStore()
    const reference = await sessions.create(newSession('active-access-token'))
    if (!reference.ok) throw new Error('expected active session')
    const cookies = { set: vi.fn() }

    await sessions.invalidate(reference.value)
    expireDashboardSessionCookie(cookies)
    const afterLogoutFetch = vi.fn()
    await expect(
      apiClient(sessions, reference.value, afterLogoutFetch).request('/api/v1/services')
    ).resolves.toMatchObject({ ok: false, error: { code: 'AUTHENTICATION_REQUIRED' } })
    expect(afterLogoutFetch).not.toHaveBeenCalled()
    expect(cookies.set).toHaveBeenCalledWith(
      '__Host-bolt-session',
      '',
      expect.objectContaining({ maxAge: 0 })
    )

    const expired = await sessions.create({
      ...newSession('expired-access-token'),
      expiresAt: 1_784_116_800,
    })
    if (!expired.ok) throw new Error('expected expired session fixture')
    const afterExpiryFetch = vi.fn()
    await expect(
      apiClient(sessions, expired.value, afterExpiryFetch).request('/api/v1/services')
    ).resolves.toMatchObject({ ok: false, error: { code: 'AUTHENTICATION_REQUIRED' } })
    expect(afterExpiryFetch).not.toHaveBeenCalled()
  })

  it('uses one refresh result for concurrent authenticated API requests', async () => {
    const sessions = new FakeSessionStore()
    const created = await sessions.create({
      ...newSession('expiring-access-token'),
      tokens: { ...tokens('expiring-access-token'), accessTokenExpiresAt: 0 },
    })
    if (!created.ok) throw new Error('expected session')
    const releaseRefresh = deferred<void>()
    const provider = {
      refresh: vi.fn().mockImplementation(async () => {
        await releaseRefresh.promise
        return ok(tokens('rotated-access-token'))
      }),
    }
    const fetch = vi.fn().mockImplementation(() => Promise.resolve(successResponse()))
    const first = apiClient(sessions, created.value, fetch, provider).request('/api/v1/services')
    const second = apiClient(sessions, created.value, fetch, provider).request('/api/v1/services')

    await vi.waitFor(() => expect(provider.refresh).toHaveBeenCalledOnce())
    releaseRefresh.resolve()
    await expect(Promise.all([first, second])).resolves.toEqual([
      ok({ services: [] }),
      ok({ services: [] }),
    ])
    expect(provider.refresh).toHaveBeenCalledOnce()
    for (const [, init] of fetch.mock.calls) {
      expect(new Headers((init as RequestInit).headers).get('Authorization')).toBe(
        'Bearer rotated-access-token'
      )
    }
  })
})

class FakeTransactionStore {
  private readonly transactions = new Map<AuthTransactionReference, AuthTransaction>()
  private sequence = 0

  async create(draft: AuthTransactionDraft) {
    const reference = `transaction-${++this.sequence}` as AuthTransactionReference
    this.transactions.set(reference, { reference, ...draft })
    return ok(reference)
  }

  async read(reference: AuthTransactionReference, flow: AuthFlow) {
    const transaction = this.transactions.get(reference)
    return ok(transaction?.flow === flow ? transaction : null)
  }

  async consume(reference: AuthTransactionReference, flow: AuthFlow) {
    const transaction = this.transactions.get(reference)
    if (!transaction || transaction.flow !== flow)
      throw new Error('unexpected transaction consumption')
    this.transactions.delete(reference)
    return ok(undefined)
  }

  async invalidate(reference: AuthTransactionReference) {
    this.transactions.delete(reference)
    return ok(undefined)
  }

  persistedState() {
    return JSON.stringify([...this.transactions.values()])
  }
}

class FakeSessionStore {
  private readonly sessions = new Map<DashboardSessionReference, DashboardSession>()
  private sequence = 0
  private refreshing: Promise<
    Result<DashboardSession, { readonly kind: 'session-invalid' }>
  > | null = null

  async create(session: NewDashboardSession) {
    const reference = `session-${++this.sequence}` as DashboardSessionReference
    this.sessions.set(reference, { reference, ...session, version: 1 })
    return ok(reference)
  }

  async read(reference: DashboardSessionReference) {
    const session = this.sessions.get(reference)
    return ok(session && session.expiresAt > 1_784_116_800 ? session : null)
  }

  async refresh(reference: DashboardSessionReference, provider: Pick<IdentityProvider, 'refresh'>) {
    if (this.refreshing) return this.refreshing
    this.refreshing = this.refreshSession(reference, provider)
    try {
      return await this.refreshing
    } finally {
      this.refreshing = null
    }
  }

  async invalidate(reference: DashboardSessionReference) {
    this.sessions.delete(reference)
    return ok(undefined)
  }

  private async refreshSession(
    reference: DashboardSessionReference,
    provider: Pick<IdentityProvider, 'refresh'>
  ) {
    const session = this.sessions.get(reference)
    if (!session) return { ok: false as const, error: { kind: 'session-invalid' as const } }
    const refreshed = await provider.refresh({ refreshToken: session.tokens.refreshToken })
    if (!refreshed.ok) return { ok: false as const, error: { kind: 'session-invalid' as const } }
    const next = { ...session, tokens: refreshed.value, version: session.version + 1 }
    this.sessions.set(reference, next)
    return ok(next)
  }
}

function apiClient(
  sessions: FakeSessionStore,
  reference: DashboardSessionReference,
  fetch: (input: string, init?: RequestInit) => Promise<Response>,
  identityProvider = { refresh: vi.fn().mockResolvedValue(ok(tokens('unused-access-token'))) }
) {
  return createAuthenticatedMonitorApiClient({
    baseUrl: 'https://api.example.test',
    fetch,
    readSession: () => sessions.read(reference),
    sessionStore: sessions,
    identityProvider,
  })
}

function authenticated(accessToken: string) {
  return {
    kind: 'authenticated' as const,
    subject: 'operator-subject',
    tokens: tokens(accessToken),
  }
}

function newSession(accessToken: string): NewDashboardSession {
  return { subject: 'operator-subject', tokens: tokens(accessToken), expiresAt: sessionExpiry }
}

function tokens(accessToken: string): TokenBundle {
  return {
    accessToken,
    idToken: 'server-only-id-token',
    refreshToken: 'server-only-refresh-token',
    accessTokenExpiresAt: 1_784_117_100,
  }
}

function successResponse() {
  return new Response(JSON.stringify({ status: 'success', data: { services: [] } }), {
    status: 200,
  })
}

function deferred<T>() {
  let resolve: (value: T) => void = () => undefined
  const promise = new Promise<T>((complete) => {
    resolve = complete
  })
  return { promise, resolve }
}
