import type { AuthError } from '@/lib/auth/contracts'

export type AuthFeedback =
  | 'sign-in-failed'
  | 'activation-failed'
  | 'password-reset-failed'
  | 'totp-verification-failed'

export type AuthFeedbackFlow = 'sign-in' | 'activation' | 'password-reset' | 'totp'
export type AuthFailure = AuthError | 'validation-failed'

// Public feedback intentionally collapses provider and account-state outcomes.
export function feedbackForAuthFailure(
  _failure: AuthFailure,
  flow: AuthFeedbackFlow
): AuthFeedback {
  switch (flow) {
    case 'sign-in':
      return 'sign-in-failed'
    case 'activation':
      return 'activation-failed'
    case 'password-reset':
      return 'password-reset-failed'
    case 'totp':
      return 'totp-verification-failed'
  }
}

export function messageForAuthFeedback(feedback: AuthFeedback): string {
  switch (feedback) {
    case 'sign-in-failed':
      return 'Unable to sign in with those credentials.'
    case 'activation-failed':
      return 'Unable to activate this invitation. Start again from sign in.'
    case 'password-reset-failed':
      return 'Unable to reset your password. Request new recovery instructions and try again.'
    case 'totp-verification-failed':
      return 'Unable to verify that code. Start again from sign in.'
  }
}
