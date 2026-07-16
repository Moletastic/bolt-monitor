import 'server-only'

import {
  DynamoDBClient,
  DeleteItemCommand,
  GetItemCommand,
  PutItemCommand,
  UpdateItemCommand,
  type AttributeValue,
} from '@aws-sdk/client-dynamodb'
import { getUnixTime } from 'date-fns'

import { now } from '@/lib/clock'
import type {
  AuthResult,
  DashboardSessionReference,
  DashboardSessionStore,
  NewDashboardSession,
  TokenBundle,
} from '@/lib/auth/contracts'
import { err, ok } from '@/lib/result'

import {
  type EncryptionKey,
  createOpaqueReference,
  createSsmKeyLoader,
  encryptionContext,
  encrypt,
  hashOpaqueReference,
  tryDecrypt,
} from './crypto'

const DASHBOARD_SESSION_LIFETIME_SECONDS = 12 * 60 * 60
const DASHBOARD_SESSION_KIND = 'SESSION'
const DASHBOARD_SESSION_SK = 'META'
const REFRESH_LEASE_SECONDS = 15
const REFRESH_REREAD_ATTEMPTS = 3

type DynamoClient = Pick<DynamoDBClient, 'send'>

export const DASHBOARD_SESSION_COOKIE = {
  name: '__Host-bolt-session',
  httpOnly: true,
  secure: true,
  sameSite: 'lax',
  path: '/',
} as const

/** Cookie attributes that remove a presented dashboard session from the browser. */
export const DASHBOARD_SESSION_EXPIRY_COOKIE = {
  ...DASHBOARD_SESSION_COOKIE,
  maxAge: 0,
} as const

export interface DashboardSessionCookieWriter {
  set(
    name: string,
    value: string,
    options: Omit<typeof DASHBOARD_SESSION_EXPIRY_COOKIE, 'name'>
  ): void
}

/** Expiry is sent even if the corresponding server record no longer exists. */
export function expireDashboardSessionCookie(cookies: DashboardSessionCookieWriter): void {
  const { name, ...options } = DASHBOARD_SESSION_EXPIRY_COOKIE
  cookies.set(name, '', options)
}

export interface DynamoDashboardSessionStoreOptions {
  readonly tableName: string
  readonly stage: string
  readonly application?: string
  readonly keyParameterName?: string
  readonly dynamo?: DynamoClient
  readonly ssm?: Pick<import('@aws-sdk/client-ssm').SSMClient, 'send'>
  /** Test-only injection for an already validated active key. */
  readonly loadKey?: () => Promise<EncryptionKey>
  /** Test-only bounded loser wait. */
  readonly waitForRefresh?: () => Promise<void>
  /** Test-only refresh lease owner factory. */
  readonly createRefreshOwner?: () => string
}

/** The only dashboard session lifetime accepted by the server-side storage boundary. */
export { DASHBOARD_SESSION_LIFETIME_SECONDS }

/**
 * Creates the authoritative AuthTable session store. This task intentionally
 * owns the conditional refresh protocol as well as session creation and lookup.
 */
