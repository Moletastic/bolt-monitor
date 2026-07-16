import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))

import { validateDashboardCsrf } from './csrf'

const dashboardOrigin = 'https://dashboard.example.com'

function cloudFrontHeaders(entries: Record<string, string> = {}) {
  return new Headers({
    'x-amz-cf-id': 'cloudfront-request-id',
    'x-forwarded-host': 'dashboard.example.com',
    'x-forwarded-proto': 'https',
    ...entries,
  })
}

describe('dashboard CSRF validation', () => {
  it('accepts an exact canonical Origin and trusted effective origin', () => {
    expect(
      validateDashboardCsrf(cloudFrontHeaders({ Origin: dashboardOrigin }), dashboardOrigin)
    ).toEqual({ ok: true })
  })

  it('rejects a mismatched Origin before a state-changing request can proceed', () => {
    expect(
      validateDashboardCsrf(
        cloudFrontHeaders({ Origin: 'https://attacker.example' }),
        dashboardOrigin
      )
    ).toEqual({ ok: false, reason: 'origin' })
  })

  it('rejects a mismatched effective forwarded host or protocol', () => {
    expect(
      validateDashboardCsrf(
        cloudFrontHeaders({ 'x-forwarded-host': 'attacker.example', Origin: dashboardOrigin }),
        dashboardOrigin
      )
    ).toEqual({ ok: false, reason: 'effective-origin' })
    expect(
      validateDashboardCsrf(
        cloudFrontHeaders({ 'x-forwarded-proto': 'http', Origin: dashboardOrigin }),
        dashboardOrigin
      )
    ).toEqual({ ok: false, reason: 'effective-origin' })
  })

  it('does not trust forwarded authority without the CloudFront boundary', () => {
    expect(
      validateDashboardCsrf(
        new Headers({
          'x-forwarded-host': 'dashboard.example.com',
          'x-forwarded-proto': 'https',
          Origin: dashboardOrigin,
        }),
        dashboardOrigin
      )
    ).toEqual({ ok: false, reason: 'effective-origin' })
  })

  it('permits a missing Origin only for same-origin Fetch Metadata', () => {
    expect(
      validateDashboardCsrf(cloudFrontHeaders({ 'sec-fetch-site': 'same-origin' }), dashboardOrigin)
    ).toEqual({ ok: true })
    expect(validateDashboardCsrf(cloudFrontHeaders(), dashboardOrigin)).toEqual({
      ok: false,
      reason: 'origin',
    })
    expect(
      validateDashboardCsrf(cloudFrontHeaders({ 'sec-fetch-site': 'same-site' }), dashboardOrigin)
    ).toEqual({ ok: false, reason: 'origin' })
  })

  it('fails closed for absent or non-canonical configuration and malformed evidence', () => {
    expect(
      validateDashboardCsrf(cloudFrontHeaders({ Origin: dashboardOrigin }), undefined)
    ).toEqual({
      ok: false,
      reason: 'invalid-configuration',
    })
    expect(
      validateDashboardCsrf(cloudFrontHeaders({ Origin: dashboardOrigin }), `${dashboardOrigin}/`)
    ).toEqual({ ok: false, reason: 'invalid-configuration' })
    expect(
      validateDashboardCsrf(
        cloudFrontHeaders({ Origin: `${dashboardOrigin},https://attacker.example` }),
        dashboardOrigin
      )
    ).toEqual({ ok: false, reason: 'origin' })
  })
})
