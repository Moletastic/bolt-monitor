const DEFAULT_RETURN_TARGET = '/'

const AUTH_PATHS = new Set([
  '/login',
  '/activate',
  '/forgot-password',
  '/forgot-password/acknowledgement',
  '/reset-password',
  '/totp/challenge',
  '/totp/enroll',
])

/**
 * Keeps post-authentication navigation within a normalized application path.
 */
export function sanitizeReturnTarget(value: unknown): string {
  if (typeof value !== 'string' || value.length === 0) return DEFAULT_RETURN_TARGET
  if (/[%\\\u0000-\u001f\u007f]/.test(value)) return DEFAULT_RETURN_TARGET
  if (!value.startsWith('/') || value.startsWith('//')) return DEFAULT_RETURN_TARGET

  const target = new URL(value, 'https://dashboard.invalid')
  if (target.origin !== 'https://dashboard.invalid') return DEFAULT_RETURN_TARGET

  const normalized = `${target.pathname}${target.search}${target.hash}`
  if (normalized !== value || AUTH_PATHS.has(target.pathname)) return DEFAULT_RETURN_TARGET

  return normalized
}

export function loginPath(returnTarget: unknown): string {
  const target = sanitizeReturnTarget(returnTarget)
  return target === DEFAULT_RETURN_TARGET
    ? '/login'
    : `/login?returnTo=${encodeURIComponent(target)}`
}
