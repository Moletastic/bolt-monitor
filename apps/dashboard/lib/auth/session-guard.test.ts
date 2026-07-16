import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))

import type { DashboardSessionReference } from './contracts'
import { readDashboardSession } from './session-guard'
import { err, ok } from '@/lib/result'

const reference = 'opaque-session-reference' as DashboardSessionReference
const session = {
  reference,
  subject: 'operator-subject',
  tokens: {
    accessToken: 'access-token',
    idToken: 'id-token',
    refreshToken: 'refresh-token',
    accessTokenExpiresAt: 900,
  },
  expiresAt: 1000,
  version: 1,
}

describe('dashboard session guard storage read', () => {
  it('accepts an authoritative valid session', async () => {
    const read = vi.fn().mockResolvedValue(ok(session))

    await expect(readDashboardSession(reference, { read })).resolves.toEqual(session)
    expect(read).toHaveBeenCalledWith(reference)
  })

  it('fails closed for missing and unavailable sessions', async () => {
    await expect(
      readDashboardSession(reference, { read: vi.fn().mockResolvedValue(ok(null)) })
    ).resolves.toBe(null)
    await expect(
      readDashboardSession(reference, {
        read: vi.fn().mockResolvedValue(err({ kind: 'storage-unavailable' })),
      })
    ).resolves.toBe(null)
  })
})
