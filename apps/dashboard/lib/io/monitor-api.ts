import 'server-only'

import { getUnixTime } from 'date-fns'
import { cookies } from 'next/headers'

import { type ApiResponse, isError, Status } from '@/lib/api-response'
import type {
  DashboardSession,
  DashboardSessionStore,
  IdentityProvider,
} from '@/lib/auth/contracts'
import { now } from '@/lib/clock'
import { ApiError, ApiErrorCode, fromEnvelope } from '@/lib/errors'
import {
  DASHBOARD_SESSION_COOKIE,
  createDynamoDashboardSessionStoreFromEnv,
} from '@/lib/io/auth/sessions'
import { err, ok, type Result } from '@/lib/result'

import { createCognitoIdentityProviderFromEnv } from './auth/cognito'
import { parseJson, tryCatch } from './server-action'

const ACCESS_TOKEN_REFRESH_SKEW_SECONDS = 60

type Fetch = (input: string, init?: RequestInit) => Promise<Response>
type SuccessfulApiResponse<T> = ApiResponse<T> & { readonly data: T }

export interface AuthenticatedMonitorApiClient {
  request<T>(path: string, init?: RequestInit): Promise<Result<T, ApiError>>
  requestResponse<T>(
    path: string,
    init?: RequestInit
  ): Promise<Result<SuccessfulApiResponse<T>, ApiError>>
}

export interface PublicHealthApiClient {
  get<T>(): Promise<Result<T, ApiError>>
}

export interface AuthenticatedMonitorApiClientOptions {
  readonly baseUrl: string
  readonly fetch?: Fetch
  readonly readSession: () => Promise<Result<DashboardSession | null, ApiError>>
  readonly sessionStore: Pick<DashboardSessionStore, 'refresh' | 'invalidate'>
  readonly identityProvider: Pick<IdentityProvider, 'refresh'>
}

/**
 * The sole server-side adapter for protected monitor API calls. It reads the
 * opaque dashboard session and forwards only its Cognito access token.
 */
export function createAuthenticatedMonitorApiClient(
  options: AuthenticatedMonitorApiClientOptions
): AuthenticatedMonitorApiClient {
  const fetchImpl = options.fetch ?? fetch

  return {
    async requestResponse<T>(
      path: string,
      init?: RequestInit
    ): Promise<Result<SuccessfulApiResponse<T>, ApiError>> {
      const session = await options.readSession()
      if (!session.ok || !session.value) return err(authenticationRequired())
      const sessionReference = session.value.reference

      const accessToken = await usableAccessToken(
        session.value,
        options.sessionStore,
        options.identityProvider
      )
      if (!accessToken.ok) return accessToken

      const headers = new Headers(init?.headers)
      headers.set('Content-Type', 'application/json')
      // Never forward caller-controlled authorization or browser credentials.
      headers.delete('Cookie')
      headers.set('Authorization', `Bearer ${accessToken.value}`)

      return requestResponse<T>(
        fetchImpl,
        options.baseUrl,
        path,
        { ...init, headers },
        async () => {
          await options.sessionStore.invalidate(sessionReference)
        }
      )
    },
    async request<T>(path: string, init?: RequestInit): Promise<Result<T, ApiError>> {
      const response = await this.requestResponse<T>(path, init)
      return response.ok ? ok(response.value.data as T) : response
    },
  }
}

/** Public health deliberately has no session lookup or Authorization header. */
export function createPublicHealthApiClient(options: {
  readonly baseUrl: string
  readonly fetch?: Fetch
}): PublicHealthApiClient {
  const fetchImpl = options.fetch ?? fetch
  return {
    async get<T>() {
      const response = await requestResponse<T>(fetchImpl, options.baseUrl, '/api/health', {
        cache: 'no-store',
        headers: { 'Content-Type': 'application/json' },
      })
      return response.ok ? ok(response.value.data as T) : response
    },
  }
}

/** Builds the authenticated adapter from server-only cookies and SST configuration. */
export function createAuthenticatedMonitorApiClientFromEnv(): AuthenticatedMonitorApiClient {
  const store = createDynamoDashboardSessionStoreFromEnv()
  return createAuthenticatedMonitorApiClient({
    baseUrl: apiBaseUrl(),
    readSession: async () => {
      const cookieStore = await cookies()
      const reference = cookieStore.get(DASHBOARD_SESSION_COOKIE.name)?.value
      if (!reference) return err(authenticationRequired())
      const session = await store.read(reference as DashboardSession['reference'])
      return session.ok ? ok(session.value) : err(authenticationRequired())
    },
    sessionStore: store,
    identityProvider: createCognitoIdentityProviderFromEnv(),
  })
}

export function createPublicHealthApiClientFromEnv(): PublicHealthApiClient {
  return createPublicHealthApiClient({ baseUrl: apiBaseUrl() })
}

async function usableAccessToken(
  session: DashboardSession,
  store: Pick<DashboardSessionStore, 'refresh'>,
  identityProvider: Pick<IdentityProvider, 'refresh'>
): Promise<Result<string, ApiError>> {
  if (session.tokens.accessTokenExpiresAt > getUnixTime(now()) + ACCESS_TOKEN_REFRESH_SKEW_SECONDS)
    return ok(session.tokens.accessToken)

  const refreshed = await store.refresh(session.reference, identityProvider)
  return refreshed.ok ? ok(refreshed.value.tokens.accessToken) : err(authenticationRequired())
}

async function requestResponse<T>(
  fetchImpl: Fetch,
  baseUrl: string,
  path: string,
  init: RequestInit,
  onAuthorizationDenied?: () => Promise<void>
): Promise<Result<SuccessfulApiResponse<T>, ApiError>> {
  const response = await tryCatch(
    () => fetchImpl(`${baseUrl.replace(/\/$/, '')}${path}`, { ...init, cache: 'no-store' }),
    () => internalError()
  )
  if (!response.ok) return response

  const body = await tryCatch(
    () => response.value.text(),
    () => internalError()
  )
  if (!body.ok) return body
  const parsed = body.value ? parseJson<ApiResponse<T>>(body.value) : err('empty response')
  // Gateway rejects invalid access credentials before Lambda, so its 401 has
  // no application envelope to parse.
  if (!parsed.ok)
    return response.value.status === 401 ? err(authenticationRequired()) : err(internalError())

  if (!response.value.ok) {
    if (isError(parsed.value)) {
      const error = attachMessage(
        fromEnvelope(parsed.value.reason, response.value.status),
        parsed.value.message
      )
      if (error.code === ApiErrorCode.AuthorizationDenied) {
        if (onAuthorizationDenied) await onAuthorizationDenied()
        return err(authenticationRequired())
      }
      return err(error)
    }
    if (response.value.status === 401) return err(authenticationRequired())
    return err(internalError(response.value.status))
  }
  if (parsed.value.status !== Status.Success || parsed.value.data === undefined)
    return err(internalError(response.value.status))
  return ok(parsed.value as SuccessfulApiResponse<T>)
}

function apiBaseUrl(): string {
  const baseUrl = process.env.NEXT_PUBLIC_MONITOR_API_BASE_URL
  return baseUrl ? baseUrl.replace(/\/$/, '') : ''
}

function authenticationRequired(): ApiError {
  return new ApiError(ApiErrorCode.AuthenticationRequired, 401)
}

function internalError(status = 500): ApiError {
  return new ApiError(ApiErrorCode.Internal, status)
}

function attachMessage(error: ApiError, message: string | undefined): ApiError {
  return message ? new ApiError(error.code, error.status, error.details, message) : error
}
