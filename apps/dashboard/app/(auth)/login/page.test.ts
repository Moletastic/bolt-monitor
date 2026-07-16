import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const loginPage = readFileSync(join(process.cwd(), 'app/(auth)/login/page.tsx'), 'utf8')
const loginAction = readFileSync(join(process.cwd(), 'app/(auth)/login/actions.ts'), 'utf8')
const monitoringLayout = readFileSync(join(process.cwd(), 'app/(monitoring)/layout.tsx'), 'utf8')

describe('public sign-in route', () => {
  it('renders the custom form outside the monitoring shell', () => {
    expect(loginPage).toContain('<SignInForm />')
    expect(loginPage).not.toContain('AppShell')
    expect(loginPage).not.toContain('PollingProvider')
    expect(monitoringLayout).toContain('PollingProvider')
  })

  it('uses only server-side authentication adapters and never reads the monitor API', () => {
    expect(loginAction).toContain("'use server'")
    expect(loginAction).toContain('createCognitoIdentityProviderFromEnv')
    expect(loginAction).toContain('createDynamoDashboardSessionStoreFromEnv')
    expect(loginAction).not.toContain('/api/v1/')
    expect(loginPage).not.toContain('/api/v1/')
  })

  it('redirects after establishing an authenticated dashboard session', () => {
    expect(loginAction).toContain('DASHBOARD_SESSION_COOKIE.name')
    expect(loginAction).toContain('cookieStore.set')
    expect(loginAction).toContain("redirect('/')")
  })
})