export function createDynamoDashboardSessionStore(
  options: DynamoDashboardSessionStoreOptions
): DashboardSessionStore {
  const dynamo = options.dynamo ?? new DynamoDBClient({})
  const application = options.application ?? 'bolt-monitor'
  const loadKey = options.loadKey ?? createSsmKeyLoader(options.ssm, options.keyParameterName)
  const waitForRefresh =
    options.waitForRefresh ?? (() => new Promise((resolve) => setTimeout(resolve, 20)))
  const createRefreshOwner = options.createRefreshOwner ?? createOpaqueReference

  return {
    async create(session) {
      const timestamp = unixNow()
      if (!isValidNewSession(session, timestamp)) return err({ kind: 'session-invalid' })

      return inIoBoundary(async () => {
        const reference = createOpaqueReference() as DashboardSessionReference
        const recordHash = hashOpaqueReference(reference)
        const key = await loadKey()
        const encryptedTokens = encrypt(
          session.tokens,
          encryptionContext(
            application,
            options.stage,
            DASHBOARD_SESSION_KIND,
            key.generation,
            recordHash
          ),
          key.value
        )
        await dynamo.send(
          new PutItemCommand({
            TableName: options.tableName,
            Item: {
              PK: stringValue(`${DASHBOARD_SESSION_KIND}#${recordHash}`),
              SK: stringValue(DASHBOARD_SESSION_SK),
              Subject: stringValue(session.subject),
              EncryptedTokens: stringValue(encryptedTokens),
              ExpiresAt: numberValue(session.expiresAt),
              TTL: numberValue(session.expiresAt),
              KeyGeneration: stringValue(key.generation),
              Version: numberValue(1),
            },
            ConditionExpression: 'attribute_not_exists(PK)',
          })
        )
        return ok(reference)
      })
    },

    async read(reference) {
      return inIoBoundary(async () => {
        const recordHash = hashOpaqueReference(reference)
        const response = await dynamo.send(
          new GetItemCommand({
            TableName: options.tableName,
            Key: sessionKey(recordHash),
            ConsistentRead: true,
          })
        )
        if (!response.Item) return ok(null)
        const parsed = parseRecord(response.Item)
        if (!parsed || parsed.expiresAt <= unixNow()) return err({ kind: 'session-invalid' })

        const key = await loadKey()
        if (key.generation !== parsed.generation) return err({ kind: 'session-invalid' })
        const tokens = tryDecrypt(
          parsed.encryptedTokens,
          encryptionContext(
            application,
            options.stage,
            DASHBOARD_SESSION_KIND,
            key.generation,
            recordHash
          ),
          key.value
        )
        if (!isTokenBundle(tokens)) return err({ kind: 'session-invalid' })
        return ok({
          reference,
          subject: parsed.subject,
          tokens,
          expiresAt: parsed.expiresAt,
          version: parsed.version,
        })
      })
    },

    async refresh(reference, provider) {
      return inIoBoundary(async () => {
        const recordHash = hashOpaqueReference(reference)
        let observedVersion: number | null = null

        for (let attempt = 0; attempt < REFRESH_REREAD_ATTEMPTS; attempt += 1) {
          const current = await readSession(reference, recordHash)
          if (!current) return err({ kind: 'session-invalid' })
          if (observedVersion !== null && current.version > observedVersion) return ok(current)
          observedVersion = current.version

          const owner = createRefreshOwner()
          const leaseUntil = unixNow() + REFRESH_LEASE_SECONDS
          if (!(await acquireRefreshLease(recordHash, current.version, owner, leaseUntil))) {
            await waitForRefresh()
            continue
          }

          const refreshed = await provider.refresh({ refreshToken: current.tokens.refreshToken })
          if (!refreshed.ok) {
            if (refreshed.error.kind === 'refresh-failed' && !refreshed.error.retryable) {
              if (await invalidateTerminalRefresh(recordHash, current.version, owner))
                return err({ kind: 'session-invalid' })
              await waitForRefresh()
              continue
            }
            await releaseRefreshLease(recordHash, current.version, owner)
            return err(refreshed.error)
          }

          const key = await loadKey()
          if (key.generation !== current.generation) return err({ kind: 'session-invalid' })
          const encryptedTokens = encrypt(
            refreshed.value,
            encryptionContext(
              application,
              options.stage,
              DASHBOARD_SESSION_KIND,
              key.generation,
              recordHash
            ),
            key.value
          )
          if (await persistRotatedTokens(recordHash, current.version, owner, encryptedTokens))
            return ok({ ...current, tokens: refreshed.value, version: current.version + 1 })

          // The lease expired or another writer won. Never overwrite its token family.
          await waitForRefresh()
        }
        return err({ kind: 'refresh-failed', retryable: true })
      })
    },

    async replace(reference, session) {
      const invalidated = await this.invalidate(reference)
      if (!invalidated.ok) return invalidated
      return this.create(session)
    },

    async invalidate(reference) {
      return inIoBoundary(async () => {
        await dynamo.send(
          new DeleteItemCommand({
            TableName: options.tableName,
            Key: sessionKey(hashOpaqueReference(reference)),
          })
        )
        return ok(undefined)
      })
    },
  }

  async function readSession(
    reference: DashboardSessionReference,
    recordHash: string
  ): Promise<DashboardSessionWithGeneration | null> {
    const response = await dynamo.send(
      new GetItemCommand({
        TableName: options.tableName,
        Key: sessionKey(recordHash),
        ConsistentRead: true,
      })
    )
    if (!response.Item) return null
    const parsed = parseRecord(response.Item)
    if (!parsed || parsed.expiresAt <= unixNow()) return null
    const key = await loadKey()
    if (key.generation !== parsed.generation) return null
    const tokens = tryDecrypt(
      parsed.encryptedTokens,
      encryptionContext(
        application,
        options.stage,
        DASHBOARD_SESSION_KIND,
        key.generation,
        recordHash
      ),
      key.value
    )
    if (!isTokenBundle(tokens)) return null
    return {
      reference,
      subject: parsed.subject,
      tokens,
      expiresAt: parsed.expiresAt,
      version: parsed.version,
      generation: parsed.generation,
    }
  }

  async function acquireRefreshLease(
    recordHash: string,
    version: number,
    owner: string,
    leaseUntil: number
  ): Promise<boolean> {
    try {
      await dynamo.send(
        new UpdateItemCommand({
          TableName: options.tableName,
          Key: sessionKey(recordHash),
          UpdateExpression: 'SET RefreshOwner = :owner, RefreshLeaseUntil = :leaseUntil',
          ConditionExpression:
            'Version = :version AND (attribute_not_exists(RefreshLeaseUntil) OR RefreshLeaseUntil <= :now)',
          ExpressionAttributeValues: {
            ':owner': stringValue(owner),
            ':leaseUntil': numberValue(leaseUntil),
            ':version': numberValue(version),
            ':now': numberValue(unixNow()),
          },
        })
      )
      return true
    } catch (cause) {
      if (isConditionalFailure(cause)) return false
      throw cause
    }
  }

  async function persistRotatedTokens(
    recordHash: string,
    version: number,
    owner: string,
    encryptedTokens: string
  ): Promise<boolean> {
    try {
      await dynamo.send(
        new UpdateItemCommand({
          TableName: options.tableName,
          Key: sessionKey(recordHash),
          UpdateExpression:
            'SET EncryptedTokens = :tokens, Version = :nextVersion REMOVE RefreshOwner, RefreshLeaseUntil',
          ConditionExpression:
            'Version = :version AND RefreshOwner = :owner AND RefreshLeaseUntil > :now',
          ExpressionAttributeValues: {
            ':tokens': stringValue(encryptedTokens),
            ':nextVersion': numberValue(version + 1),
            ':version': numberValue(version),
            ':owner': stringValue(owner),
            ':now': numberValue(unixNow()),
          },
        })
      )
      return true
    } catch (cause) {
      if (isConditionalFailure(cause)) return false
      throw cause
    }
  }

  async function releaseRefreshLease(
    recordHash: string,
    version: number,
    owner: string
  ): Promise<void> {
    try {
      await dynamo.send(
        new UpdateItemCommand({
          TableName: options.tableName,
          Key: sessionKey(recordHash),
          UpdateExpression: 'REMOVE RefreshOwner, RefreshLeaseUntil',
          ConditionExpression: 'Version = :version AND RefreshOwner = :owner',
          ExpressionAttributeValues: {
            ':version': numberValue(version),
            ':owner': stringValue(owner),
          },
        })
      )
    } catch (cause) {
      if (!isConditionalFailure(cause)) throw cause
    }
  }

  async function invalidateTerminalRefresh(
    recordHash: string,
    version: number,
    owner: string
  ): Promise<boolean> {
    try {
      await dynamo.send(
        new DeleteItemCommand({
          TableName: options.tableName,
          Key: sessionKey(recordHash),
          ConditionExpression: 'Version = :version AND RefreshOwner = :owner',
          ExpressionAttributeValues: {
            ':version': numberValue(version),
            ':owner': stringValue(owner),
          },
        })
      )
      return true
    } catch (cause) {
      if (isConditionalFailure(cause)) return false
      throw cause
    }
  }
}

