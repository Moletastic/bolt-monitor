import 'server-only'

import {
  AssociateSoftwareTokenCommand,
  CognitoIdentityProviderClient,
  ConfirmForgotPasswordCommand,
  ForgotPasswordCommand,
  GetTokensFromRefreshTokenCommand,
  InitiateAuthCommand,
  RespondToAuthChallengeCommand,
  RevokeTokenCommand,
  VerifySoftwareTokenCommand,
  type AuthenticationResultType,
} from '@aws-sdk/client-cognito-identity-provider'
import { getUnixTime } from 'date-fns'
import { createHmac } from 'node:crypto'

import { now } from '@/lib/clock'
import type {
  AuthChallenge,
  AuthError,
  AuthResult,
  IdentityProvider,
  SignInOutcome,
  TokenBundle,
  TotpEnrollment,
} from '@/lib/auth/contracts'
import { err, ok } from '@/lib/result'

type CognitoClient = Pick<CognitoIdentityProviderClient, 'send'>
type CognitoChallengeName = 'NEW_PASSWORD_REQUIRED' | 'SOFTWARE_TOKEN_MFA' | 'MFA_SETUP'

type CognitoContinuation = {
  readonly username: string
  session: string
  readonly challenge: CognitoChallengeName
}

export interface CognitoIdentityProviderOptions {
  readonly clientId: string
  readonly clientSecret?: string
  readonly client?: CognitoClient
}

/** Cognito command and exception details are contained in this I/O adapter. */
export function createCognitoIdentityProvider(
  options: CognitoIdentityProviderOptions
): IdentityProvider {
  const client = options.client ?? new CognitoIdentityProviderClient({})
  const clientSecret = options.clientSecret

  const secretHash = (username: string): string | undefined => {
    if (!clientSecret) return undefined
    return createHmac('sha256', clientSecret)
      .update(`${username}${options.clientId}`)
      .digest('base64')
  }
  const challengeResponses = (username: string, values: Record<string, string>) => {
    const hash = secretHash(username)
    return hash
      ? { USERNAME: username, SECRET_HASH: hash, ...values }
      : { USERNAME: username, ...values }
  }
  const continuation = (
    value: unknown,
    expected: CognitoChallengeName
  ): CognitoContinuation | null =>
    isContinuation(value) && value.challenge === expected ? value : null

  return {
    async beginSignIn({ username, password }) {
      return invoke(
        async () =>
          toOutcome(
            await client.send(
              new InitiateAuthCommand({
                AuthFlow: 'USER_PASSWORD_AUTH',
                ClientId: options.clientId,
                AuthParameters: challengeResponses(username, { PASSWORD: password }),
              })
            ),
            username
          ),
        'authentication-failed'
      )
    },
    async answerNewPassword({ continuation: value, newPassword }) {
      const state = continuation(value, 'NEW_PASSWORD_REQUIRED')
      if (!state) return err({ kind: 'challenge-failed' })
      return invoke(
        async () =>
          toOutcome(
            await client.send(
              new RespondToAuthChallengeCommand({
                ChallengeName: 'NEW_PASSWORD_REQUIRED',
                ClientId: options.clientId,
                Session: state.session,
                ChallengeResponses: challengeResponses(state.username, {
                  NEW_PASSWORD: newPassword,
                }),
              })
            ),
            state.username
          ),
        'challenge-failed'
      )
    },
    async answerTotpChallenge({ continuation: value, code }) {
      const state = continuation(value, 'SOFTWARE_TOKEN_MFA')
      if (!state) return err({ kind: 'challenge-failed' })
      return invoke(
        async () =>
          toOutcome(
            await client.send(
              new RespondToAuthChallengeCommand({
                ChallengeName: 'SOFTWARE_TOKEN_MFA',
                ClientId: options.clientId,
                Session: state.session,
                ChallengeResponses: challengeResponses(state.username, {
                  SOFTWARE_TOKEN_MFA_CODE: code,
                }),
              })
            ),
            state.username
          ),
        'totp-failed'
      )
    },
    async associateTotp({ continuation: value }) {
      const state = continuation(value, 'MFA_SETUP')
      if (!state) return err({ kind: 'totp-failed' })
      return invoke(async () => {
        const response = await client.send(
          new AssociateSoftwareTokenCommand({ Session: state.session })
        )
        if (!response.SecretCode || !response.Session) return err({ kind: 'provider-unavailable' })
        // The server-held continuation advances to the association session.
        state.session = response.Session
        const enrollment: TotpEnrollment = {
          secret: response.SecretCode,
          issuer: 'Bolt Monitor',
          accountName: state.username,
        }
        return ok(enrollment)
      }, 'totp-failed')
    },
    async verifyTotpEnrollment({ continuation: value, code }) {
      const state = continuation(value, 'MFA_SETUP')
      if (!state) return err({ kind: 'totp-failed' })
      return invoke(async () => {
        const verified = await client.send(
          new VerifySoftwareTokenCommand({ Session: state.session, UserCode: code })
        )
        if (verified.Status !== 'SUCCESS' || !verified.Session) return err({ kind: 'totp-failed' })
        return toOutcome(
          await client.send(
            new RespondToAuthChallengeCommand({
              ChallengeName: 'MFA_SETUP',
              ClientId: options.clientId,
              Session: verified.Session,
              ChallengeResponses: challengeResponses(state.username, {}),
            })
          ),
          state.username
        )
      }, 'totp-failed')
    },
    async beginPasswordRecovery({ username }) {
      return invoke(async () => {
        await client.send(
          new ForgotPasswordCommand({
            ClientId: options.clientId,
            Username: username,
            SecretHash: secretHash(username),
          })
        )
        return ok(undefined)
      }, 'recovery-failed')
    },
    async confirmPasswordRecovery({ username, code, newPassword }) {
      return invoke(async () => {
        await client.send(
          new ConfirmForgotPasswordCommand({
            ClientId: options.clientId,
            Username: username,
            ConfirmationCode: code,
            Password: newPassword,
            SecretHash: secretHash(username),
          })
        )
        return ok(undefined)
      }, 'recovery-failed')
    },
    async refresh({ refreshToken }) {
      return invoke(async () => {
        const response = await client.send(
          new GetTokensFromRefreshTokenCommand({
            ClientId: options.clientId,
            RefreshToken: refreshToken,
          })
        )
        const tokens = tokenBundle(response.AuthenticationResult)
        return tokens ? ok(tokens) : err({ kind: 'refresh-failed', retryable: false })
      }, 'refresh-failed')
    },
    async revoke({ refreshToken }) {
      return invoke(async () => {
        await client.send(
          new RevokeTokenCommand({
            ClientId: options.clientId,
            Token: refreshToken,
            ClientSecret: clientSecret,
          })
        )
        return ok(undefined)
      }, 'provider-unavailable')
    },
  }
}

