import { describe, expect, it, vi } from 'vitest'

const clock = vi.hoisted(() => ({ now: new Date('2026-07-15T12:00:00Z') }))

vi.mock('server-only', () => ({}))
vi.mock('@/lib/clock', () => ({ now: () => clock.now }))

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

  it('accepts a session with a normal remaining duration after authentication advances time', async () => {
    const client = { send: vi.fn().mockResolvedValue({}) }
    const store = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: client,
      loadKey,
    })
    clock.now = new Date('2026-07-15T12:00:01Z')

    try {
      const created = await store.create(session())

      expect(isOk(created)).toBe(true)
      expect(client.send).toHaveBeenCalledWith(expect.any(PutItemCommand))
    } finally {
      clock.now = new Date('2026-07-15T12:00:00Z')
    }
  })

  it('rejects expired and oversized session expiries', async () => {
    const client = { send: vi.fn().mockResolvedValue({}) }
    const store = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: client,
      loadKey,
    })

    await expect(store.create({ ...session(), expiresAt: timestamp })).resolves.toEqual({
      ok: false,
      error: { kind: 'session-invalid' },
    })
    await expect(
      store.create({
        ...session(),
        expiresAt: timestamp + DASHBOARD_SESSION_LIFETIME_SECONDS + 1,
      })
    ).resolves.toEqual({ ok: false, error: { kind: 'session-invalid' } })
    expect(client.send).not.toHaveBeenCalled()
  })

  it('issues distinct 256-bit cookie values without persisting raw identifiers or token material', async () => {
    const client = { send: vi.fn().mockResolvedValue({}) }
    const store = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: client,
      loadKey,
    })

    const first = await store.create(session())
    const second = await store.create(session())

    expect(isOk(first) && isOk(second)).toBe(true)
    if (!isOk(first) || !isOk(second)) throw new Error('expected created sessions')
    expect(first.value).not.toBe(second.value)
    expect(Buffer.from(first.value, 'base64url')).toHaveLength(32)
    expect(Buffer.from(second.value, 'base64url')).toHaveLength(32)
    for (const [command] of client.send.mock.calls) {
      const item = putItemFromCall(command)
      expect(JSON.stringify(item)).not.toContain(first.value)
      expect(JSON.stringify(item)).not.toContain(second.value)
      expect(JSON.stringify(item)).not.toContain(session().tokens.accessToken)
      expect(JSON.stringify(item)).not.toContain(session().tokens.idToken)
      expect(JSON.stringify(item)).not.toContain(session().tokens.refreshToken)
    }
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

  it('rejects ciphertext copied to a different session record without re-encrypting it', async () => {
    const { reference, item } = await persistedSession()
    const differentReference = 'different-session-reference' as never
    const client = { send: vi.fn().mockResolvedValue({ Item: item }) }
    const store = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: client,
      loadKey,
    })

    await expect(store.read(differentReference)).resolves.toEqual({
      ok: false,
      error: { kind: 'session-invalid' },
    })
    expect(client.send).toHaveBeenCalledWith(expect.any(GetItemCommand))
    expect(client.send).not.toHaveBeenCalledWith(expect.any(UpdateItemCommand))
    expect(reference).not.toBe(differentReference)
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

  it('coordinates concurrent refreshes so one winner persists tokens and a waiting reader uses them', async () => {
    const { reference, item } = await persistedSession()
    const client = concurrentRefreshClient(item)
    const providerStarted = deferred<void>()
    const finishProvider = deferred<void>()
    const winnerProvider = {
      refresh: vi.fn().mockImplementation(async () => {
        providerStarted.resolve()
        await finishProvider.promise
        return { ok: true as const, value: refreshedTokens() }
      }),
    }
    const readerProvider = { refresh: vi.fn() }
    const readerWait = deferred<void>()
    const winner = refreshStore(client).refresh(reference, winnerProvider)

    await providerStarted.promise
    const reader = refreshStore(client, { waitForRefresh: () => readerWait.promise }).refresh(
      reference,
      readerProvider
    )
    await vi.waitFor(() => expect(client.leaseFailures).toBe(1))

    finishProvider.resolve()
    await winner
    readerWait.resolve()

    const readerResult = await reader
    expect(isOk(readerResult)).toBe(true)
    if (!isOk(readerResult)) throw new Error('expected reader to use the winning refresh')
    expect(readerResult.value.tokens).toEqual(refreshedTokens())
    expect(readerResult.value.version).toBe(2)
    expect(winnerProvider.refresh).toHaveBeenCalledTimes(1)
    expect(readerProvider.refresh).not.toHaveBeenCalled()
    expect(client.item.Version?.N).toBe('2')
    expect(client.item.RefreshOwner).toBeUndefined()
    expect(client.item.RefreshLeaseUntil).toBeUndefined()
  })

  it('persists rotated tokens so a later reader decrypts the winning bundle', async () => {
    const { reference, item } = await persistedSession()
    const client = { send: vi.fn().mockResolvedValueOnce({ Item: item }).mockResolvedValue({}) }
    const store = refreshStore(client)

    await expect(
      store.refresh(reference, {
        refresh: vi.fn().mockResolvedValue({ ok: true, value: refreshedTokens() }),
      })
    ).resolves.toMatchObject({ ok: true, value: { version: 2 } })

    const persisted = structuredClone(item)
    const update = updateCommandFromCall(client.send.mock.calls[2]?.[0])
    const encryptedTokens = update.input.ExpressionAttributeValues?.[':tokens']
    if (!encryptedTokens) throw new Error('expected rotated token ciphertext')
    persisted.EncryptedTokens = encryptedTokens
    persisted.Version = { N: '2' }
    const reader = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: { send: vi.fn().mockResolvedValue({ Item: persisted }) },
      loadKey,
    })

    const read = await reader.read(reference)
    expect(isOk(read)).toBe(true)
    if (!isOk(read)) throw new Error('expected persisted rotated tokens')
    if (!read.value) throw new Error('expected persisted rotated session')
    expect(read.value.tokens).toEqual(refreshedTokens())
    expect(read.value.version).toBe(2)
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

  it('releases the lease and fails only the current request after a transient refresh failure', async () => {
    const { reference, item } = await persistedSession()
    const client = { send: vi.fn().mockResolvedValueOnce({ Item: item }).mockResolvedValue({}) }
    const store = refreshStore(client)

    await expect(
      store.refresh(reference, {
        refresh: vi
          .fn()
          .mockResolvedValue({ ok: false, error: { kind: 'refresh-failed', retryable: true } }),
      })
    ).resolves.toEqual({ ok: false, error: { kind: 'refresh-failed', retryable: true } })

    const release = updateCommandFromCall(client.send.mock.calls.at(-1)?.[0])
    expect(release.input.UpdateExpression).toBe('REMOVE RefreshOwner, RefreshLeaseUntil')
    expect(release.input.ConditionExpression).toBe('Version = :version AND RefreshOwner = :owner')
    expect(client.send).not.toHaveBeenCalledWith(expect.any(DeleteItemCommand))
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

  it('fails closed without leaking tokens when key loading or storage fails', async () => {
    const keyFailure = new Error('key value must not escape')
    const keyStore = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: { send: vi.fn() },
      loadKey: async () => Promise.reject(keyFailure),
    })
    const storageStore = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: { send: vi.fn().mockRejectedValue(new Error('dynamodb unavailable')) },
      loadKey,
    })

    await expect(keyStore.create(session())).resolves.toEqual({
      ok: false,
      error: { kind: 'storage-unavailable' },
    })
    await expect(storageStore.create(session())).resolves.toEqual({
      ok: false,
      error: { kind: 'storage-unavailable' },
    })
  })

  it('rejects every prior-generation session using only the active key and no write fallback', async () => {
    const first = await persistedSession()
    const second = await persistedSession()
    const client = {
      send: vi
        .fn()
        .mockResolvedValueOnce({ Item: first.item })
        .mockResolvedValueOnce({ Item: second.item }),
    }
    const activeGenerationStore = createDynamoDashboardSessionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: client,
      loadKey: async () => ({ generation: '43', value: Buffer.alloc(32, 8) }),
    })

    await expect(activeGenerationStore.read(first.reference)).resolves.toEqual({
      ok: false,
      error: { kind: 'session-invalid' },
    })
    await expect(activeGenerationStore.read(second.reference)).resolves.toEqual({
      ok: false,
      error: { kind: 'session-invalid' },
    })
    expect(client.send).toHaveBeenCalledTimes(2)
    for (const [command] of client.send.mock.calls) {
      expect(command).toBeInstanceOf(GetItemCommand)
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

function deferred<T>() {
  let resolve: (value: T) => void = () => undefined
  const promise = new Promise<T>((complete) => {
    resolve = complete
  })
  return { promise, resolve }
}

function concurrentRefreshClient(initialItem: Record<string, AttributeValue>) {
  const item = structuredClone(initialItem)
  let leaseFailures = 0

  return {
    get item() {
      return item
    },
    get leaseFailures() {
      return leaseFailures
    },
    send: vi.fn().mockImplementation(async (command: unknown) => {
      if (command instanceof GetItemCommand) return { Item: structuredClone(item) }
      if (!(command instanceof UpdateItemCommand)) throw new Error('expected session update')

      const values = command.input.ExpressionAttributeValues
      const expression = command.input.UpdateExpression
      if (!values) throw new Error('expected update values')
      if (!expression) throw new Error('expected update expression')
      if (expression.startsWith('SET RefreshOwner')) {
        const leaseUntil = Number(values[':leaseUntil']?.N)
        const now = Number(values[':now']?.N)
        if (item.RefreshLeaseUntil && Number(item.RefreshLeaseUntil.N) > now) {
          leaseFailures += 1
          throw conditionalFailure()
        }
        item.RefreshOwner = values[':owner']
        item.RefreshLeaseUntil = { N: String(leaseUntil) }
        return {}
      }
      if (expression.startsWith('SET EncryptedTokens')) {
        item.EncryptedTokens = values[':tokens']
        item.Version = values[':nextVersion']
        delete item.RefreshOwner
        delete item.RefreshLeaseUntil
        return {}
      }
      throw new Error('unexpected session update')
    }),
  }
}

function conditionalFailure(): Error {
  return Object.assign(new Error('conditional'), { name: 'ConditionalCheckFailedException' })
}
