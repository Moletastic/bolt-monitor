import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))

import type { DashboardSession } from '@/lib/auth/contracts'
import { ok } from '@/lib/result'

import { createAuthenticatedMonitorApiClient, createPublicHealthApiClient } from './monitor-api'

const session: DashboardSession = {
  reference: 'opaque-session-reference' as DashboardSession['reference'],
  subject: 'operator-subject',
  tokens: {
    accessToken: 'access-token',
    idToken: 'id-token',
    refreshToken: 'refresh-token',
    accessTokenExpiresAt: 4_102_444_800,
  },
  expiresAt: 4_102_444_800,
  version: 1,
}

describe('authenticated monitor API adapter', () => {
  it('reads the server session and forwards only its Bearer access token', async () => {
    const fetch = vi
      .fn()
      .mockResolvedValue(
        new Response(JSON.stringify({ status: 'success', data: { services: [] } }), { status: 200 })
      )
    const readSession = vi.fn().mockResolvedValue(ok(session))
    const client = createAuthenticatedMonitorApiClient({
      baseUrl: 'https://api.example.test/',
      fetch,
      readSession,
      sessionStore: { refresh: vi.fn(), invalidate: vi.fn() },
      identityProvider: { refresh: vi.fn() },
    })

    const result = await client.request<{ services: unknown[] }>('/api/v1/services', {
      headers: { Authorization: 'Basic caller-controlled', Cookie: 'browser-cookie' },
    })

    expect(result).toEqual(ok({ services: [] }))
    expect(readSession).toHaveBeenCalledOnce()
    const [, init] = fetch.mock.calls[0] as [string, RequestInit]
    const headers = new Headers(init.headers)
    expect(headers.get('Authorization')).toBe('Bearer access-token')
    expect(headers.get('Authorization')).not.toContain('id-token')
    expect(headers.get('Authorization')).not.toContain('refresh-token')
    expect(headers.get('Cookie')).toBeNull()
  })

  it('refreshes an expiring access token before calling the protected API', async () => {
    const refreshed = {
      ...session,
      tokens: {
        ...session.tokens,
        accessToken: 'rotated-access-token',
        accessTokenExpiresAt: 4_102_444_800,
      },
    }
    const refresh = vi.fn().mockResolvedValue(ok(refreshed))
    const fetch = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ status: 'success', data: { serviceId: 'svc' } }), {
        status: 200,
      })
    )
    const client = createAuthenticatedMonitorApiClient({
      baseUrl: 'https://api.example.test',
      fetch,
      readSession: vi
        .fn()
        .mockResolvedValue(
          ok({ ...session, tokens: { ...session.tokens, accessTokenExpiresAt: 0 } })
        ),
      sessionStore: { refresh, invalidate: vi.fn() },
      identityProvider: { refresh: vi.fn() },
    })

    await client.request('/api/v1/services/svc')

    expect(refresh).toHaveBeenCalledWith(session.reference, expect.any(Object))
    const [, init] = fetch.mock.calls[0] as [string, RequestInit]
    expect(new Headers(init.headers).get('Authorization')).toBe('Bearer rotated-access-token')
  })

  it('invalidates the local session and returns sign-in-required for application authorization denial', async () => {
    const invalidate = vi.fn().mockResolvedValue(ok(undefined))
    const client = createAuthenticatedMonitorApiClient({
      baseUrl: 'https://api.example.test',
      fetch: vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            status: 'error',
            reason: { code: 'AUTHORIZATION_DENIED', details: {} },
          }),
          { status: 403 }
        )
      ),
      readSession: vi.fn().mockResolvedValue(ok(session)),
      sessionStore: { refresh: vi.fn(), invalidate },
      identityProvider: { refresh: vi.fn() },
    })

    const result = await client.request('/api/v1/services')

    expect(result).toMatchObject({
      ok: false,
      error: { code: 'AUTHENTICATION_REQUIRED', status: 401 },
    })
    expect(invalidate).toHaveBeenCalledWith(session.reference)
  })

  it('returns sign-in-required for a non-envelope Gateway 401 without invalidating the session', async () => {
    const invalidate = vi.fn()
    const client = createAuthenticatedMonitorApiClient({
      baseUrl: 'https://api.example.test',
      fetch: vi.fn().mockResolvedValue(new Response('Unauthorized', { status: 401 })),
      readSession: vi.fn().mockResolvedValue(ok(session)),
      sessionStore: { refresh: vi.fn(), invalidate },
      identityProvider: { refresh: vi.fn() },
    })

    const result = await client.request('/api/v1/services')

    expect(result).toMatchObject({
      ok: false,
      error: { code: 'AUTHENTICATION_REQUIRED', status: 401 },
    })
    expect(invalidate).not.toHaveBeenCalled()
  })
})

describe('public health API adapter', () => {
  it('does not attach an Authorization header', async () => {
    const fetch = vi
      .fn()
      .mockResolvedValue(
        new Response(JSON.stringify({ status: 'success', data: { status: 'ok' } }), { status: 200 })
      )
    const client = createPublicHealthApiClient({ baseUrl: 'https://api.example.test', fetch })

    await expect(client.get<{ status: string }>()).resolves.toEqual(ok({ status: 'ok' }))

    expect(fetch).toHaveBeenCalledWith(
      'https://api.example.test/api/health',
      expect.objectContaining({ cache: 'no-store' })
    )
    const [, init] = fetch.mock.calls[0] as [string, RequestInit]
    expect(new Headers(init.headers).get('Authorization')).toBeNull()
  })
})
