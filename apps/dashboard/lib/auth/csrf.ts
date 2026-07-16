import 'server-only'

import { headers } from 'next/headers'

type HeaderReader = Pick<Headers, 'get'>

export type CsrfValidationResult =
  | { readonly ok: true }
  | { readonly ok: false; readonly reason: 'invalid-configuration' | 'origin' | 'effective-origin' }

/**
 * CloudFront supplies this header when it invokes the SST Next.js origin. Only
 * that platform boundary may supply the forwarded authority and protocol.
 */
function isTrustedCloudFrontRequest(requestHeaders: HeaderReader): boolean {
  return Boolean(requestHeaders.get('x-amz-cf-id'))
}

function singleHeader(requestHeaders: HeaderReader, name: string): string | null {
  const value = requestHeaders.get(name)?.trim()
  return value && !value.includes(',') ? value : null
}

function canonicalOrigin(origin: string): URL | null {
  if (!URL.canParse(origin)) return null
  const parsed = new URL(origin)
  return parsed.origin === origin && !parsed.username && !parsed.password ? parsed : null
}

function effectiveOrigin(requestHeaders: HeaderReader): string | null {
  if (!isTrustedCloudFrontRequest(requestHeaders)) return null

  const host = singleHeader(requestHeaders, 'x-forwarded-host')
  const protocol = singleHeader(requestHeaders, 'x-forwarded-proto')
  if (!host || (protocol !== 'http' && protocol !== 'https')) return null

  const origin = `${protocol}://${host}`
  return URL.canParse(origin) ? new URL(origin).origin : null
}

/** Validates cross-site request evidence before any auth, session, or app mutation. */
export function validateDashboardCsrf(
  requestHeaders: HeaderReader,
  configuredOrigin: string | undefined
): CsrfValidationResult {
  if (!configuredOrigin) return { ok: false, reason: 'invalid-configuration' }

  const dashboardOrigin = canonicalOrigin(configuredOrigin)
  if (!dashboardOrigin) return { ok: false, reason: 'invalid-configuration' }
  if (effectiveOrigin(requestHeaders) !== dashboardOrigin.origin) {
    return { ok: false, reason: 'effective-origin' }
  }

  const requestOrigin = singleHeader(requestHeaders, 'origin')
  if (requestOrigin) {
    return requestOrigin === dashboardOrigin.origin ? { ok: true } : { ok: false, reason: 'origin' }
  }

  // Origin-less same-origin browser navigations are the only permitted fallback.
  return singleHeader(requestHeaders, 'sec-fetch-site') === 'same-origin'
    ? { ok: true }
    : { ok: false, reason: 'origin' }
}

export async function requireDashboardCsrf(): Promise<void> {
  const result = validateDashboardCsrf(await headers(), process.env.DASHBOARD_ORIGIN)
  if (!result.ok) throw new Error('Invalid cross-site request.')
}
