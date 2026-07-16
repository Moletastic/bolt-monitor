import { describe, expect, it } from 'vitest'

import { loginPath, sanitizeReturnTarget } from './return-target'

describe('sanitizeReturnTarget', () => {
  it('accepts normalized root-relative application paths', () => {
    expect(sanitizeReturnTarget('/services')).toBe('/services')
    expect(sanitizeReturnTarget('/services?tab=monitors#recent')).toBe(
      '/services?tab=monitors#recent'
    )
  })

  it.each([
    'https://attacker.example',
    '//attacker.example',
    '/services/../admin',
    '/%2f%2fattacker.example',
    '/services%2fdetail',
    '/services\\attacker',
    '/services\u0000',
    '/login',
    '/activate',
    '/totp/challenge',
    '',
    undefined,
  ])('defaults unsafe targets to the dashboard root: %j', (value) => {
    expect(sanitizeReturnTarget(value)).toBe('/')
  })

  it('only appends a non-default safe target to the sign-in path', () => {
    expect(loginPath('/services')).toBe('/login?returnTo=%2Fservices')
    expect(loginPath('//attacker.example')).toBe('/login')
  })
})