export function createDynamoDashboardSessionStoreFromEnv(): Pick<
  DashboardSessionStore,
  'create' | 'read' | 'replace' | 'invalidate'
> {
  return createDynamoDashboardSessionStore({
    tableName: process.env.AUTH_TABLE_NAME ?? '',
    stage: process.env.AUTH_STAGE ?? '',
    keyParameterName: process.env.AUTH_ENCRYPTION_KEY_PARAMETER_NAME,
  })
}

function isValidNewSession(session: NewDashboardSession, timestamp: number): boolean {
  return (
    session.subject.length > 0 &&
    isTokenBundle(session.tokens) &&
    session.expiresAt === timestamp + DASHBOARD_SESSION_LIFETIME_SECONDS
  )
}

function parseRecord(item: Record<string, AttributeValue>): ParsedRecord | null {
  const subject = item.Subject?.S
  const encryptedTokens = item.EncryptedTokens?.S
  const generation = item.KeyGeneration?.S
  const expiresAt = Number(item.ExpiresAt?.N)
  const version = Number(item.Version?.N)
  if (
    !subject ||
    !encryptedTokens ||
    !generation ||
    !Number.isInteger(expiresAt) ||
    !Number.isInteger(version) ||
    version < 1
  )
    return null
  return { subject, encryptedTokens, generation, expiresAt, version }
}

interface ParsedRecord {
  readonly subject: string
  readonly encryptedTokens: string
  readonly generation: string
  readonly expiresAt: number
  readonly version: number
}

interface DashboardSessionWithGeneration {
  readonly reference: DashboardSessionReference
  readonly subject: string
  readonly tokens: TokenBundle
  readonly expiresAt: number
  readonly version: number
  readonly generation: string
}

function isTokenBundle(value: unknown): value is TokenBundle {
  if (!value || typeof value !== 'object') return false
  const tokens = value as Record<string, unknown>
  return (
    typeof tokens.accessToken === 'string' &&
    tokens.accessToken.length > 0 &&
    typeof tokens.idToken === 'string' &&
    tokens.idToken.length > 0 &&
    typeof tokens.refreshToken === 'string' &&
    tokens.refreshToken.length > 0 &&
    Number.isInteger(tokens.accessTokenExpiresAt)
  )
}

function sessionKey(recordHash: string): Record<string, AttributeValue> {
  return {
    PK: stringValue(`${DASHBOARD_SESSION_KIND}#${recordHash}`),
    SK: stringValue(DASHBOARD_SESSION_SK),
  }
}

function stringValue(value: string): AttributeValue {
  return { S: value }
}

function numberValue(value: number): AttributeValue {
  return { N: String(value) }
}

function unixNow(): number {
  return getUnixTime(now())
}

async function inIoBoundary<T>(operation: () => Promise<AuthResult<T>>): Promise<AuthResult<T>> {
  try {
    return await operation()
  } catch {
    return err({ kind: 'storage-unavailable' })
  }
}

function isConditionalFailure(cause: unknown): boolean {
  return cause instanceof Error && cause.name === 'ConditionalCheckFailedException'
}
