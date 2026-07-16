import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))

import {
  AssociateSoftwareTokenCommand,
  ConfirmForgotPasswordCommand,
  ForgotPasswordCommand,
  GetTokensFromRefreshTokenCommand,
  InitiateAuthCommand,
  RespondToAuthChallengeCommand,
  RevokeTokenCommand,
  VerifySoftwareTokenCommand,
} from '@aws-sdk/client-cognito-identity-provider'

import { isErr, isOk } from '@/lib/result'

import { createCognitoIdentityProvider } from './cognito'

const idToken = `header.${Buffer.from(JSON.stringify({ sub: 'operator-subject' })).toString('base64url')}.signature`
const authenticationResult = {
  AccessToken: 'access-token',
  IdToken: idToken,
  RefreshToken: 'refresh-token',
  ExpiresIn: 60,
}

describe('Cognito identity provider', () => {
  it('maps password authentication to provider-neutral tokens and subject', async () => {
    const commands: unknown[] = []
    const client = {
      send: vi.fn(async (command: unknown) => {
        commands.push(command)
        return { AuthenticationResult: authenticationResult }
      }),
    }
    const identity = createCognitoIdentityProvider({
      clientId: 'client-id',
      clientSecret: 'client-secret',
      client,
    })

    const result = await identity.beginSignIn({
      username: 'operator@example.com',
      password: 'secret',
    })

    expect(isOk(result)).toBe(true)
    if (isOk(result))
      expect(result.value).toMatchObject({ kind: 'authenticated', subject: 'operator-subject' })
    const command = commands[0]
    if (!(command instanceof InitiateAuthCommand)) throw new Error('expected password command')
    expect(command.input.AuthParameters).toMatchObject({
      USERNAME: 'operator@example.com',
      PASSWORD: 'secret',
    })
    expect(command.input.AuthParameters?.SECRET_HASH).not.toBe('secret')
  })

  it('continues new-password and TOTP challenges through opaque continuations', async () => {
    const client = {
      send: vi
        .fn()
        .mockResolvedValueOnce({
          ChallengeName: 'NEW_PASSWORD_REQUIRED',
          Session: 'new-password-session',
        })
        .mockResolvedValueOnce({ AuthenticationResult: authenticationResult })
        .mockResolvedValueOnce({ ChallengeName: 'SOFTWARE_TOKEN_MFA', Session: 'mfa-session' })
        .mockResolvedValueOnce({ AuthenticationResult: authenticationResult }),
    }
    const identity = createCognitoIdentityProvider({ clientId: 'client-id', client })
    const passwordChallenge = await identity.beginSignIn({
      username: 'operator@example.com',
      password: 'temporary',
    })
    if (isOk(passwordChallenge) && passwordChallenge.value.kind === 'challenge') {
      expect(passwordChallenge.value.challenge.kind).toBe('new-password-required')
      await identity.answerNewPassword({
        continuation: passwordChallenge.value.challenge.continuation,
        newPassword: 'new',
      })
    }
    const totpChallenge = await identity.beginSignIn({
      username: 'operator@example.com',
      password: 'password',
    })
    if (isOk(totpChallenge) && totpChallenge.value.kind === 'challenge') {
      expect(totpChallenge.value.challenge.kind).toBe('software-token-mfa')
      await identity.answerTotpChallenge({
        continuation: totpChallenge.value.challenge.continuation,
        code: '123456',
      })
    }
    expect(client.send.mock.calls[1]?.[0]).toBeInstanceOf(RespondToAuthChallengeCommand)
    expect(client.send.mock.calls[3]?.[0]).toBeInstanceOf(RespondToAuthChallengeCommand)
  })

  it('associates, verifies, and completes TOTP setup', async () => {
    const client = {
      send: vi
        .fn()
        .mockResolvedValueOnce({ ChallengeName: 'MFA_SETUP', Session: 'setup-session' })
        .mockResolvedValueOnce({ SecretCode: 'totp-secret', Session: 'associated-session' })
        .mockResolvedValueOnce({ Status: 'SUCCESS', Session: 'verified-session' })
        .mockResolvedValueOnce({ AuthenticationResult: authenticationResult }),
    }
    const identity = createCognitoIdentityProvider({ clientId: 'client-id', client })
    const challenge = await identity.beginSignIn({
      username: 'operator@example.com',
      password: 'password',
    })
    if (!isOk(challenge) || challenge.value.kind !== 'challenge')
      throw new Error('expected setup challenge')
    const enrollment = await identity.associateTotp({
      continuation: challenge.value.challenge.continuation,
    })
    expect(isOk(enrollment) && enrollment.value.enrollment.secret).toBe('totp-secret')
    const verified = await identity.verifyTotpEnrollment({
      continuation: challenge.value.challenge.continuation,
      code: '123456',
    })

    expect(isOk(verified)).toBe(true)
    expect(client.send.mock.calls[1]?.[0]).toBeInstanceOf(AssociateSoftwareTokenCommand)
    expect(client.send.mock.calls[2]?.[0]).toBeInstanceOf(VerifySoftwareTokenCommand)
    expect(client.send.mock.calls[3]?.[0]).toBeInstanceOf(RespondToAuthChallengeCommand)
  })

  it('uses Cognito recovery, rotating refresh, and revoke commands', async () => {
    const commands: unknown[] = []
    const responses = [{}, {}, { AuthenticationResult: authenticationResult }, {}]
    const client = {
      send: vi.fn(async (command: unknown) => {
        commands.push(command)
        return responses.shift()
      }),
    }
    const identity = createCognitoIdentityProvider({ clientId: 'client-id', client })

    await identity.beginPasswordRecovery({ username: 'operator@example.com' })
    await identity.confirmPasswordRecovery({
      username: 'operator@example.com',
      code: '123456',
      newPassword: 'new',
    })
    const refreshed = await identity.refresh({ refreshToken: 'old-refresh-token' })
    await identity.revoke({ refreshToken: 'old-refresh-token' })

    expect(isOk(refreshed)).toBe(true)
    expect(commands.map(commandName)).toEqual([
      'ForgotPasswordCommand',
      'ConfirmForgotPasswordCommand',
      'GetTokensFromRefreshTokenCommand',
      'RevokeTokenCommand',
    ])
  })

  it('returns only safe typed errors for SDK failures', async () => {
    const sdkError = Object.assign(new Error('password and token leaked by provider'), {
      name: 'NotAuthorizedException',
    })
    const identity = createCognitoIdentityProvider({
      clientId: 'client-id',
      client: { send: vi.fn().mockRejectedValue(sdkError) },
    })

    const signIn = await identity.beginSignIn({
      username: 'operator@example.com',
      password: 'secret',
    })
    const refresh = await identity.refresh({ refreshToken: 'old-refresh-token' })

    expect(isErr(signIn) && signIn.error).toEqual({ kind: 'authentication-failed' })
    expect(isErr(refresh) && refresh.error).toEqual({ kind: 'refresh-failed', retryable: false })
    expect(JSON.stringify(signIn)).not.toContain('leaked')
  })
})

function commandName(command: unknown): string {
  if (command instanceof ForgotPasswordCommand) return 'ForgotPasswordCommand'
  if (command instanceof ConfirmForgotPasswordCommand) return 'ConfirmForgotPasswordCommand'
  if (command instanceof GetTokensFromRefreshTokenCommand) return 'GetTokensFromRefreshTokenCommand'
  if (command instanceof RevokeTokenCommand) return 'RevokeTokenCommand'
  throw new Error('unexpected Cognito command')
}
