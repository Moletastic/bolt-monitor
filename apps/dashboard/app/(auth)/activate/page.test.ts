import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const activationPage = readFileSync(join(process.cwd(), 'app/(auth)/activate/page.tsx'), 'utf8')
const activationAction = readFileSync(join(process.cwd(), 'app/(auth)/activate/actions.ts'), 'utf8')

describe('invitation activation route', () => {
  it('renders the custom activation form outside the monitoring shell', () => {
    expect(activationPage).toContain('<ActivationForm />')
    expect(activationPage).not.toContain('AppShell')
    expect(activationPage).not.toContain('PollingProvider')
  })

  it('uses only an opaque transaction cookie and server-side auth adapters', () => {
    expect(activationAction).toContain("'use server'")
    expect(activationAction).toContain('AUTH_TRANSACTION_COOKIE')
    expect(activationAction).toContain('completeNewPasswordChallenge')
    expect(activationAction).toContain('createDynamoAuthTransactionStoreFromEnv')
    expect(activationAction).toContain('createDynamoDashboardSessionStoreFromEnv')
    expect(activationAction).not.toContain('searchParams')
    expect(activationAction).not.toContain('/api/v1/')
  })

  it('replaces the authentication transaction with an established session before redirecting', () => {
    expect(activationAction).toContain('DASHBOARD_SESSION_COOKIE.name')
    expect(activationAction).toContain('AUTH_TRANSACTION_EXPIRY_COOKIE.name')
    expect(activationAction).toContain("redirect('/')")
  })
})
