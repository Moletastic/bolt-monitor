import 'server-only'

import type { Result } from '@/lib/result'

declare const authTransactionReference: unique symbol
declare const dashboardSessionReference: unique symbol

/** Opaque browser-facing reference. Its raw value is never persisted. */
export type AuthTransactionReference = string & {
  readonly [authTransactionReference]: 'AuthTransactionReference'
}

/** Opaque `__Host-` cookie value. Its raw value is never persisted. */
export type DashboardSessionReference = string & {
  readonly [dashboardSessionReference]: 'DashboardSessionReference'
}

export type AuthFlow = 'sign-in' | 'password-recovery'

export type AuthChallenge =
  | { readonly kind: 'new-password-required'; readonly continuation: unknown }
  | { readonly kind: 'software-token-mfa'; readonly continuation: unknown }
  | { readonly kind: 'software-token-setup'; readonly continuation: unknown }

export interface TokenBundle {
  readonly accessToken: string
  readonly idToken: string
  readonly refreshToken: string
  readonly accessTokenExpiresAt: number
}

export type SignInOutcome =
  | { readonly kind: 'authenticated'; readonly subject: string; readonly tokens: TokenBundle }
  | { readonly kind: 'challenge'; readonly challenge: AuthChallenge }

export interface TotpEnrollment {
  /** Server-only enrollment material, shown only during immediate setup. */
  readonly secret: string
  readonly issuer: string
  readonly accountName: string
}

export interface AuthTransaction {
  readonly reference: AuthTransactionReference
  readonly flow: AuthFlow
  readonly challenge: AuthChallenge['kind']
  /** Encrypted by the storage adapter before persistence. */
  readonly providerState: unknown
  readonly attempts: number
  readonly expiresAt: number
}

export interface AuthTransactionDraft {
  readonly flow: AuthFlow
  readonly challenge: AuthChallenge['kind']
  readonly providerState: unknown
  readonly attempts: number
  readonly expiresAt: number
}

export interface DashboardSession {
  readonly reference: DashboardSessionReference
  readonly subject: string
  /** Encrypted by the storage adapter before persistence. */
  readonly tokens: TokenBundle
  readonly expiresAt: number
  readonly version: number
  readonly refreshOwner?: string
  readonly refreshLeaseUntil?: number
}

export interface NewDashboardSession {
  readonly subject: string
  readonly tokens: TokenBundle
  readonly expiresAt: number
}

export interface Membership {
  readonly membershipId: string
  readonly subject: string
  readonly tenantId: 'DEFAULT'
  readonly status: 'ACTIVE'
  readonly role: 'ADMIN'
  readonly authValidAfter: number
  readonly version: number
  readonly createdAt: number
  readonly updatedAt: number
}

export type AuthError =
  | { readonly kind: 'authentication-failed' }
  | { readonly kind: 'challenge-failed' }
  | { readonly kind: 'recovery-failed' }
  | { readonly kind: 'totp-failed' }
  | { readonly kind: 'transaction-invalid' }
  | { readonly kind: 'transaction-expired' }
  | { readonly kind: 'transaction-consumed' }
  | { readonly kind: 'transaction-flow-mismatch' }
  | { readonly kind: 'session-invalid' }
  | { readonly kind: 'refresh-failed'; readonly retryable: boolean }
  | { readonly kind: 'membership-unavailable' }
  | { readonly kind: 'provider-unavailable' }
  | { readonly kind: 'storage-unavailable' }

export type AuthResult<T> = Result<T, AuthError>

export interface IdentityProvider {
  beginSignIn(input: {
    readonly username: string
    readonly password: string
  }): Promise<AuthResult<SignInOutcome>>
  answerNewPassword(input: {
    readonly continuation: unknown
    readonly newPassword: string
  }): Promise<AuthResult<SignInOutcome>>
  answerTotpChallenge(input: {
    readonly continuation: unknown
    readonly code: string
  }): Promise<AuthResult<SignInOutcome>>
  associateTotp(input: { readonly continuation: unknown }): Promise<AuthResult<TotpEnrollment>>
  verifyTotpEnrollment(input: {
    readonly continuation: unknown
    readonly code: string
  }): Promise<AuthResult<SignInOutcome>>
  beginPasswordRecovery(input: { readonly username: string }): Promise<AuthResult<void>>
  confirmPasswordRecovery(input: {
    readonly username: string
    readonly code: string
    readonly newPassword: string
  }): Promise<AuthResult<void>>
  refresh(input: { readonly refreshToken: string }): Promise<AuthResult<TokenBundle>>
  revoke(input: { readonly refreshToken: string }): Promise<AuthResult<void>>
}

export interface AuthTransactionStore {
  create(draft: AuthTransactionDraft): Promise<AuthResult<AuthTransactionReference>>
  read(reference: AuthTransactionReference): Promise<AuthResult<AuthTransaction | null>>
  consume(reference: AuthTransactionReference): Promise<AuthResult<void>>
  invalidate(reference: AuthTransactionReference): Promise<AuthResult<void>>
}

export interface DashboardSessionStore {
  create(session: NewDashboardSession): Promise<AuthResult<DashboardSessionReference>>
  read(reference: DashboardSessionReference): Promise<AuthResult<DashboardSession | null>>
  replace(
    reference: DashboardSessionReference,
    session: NewDashboardSession
  ): Promise<AuthResult<DashboardSessionReference>>
  invalidate(reference: DashboardSessionReference): Promise<AuthResult<void>>
}

export interface MembershipStore {
  readBySubject(subject: string): Promise<AuthResult<Membership | null>>
}
