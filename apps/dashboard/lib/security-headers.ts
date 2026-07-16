type SecurityHeaderOptions = {
  readonly isProduction: boolean
  readonly isHttps: boolean
}

const permissionsPolicy = [
  'accelerometer=()',
  'autoplay=()',
  'camera=()',
  'display-capture=()',
  'geolocation=()',
  'gyroscope=()',
  'microphone=()',
  'payment=()',
  'picture-in-picture=()',
  'usb=()',
].join(', ')

export function securityHeaders(nonce: string, options: SecurityHeaderOptions): Headers {
  const headers = new Headers({
    'Content-Security-Policy': options.isProduction
      ? [
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
      : [
          "default-src 'self'",
          "script-src 'self' 'unsafe-inline' 'unsafe-eval'",
          "style-src 'self' 'unsafe-inline'",
          "img-src 'self' data: blob:",
          "font-src 'self'",
          "connect-src 'self'",
          "frame-ancestors 'none'",
          "base-uri 'self'",
          "form-action 'self'",
          "object-src 'none'",
        ].join('; '),
    'Permissions-Policy': permissionsPolicy,
    'Referrer-Policy': 'no-referrer',
    'X-Content-Type-Options': 'nosniff',
  })

  if (options.isProduction && options.isHttps) {
    headers.set('Strict-Transport-Security', 'max-age=31536000; includeSubDomains')
  }

  return headers
}
