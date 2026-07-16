import { NextRequest, NextResponse } from 'next/server'

import { securityHeaders } from '@/lib/security-headers'

function nonce(): string {
  const bytes = new Uint8Array(16)
  crypto.getRandomValues(bytes)
  return btoa(String.fromCharCode(...bytes))
}

function isTrustedHttpsRequest(request: NextRequest): boolean {
  return request.headers.has('x-amz-cf-id') && request.headers.get('x-forwarded-proto') === 'https'
}

export function middleware(request: NextRequest) {
  const requestHeaders = new Headers(request.headers)
  const requestNonce = nonce()
  const headers = securityHeaders(requestNonce, {
    isProduction: process.env.NODE_ENV === 'production',
    isHttps: isTrustedHttpsRequest(request),
  })

  // Next reads the request CSP to apply the nonce to its generated scripts and styles.
  requestHeaders.set('Content-Security-Policy', headers.get('Content-Security-Policy') ?? '')
  requestHeaders.set('x-nonce', requestNonce)

  const response = NextResponse.next({ request: { headers: requestHeaders } })
  headers.forEach((value, name) => response.headers.set(name, value))
  return response
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
}
