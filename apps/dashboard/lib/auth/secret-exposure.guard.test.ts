import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const dashboardRoot = process.cwd()

function sourceFiles(path: string): string[] {
  return readdirSync(join(dashboardRoot, path), { withFileTypes: true }).flatMap((entry) => {
    const entryPath = `${path}/${entry.name}`
    if (entry.isDirectory()) return sourceFiles(entryPath)
    if (!/\.(?:ts|tsx)$/.test(entry.name) || /\.test\.(?:ts|tsx)$/.test(entry.name)) return []
    return [entryPath]
  })
}

function source(path: string) {
  return readFileSync(join(dashboardRoot, path), 'utf8')
}

const clientModules = ['app', 'components', 'hooks']
  .flatMap(sourceFiles)
  .filter((path) => /^['"]use client['"]/.test(source(path)))

const authSource = sourceFiles('app/(auth)').concat(
  sourceFiles('lib/auth'),
  sourceFiles('lib/io/auth')
)

describe('auth secret exposure guards', () => {
  it('keeps Cognito and server session contracts out of client modules', () => {
    expect(clientModules).not.toHaveLength(0)

    for (const path of clientModules) {
      const fileSource = source(path)

      expect(fileSource, path).not.toMatch(
        /from ['"]@\/lib\/(?:io\/auth|auth\/(?:contracts|sign-in|totp|password-recovery|session-guard))['"]/
      )
      expect(fileSource, path).not.toMatch(
        /\b(?:TokenBundle|DashboardSession|AuthTransaction|AuthChallenge|TotpEnrollment)\b/
      )
      expect(fileSource, path).not.toMatch(
        /\b(?:AUTH_TRANSACTION_COOKIE|DASHBOARD_SESSION_COOKIE)\b/
      )
    }
  })

  it('rejects browser persistence of authentication material', () => {
    for (const path of clientModules) {
      const fileSource = source(path)

      expect(fileSource, path).not.toMatch(
        /(?:localStorage|sessionStorage|indexedDB|document\.cookie)[\s\S]{0,200}\b(?:token|session|challenge|password|code|totp|cookie)\b/i
      )
    }
  })

  it('keeps sensitive auth values out of RSC props and URLs', () => {
    for (const path of sourceFiles('app/(auth)').filter((path) => path.endsWith('/page.tsx'))) {
      const fileSource = source(path)

      expect(fileSource, path).not.toMatch(
        /<[A-Z][\w.]*\s[^>]*\b(?:accessToken|idToken|refreshToken|tokens|session|transaction|continuation|password|code|secret)\s*=/
      )
    }

    for (const path of sourceFiles('app/(auth)')) {
      expect(source(path), path).not.toMatch(
        /[?&](?:accessToken|idToken|refreshToken|token|session|challenge|password|code|secret)=/i
      )
    }
  })

  it('does not emit auth secrets through logs or telemetry fixtures', () => {
    for (const path of authSource) {
      expect(source(path), path).not.toMatch(/\b(?:console\.(?:log|warn|error)|telemetry)\b/)
    }
  })

  it('limits the TOTP secret to the immediate no-store server response', () => {
    const route = source('app/(auth)/totp/enroll/route.ts')

    expect(route).toContain("'Cache-Control': 'no-store'")
    expect(route).not.toMatch(/JSON\.stringify\(enrollment\)/)
    expect(route).not.toMatch(/<[^>]+\bsecret\s*=/)
  })
})
