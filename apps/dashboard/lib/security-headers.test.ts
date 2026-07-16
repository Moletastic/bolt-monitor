import { describe, expect, it } from 'vitest'

import { securityHeaders } from './security-headers'

const nonce = 'fixed-test-nonce'
const productionCsp = [
  "default-src 'self'",
  `script-src 'self' 'nonce-${nonce}' 'strict-dynamic'`,
  `style-src 'self' 'nonce-${nonce}'`,
  "img-src 'self' data: blob:",
  "font-src 'self'",
  "connect-src 'self'",
  "frame-ancestors 'none'",
  "base-uri 'self'",
  "form-action 'self'",
  "object-src 'none'",
].join('; ')

describe('dashboard security headers', () => {
  it('sets a nonce-based production CSP and the required restrictive headers', () => {
    const headers = securityHeaders(nonce, { isProduction: true, isHttps: true })

    expect(headers.get('Content-Security-Policy')).toBe(productionCsp)
    expect(headers.get('Content-Security-Policy')).not.toContain(
      "script-src 'self' 'unsafe-inline'"
    )
    expect(headers.get('X-Content-Type-Options')).toBe('nosniff')
    expect(headers.get('Referrer-Policy')).toBe('no-referrer')
    expect(headers.get('Permissions-Policy')).toBe(
      'accelerometer=(), autoplay=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), microphone=(), payment=(), picture-in-picture=(), usb=()'
    )
    expect(headers.get('Strict-Transport-Security')).toBe('max-age=31536000; includeSubDomains')
  })

  it('does not enable HSTS unless the production response is served over HTTPS', () => {
    expect(
      securityHeaders(nonce, { isProduction: true, isHttps: false }).get(
        'Strict-Transport-Security'
      )
    ).toBeNull()
    expect(
      securityHeaders(nonce, { isProduction: false, isHttps: true }).get(
        'Strict-Transport-Security'
      )
    ).toBeNull()
  })
})
