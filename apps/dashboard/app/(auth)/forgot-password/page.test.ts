import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const forgotPasswordPage = readFileSync(
  join(process.cwd(), 'app/(auth)/forgot-password/page.tsx'),
  'utf8'
)
const acknowledgementPage = readFileSync(
  join(process.cwd(), 'app/(auth)/forgot-password/acknowledgement/page.tsx'),
  'utf8'
)
const forgotPasswordAction = readFileSync(
  join(process.cwd(), 'app/(auth)/forgot-password/actions.ts'),
  'utf8'
)
const resetPasswordPage = readFileSync(
  join(process.cwd(), 'app/(auth)/reset-password/page.tsx'),
  'utf8'
)
const resetPasswordAction = readFileSync(
  join(process.cwd(), 'app/(auth)/reset-password/actions.ts'),
  'utf8'
)

describe('public password recovery routes', () => {
  it('renders a non-enumerating acknowledgement outside the monitoring shell', () => {
    expect(acknowledgementPage).toContain('If recovery is available for that address')
    expect(acknowledgementPage).not.toContain('AppShell')
    expect(acknowledgementPage).not.toContain('PollingProvider')
    expect(forgotPasswordPage).not.toContain('/api/v1/')
  })

  it('uses an opaque transaction cookie rather than URL or returned email state', () => {
    expect(forgotPasswordAction).toContain('AUTH_TRANSACTION_COOKIE')
    expect(forgotPasswordAction).toContain("redirect('/forgot-password/acknowledgement')")
    expect(forgotPasswordAction).not.toContain('searchParams')
    expect(forgotPasswordAction).not.toContain('return {')
    expect(resetPasswordAction).toContain('AUTH_TRANSACTION_COOKIE')
    expect(resetPasswordAction).not.toContain('searchParams')
    expect(resetPasswordAction).not.toContain('email')
  })

  it('keeps recovery code and password inputs out of URLs and returned form state', () => {
    expect(resetPasswordPage).toContain('<ResetPasswordForm />')
    expect(resetPasswordAction).toContain("formData.get('code')")
    expect(resetPasswordAction).toContain("formData.get('newPassword')")
    expect(resetPasswordAction).not.toContain('redirect(`/')
    expect(resetPasswordAction).toContain('Promise<ResetPasswordFormState>')
  })
})
