import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))
vi.mock('@/lib/clock', () => ({ now: () => new Date('2026-07-15T12:00:00Z') }))

import {
  DeleteItemCommand,
  GetItemCommand,
  PutItemCommand,
  UpdateItemCommand,
  type AttributeValue,
} from '@aws-sdk/client-dynamodb'

import { isErr, isOk } from '@/lib/result'

import {
  DASHBOARD_SESSION_COOKIE,
  DASHBOARD_SESSION_EXPIRY_COOKIE,
  DASHBOARD_SESSION_LIFETIME_SECONDS,
  createDynamoDashboardSessionStore,
  expireDashboardSessionCookie,
} from './sessions'

const timestamp = 1_784_116_800
const key = Buffer.alloc(32, 7)
const loadKey = async () => ({ generation: '42', value: key })

describe('Dynamo dashboard session store', () => {
  it('creates a 256-bit opaque session ID and persists only its hashed lookup key', async () => {
    const client = { send: vi.fn().mockResolvedValue({}) }
    const store = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: client,
      loadKey,
    })

    const created = await store.create(session())

    expect(isOk(created)).toBe(true)
    if (!isOk(created)) throw new Error('expected created session')
    expect(created.value).toMatch(/^[A-Za-z0-9_-]{43}$/)
    const command = putCommandFromCall(client.send.mock.calls[0]?.[0])
    const item = command.input.Item
    expect(item?.PK?.S).toMatch(/^SESSION#[a-f0-9]{64}$/)
    expect(JSON.stringify(item)).not.toContain(created.value)
    expect(item?.EncryptedTokens?.S).not.toContain('access-token')
    expect(item?.EncryptedTokens?.S).not.toContain('refresh-token')
    expect(item?.ExpiresAt?.N).toBe(String(timestamp + DASHBOARD_SESSION_LIFETIME_SECONDS))
    expect(item?.TTL?.N).toBe(String(timestamp + DASHBOARD_SESSION_LIFETIME_SECONDS))
    expect(item?.KeyGeneration?.S).toBe('42')
    expect(item?.Version?.N).toBe('1')
    expect(command.input.ConditionExpression).toBe('attribute_not_exists(PK)')
  })

  it('defines a compliant host-only HttpOnly cookie policy', () => {
    expect(DASHBOARD_SESSION_COOKIE).toEqual({
      name: '__Host-bolt-session',
      httpOnly: true,
      secure: true,
      sameSite: 'lax',
      path: '/',
    })
    expect(DASHBOARD_SESSION_COOKIE).not.toHaveProperty('domain')
  })

  it('expires the host-only cookie even when no server session exists', () => {
    const cookies = { set: vi.fn() }

    expireDashboardSessionCookie(cookies)

    expect(DASHBOARD_SESSION_EXPIRY_COOKIE).toEqual({
      ...DASHBOARD_SESSION_COOKIE,
      maxAge: 0,
    })
    expect(cookies.set).toHaveBeenCalledWith('__Host-bolt-session', '', {
      httpOnly: true,
      secure: true,
      sameSite: 'lax',
      path: '/',
      maxAge: 0,
    })
  })

  it('decrypts token bundles only for the matching application, stage, record, and active key generation', async () => {
    const writeClient = { send: vi.fn().mockResolvedValue({}) }
    const writer = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: writeClient,
      loadKey,
    })
    const created = await writer.create(session())
    if (!isOk(created)) throw new Error('expected created session')
    const persisted = putItemFromCall(writeClient.send.mock.calls[0]?.[0])
    const readerClient = { send: vi.fn().mockResolvedValue({ Item: persisted }) }
    const reader = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: readerClient,
      loadKey,
    })

    expect(await reader.read(created.value)).toEqual({
      ok: true,
      value: {
        reference: created.value,
        subject: 'cognito-subject',
        tokens: session().tokens,
        expiresAt: timestamp + DASHBOARD_SESSION_LIFETIME_SECONDS,
        version: 1,
      },
    })
    const wrongStage = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'production',
      dynamo: readerClient,
      loadKey,
    })
    expect(isErr(await wrongStage.read(created.value))).toBe(true)
    const wrongGeneration = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: readerClient,
      loadKey: async () => ({ generation: '43', value: key }),
    })
    expect(await wrongGeneration.read(created.value)).toEqual({
      ok: false,
      error: { kind: 'session-invalid' },
    })
  })

  it('rejects records at their explicit expiry even when DynamoDB TTL has not removed them', async () => {
    const client = { send: vi.fn().mockResolvedValue({}) }
    const store = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: client,
      loadKey,
    })
    const created = await store.create(session())
    if (!isOk(created)) throw new Error('expected created session')
    const item = putItemFromCall(client.send.mock.calls[0]?.[0])
    item.ExpiresAt = { N: String(timestamp) }
    item.TTL = { N: String(timestamp + 60) }
    client.send.mockResolvedValueOnce({ Item: item })

    expect(await store.read(created.value)).toEqual({
      ok: false,
      error: { kind: 'session-invalid' },
    })
    expect(client.send.mock.calls.at(-1)?.[0]).toBeInstanceOf(GetItemCommand)
  })

  it('makes one lease winner rotate and persist the token bundle at the next version', async () => {
    const { reference, item } = await persistedSession()
    const client = { send: vi.fn().mockResolvedValueOnce({ Item: item }).mockResolvedValue({}) }
    const store = refreshStore(client)
    const provider = { refresh: vi.fn().mockResolvedValue({ ok: true, value: refreshedTokens() }) }

    expect(await store.refresh(reference, provider)).toMatchObject({
      ok: true,
      value: { tokens: refreshedTokens(), version: 2 },
    })
    expect(provider.refresh).toHaveBeenCalledWith({ refreshToken: 'refresh-token' })
    const acquire = updateCommandFromCall(client.send.mock.calls[1]?.[0])
    const persist = updateCommandFromCall(client.send.mock.calls[2]?.[0])
    expect(acquire).toBeInstanceOf(UpdateItemCommand)
    expect(persist).toBeInstanceOf(UpdateItemCommand)
    expect(persist.input.ConditionExpression).toContain('Version = :version')
    expect(persist.input.ConditionExpression).toContain('RefreshOwner = :owner')
    expect(persist.input.UpdateExpression).toContain('Version = :nextVersion')
    expect(persist.input.UpdateExpression).toContain('REMOVE RefreshOwner, RefreshLeaseUntil')
    expect(JSON.stringify(persist.input)).not.toContain('rotated-refresh-token')
  })

  it('bounds loser rereads and accepts the newer winning version without refreshing', async () => {
    const { reference, item } = await persistedSession()
    const winningItem = structuredClone(item)
    winningItem.Version = { N: '2' }
    const conditionalFailure = Object.assign(new Error('conditional'), {
      name: 'ConditionalCheckFailedException',
    })
    const client = {
      send: vi
        .fn()
        .mockResolvedValueOnce({ Item: item })
        .mockRejectedValueOnce(conditionalFailure)
        .mockResolvedValueOnce({ Item: winningItem }),
    }
    const provider = { refresh: vi.fn() }
    const store = refreshStore(client, { waitForRefresh: vi.fn().mockResolvedValue(undefined) })

    expect(await store.refresh(reference, provider)).toMatchObject({
      ok: true,
      value: { version: 2 },
    })
    expect(provider.refresh).not.toHaveBeenCalled()
    expect(client.send.mock.calls[2]?.[0]).toBeInstanceOf(GetItemCommand)
    expect((client.send.mock.calls[2]?.[0] as GetItemCommand).input.ConsistentRead).toBe(true)
  })

  it('allows takeover after an expired lease and rejects a stale writer conditionally', async () => {
    const { reference, item } = await persistedSession()
    item.RefreshOwner = { S: 'abandoned-owner' }
    item.RefreshLeaseUntil = { N: String(timestamp - 1) }
    const staleWrite = Object.assign(new Error('conditional'), {
      name: 'ConditionalCheckFailedException',
    })
    const client = {
      send: vi
        .fn()
        .mockResolvedValueOnce({ Item: item })
        .mockResolvedValueOnce({})
        .mockRejectedValueOnce(staleWrite)
        .mockResolvedValueOnce({ Item: { ...item, Version: { N: '2' } } }),
    }
    const store = refreshStore(client, { waitForRefresh: vi.fn().mockResolvedValue(undefined) })

    expect(
      await store.refresh(reference, {
        refresh: vi.fn().mockResolvedValue({ ok: true, value: refreshedTokens() }),
      })
    ).toMatchObject({
      ok: true,
      value: { version: 2 },
    })
    const acquire = updateCommandFromCall(client.send.mock.calls[1]?.[0])
    expect(acquire.input.ConditionExpression).toContain('RefreshLeaseUntil <= :now')
    const stalePersist = updateCommandFromCall(client.send.mock.calls[2]?.[0])
    expect(stalePersist.input.ConditionExpression).toContain('RefreshLeaseUntil > :now')
  })

  it('deletes the session only when the lease owner receives a terminal refresh failure', async () => {
    const { reference, item } = await persistedSession()
    const client = { send: vi.fn().mockResolvedValueOnce({ Item: item }).mockResolvedValue({}) }
    const store = refreshStore(client)

    expect(
      await store.refresh(reference, {
        refresh: vi
          .fn()
          .mockResolvedValue({ ok: false, error: { kind: 'refresh-failed', retryable: false } }),
      })
    ).toEqual({ ok: false, error: { kind: 'session-invalid' } })
    const deletion = deleteCommandFromCall(client.send.mock.calls.at(-1)?.[0])
    expect(deletion).toBeInstanceOf(DeleteItemCommand)
    expect(deletion.input.ConditionExpression).toBe('Version = :version AND RefreshOwner = :owner')
  })

  it('invalidates a missing session idempotently', async () => {
    const client = { send: vi.fn().mockResolvedValue({}) }
    const store = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: client,
      loadKey,
    })
    const reference = 'prior-session' as never

    await expect(store.invalidate(reference)).resolves.toEqual({ ok: true, value: undefined })
    await expect(store.invalidate(reference)).resolves.toEqual({ ok: true, value: undefined })

    for (const [command] of client.send.mock.calls) {
      expect(command).toBeInstanceOf(DeleteItemCommand)
      expect((command as DeleteItemCommand).input.ConditionExpression).toBeUndefined()
    }
  })

  it('replaces an old session only after invalidating its server record', async () => {
    const client = { send: vi.fn().mockResolvedValue({}) }
    const store = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: client,
      loadKey,
    })

    const replaced = await store.replace('prior-session' as never, session())

    expect(isOk(replaced)).toBe(true)
    expect(client.send.mock.calls[0]?.[0]).toBeInstanceOf(DeleteItemCommand)
    expect(client.send.mock.calls[1]?.[0]).toBeInstanceOf(PutItemCommand)
  })
})