/** Build the server adapter from SST-injected, non-public dashboard configuration. */
export function createCognitoIdentityProviderFromEnv(): IdentityProvider {
  return createCognitoIdentityProvider({
    clientId: process.env.COGNITO_DASHBOARD_CLIENT_ID ?? '',
    clientSecret: process.env.COGNITO_DASHBOARD_CLIENT_SECRET,
  })
}

function toOutcome(
  response: {
    readonly AuthenticationResult?: AuthenticationResultType
    readonly ChallengeName?: string
    readonly Session?: string
  },
  username: string
): AuthResult<SignInOutcome> {
  if (response.AuthenticationResult) {
    const tokens = tokenBundle(response.AuthenticationResult)
    if (!tokens) return err({ kind: 'provider-unavailable' })
    const subject = subjectFromIdToken(tokens.idToken)
    return subject
      ? ok({ kind: 'authenticated', subject, tokens })
      : err({ kind: 'provider-unavailable' })
  }
  const challenge = toChallenge(response.ChallengeName, response.Session, username)
  return challenge ? ok({ kind: 'challenge', challenge }) : err({ kind: 'challenge-failed' })
}

function toChallenge(
  name: string | undefined,
  session: string | undefined,
  username: string
): AuthChallenge | null {
  const challenge = challengeName(name)
  if (!challenge || !session) return null
  return {
    kind:
      challenge === 'NEW_PASSWORD_REQUIRED'
        ? 'new-password-required'
        : challenge === 'SOFTWARE_TOKEN_MFA'
          ? 'software-token-mfa'
          : 'software-token-setup',
    continuation: { username, session, challenge },
  }
}

function challengeName(value: string | undefined): CognitoChallengeName | null {
  return value === 'NEW_PASSWORD_REQUIRED' ||
    value === 'SOFTWARE_TOKEN_MFA' ||
    value === 'MFA_SETUP'
    ? value
    : null
}

function isContinuation(value: unknown): value is CognitoContinuation {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Record<string, unknown>
  return (
    typeof candidate.username === 'string' &&
    typeof candidate.session === 'string' &&
    challengeName(typeof candidate.challenge === 'string' ? candidate.challenge : undefined) !==
      null
  )
}

function tokenBundle(result: AuthenticationResultType | undefined): TokenBundle | null {
  if (!result?.AccessToken || !result.IdToken || !result.RefreshToken || !result.ExpiresIn)
    return null
  return {
    accessToken: result.AccessToken,
    idToken: result.IdToken,
    refreshToken: result.RefreshToken,
    accessTokenExpiresAt: getUnixTime(now()) + result.ExpiresIn,
  }
}

function subjectFromIdToken(token: string): string | null {
  try {
    const payload = token.split('.')[1]
    if (!payload) return null
    const decoded: unknown = JSON.parse(Buffer.from(payload, 'base64url').toString('utf8'))
    if (!decoded || typeof decoded !== 'object') return null
    const subject = (decoded as Record<string, unknown>).sub
    return typeof subject === 'string' && subject.length > 0 ? subject : null
  } catch {
    return null
  }
}

async function invoke<T>(
  operation: () => Promise<AuthResult<T>>,
  failure: AuthError['kind']
): Promise<AuthResult<T>> {
  try {
    return await operation()
  } catch (cause) {
    return err(mapCognitoError(cause, failure))
  }
}

function mapCognitoError(cause: unknown, failure: AuthError['kind']): AuthError {
  const name = cause instanceof Error ? cause.name : ''
  if (failure === 'refresh-failed') {
    return {
      kind: 'refresh-failed',
      retryable:
        name === 'TooManyRequestsException' ||
        name === 'InternalErrorException' ||
        name === 'ServiceUnavailableException',
    }
  }
  return { kind: failure }
}