function session() {
  return {
    subject: 'cognito-subject',
    tokens: {
      accessToken: 'access-token',
      idToken: 'id-token',
      refreshToken: 'refresh-token',
      accessTokenExpiresAt: timestamp + 300,
    },
    expiresAt: timestamp + DASHBOARD_SESSION_LIFETIME_SECONDS,
  }
}

function putItemFromCall(value: unknown): Record<string, AttributeValue> {
  return putCommandFromCall(value).input.Item ?? fail('expected put item')
}

function putCommandFromCall(value: unknown): PutItemCommand {
  if (!(value instanceof PutItemCommand)) throw new Error('expected put item')
  return value
}

function fail(message: string): never {
  throw new Error(message)
}

function updateCommandFromCall(value: unknown): UpdateItemCommand {
  if (!(value instanceof UpdateItemCommand)) throw new Error('expected update item')
  return value
}

function deleteCommandFromCall(value: unknown): DeleteItemCommand {
  if (!(value instanceof DeleteItemCommand)) throw new Error('expected delete item')
  return value
}

async function persistedSession() {
  const writerClient = { send: vi.fn().mockResolvedValue({}) }
  const writer = createDynamoDashboardSessionStore({
    tableName: 'auth-table',
    stage: 'staging',
    dynamo: writerClient,
    loadKey,
  })
  const created = await writer.create(session())
  if (!isOk(created)) throw new Error('expected created session')
  return { reference: created.value, item: putItemFromCall(writerClient.send.mock.calls[0]?.[0]) }
}

function refreshStore(
  client: { send: ReturnType<typeof vi.fn> },
  overrides: { readonly waitForRefresh?: () => Promise<void> } = {}
) {
  return createDynamoDashboardSessionStore({
    tableName: 'auth-table',
    stage: 'staging',
    dynamo: client as never,
    loadKey,
    waitForRefresh: overrides.waitForRefresh,
    createRefreshOwner: () => 'refresh-owner',
  })
}

function refreshedTokens() {
  return {
    accessToken: 'rotated-access-token',
    idToken: 'rotated-id-token',
    refreshToken: 'rotated-refresh-token',
    accessTokenExpiresAt: timestamp + 600,
  }
}
