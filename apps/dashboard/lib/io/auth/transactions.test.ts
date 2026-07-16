import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))
vi.mock('@/lib/clock', () => ({ now: () => new Date('2026-07-15T12:00:00Z') }))

import {
  GetItemCommand,
  PutItemCommand,
  UpdateItemCommand,
  type AttributeValue,
} from '@aws-sdk/client-dynamodb'

import { isErr, isOk } from '@/lib/result'

import {
  AUTH_TRANSACTION_LIFETIME_SECONDS,
  MAX_AUTH_TRANSACTION_ATTEMPTS,
  createDynamoAuthTransactionStore,
} from './transactions'

const timestamp = 1_784_116_800
const key = Buffer.alloc(32, 7)
const loadKey = async () => ({ generation: '42', value: key })

describe('Dynamo auth transaction store', () => {
  it('stores only a SHA-256 reference key, encrypted state, explicit expiry, and TTL cleanup', async () => {
    const client = { send: vi.fn().mockResolvedValue({}) }
    const store = createDynamoAuthTransactionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: client,
      loadKey,
    })

    const created = await store.create(draft())

    expect(isOk(created)).toBe(true)
    const command = putCommandFromCall(client.send.mock.calls[0]?.[0])
    const item = command.input.Item
    expect(item?.PK?.S).toMatch(/^AUTH_TX#[a-f0-9]{64}$/)
    expect(JSON.stringify(item)).not.toContain(isOk(created) ? created.value : '')
    expect(item?.EncryptedState?.S).not.toContain('cognito-challenge-session')
    expect(item?.ExpiresAt?.N).toBe(String(timestamp + AUTH_TRANSACTION_LIFETIME_SECONDS))
    expect(item?.TTL?.N).toBe(String(timestamp + AUTH_TRANSACTION_LIFETIME_SECONDS + 60))
    expect(item?.KeyGeneration?.S).toBe('42')
    expect(command.input.ConditionExpression).toBe('attribute_not_exists(PK)')
  })

  it('decrypts only with matching application, stage, record, and active generation context', async () => {
    const writeClient = { send: vi.fn().mockResolvedValue({}) }
    const writer = createDynamoAuthTransactionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: writeClient,
      loadKey,
    })
    const created = await writer.create(draft())
    if (!isOk(created)) throw new Error('expected created transaction')
    const persisted = putItemFromCall(writeClient.send.mock.calls[0]?.[0])
    const readerClient = { send: vi.fn().mockResolvedValue({ Item: persisted }) }
    const reader = createDynamoAuthTransactionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: readerClient,
      loadKey,
    })

    const read = await reader.read(created.value, 'sign-in')

    expect(isOk(read) && read.value?.providerState).toEqual({
      session: 'cognito-challenge-session',
    })
    const wrongStage = createDynamoAuthTransactionStore({
      tableName: 'auth-table',
      stage: 'production',
      dynamo: readerClient,
      loadKey,
    })
    expect(isErr(await wrongStage.read(created.value, 'sign-in'))).toBe(true)
    const wrongGeneration = createDynamoAuthTransactionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: readerClient,
      loadKey: async () => ({ generation: '43', value: key }),
    })
    expect(await wrongGeneration.read(created.value, 'sign-in')).toEqual({
      ok: false,
      error: { kind: 'transaction-invalid' },
    })
  })

  it('enforces flow, attempts, explicit expiry, and conditional single use', async () => {
    const client = { send: vi.fn().mockResolvedValue({}) }
    const store = createDynamoAuthTransactionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: client,
      loadKey,
    })

    expect(await store.create({ ...draft(), attempts: MAX_AUTH_TRANSACTION_ATTEMPTS })).toEqual({
      ok: false,
      error: { kind: 'transaction-invalid' },
    })
    const created = await store.create(draft())
    if (!isOk(created)) throw new Error('expected created transaction')
    const item = putItemFromCall(client.send.mock.calls[0]?.[0])
    client.send.mockResolvedValueOnce({ Item: item })
    expect(await store.read(created.value, 'password-recovery')).toEqual({
      ok: false,
      error: { kind: 'transaction-flow-mismatch' },
    })
    await store.consume(created.value, 'sign-in')
    const consume = updateCommandFromCall(client.send.mock.calls.at(-1)?.[0])
    expect(consume.input.ConditionExpression).toContain('attribute_not_exists(ConsumedAt)')
    expect(consume.input.ConditionExpression).toContain('ExpiresAt > :now')
    expect(consume.input.ConditionExpression).toContain('Flow = :flow')
  })

  it('does not accept expired records that TTL has not removed', async () => {
    const client = { send: vi.fn().mockResolvedValue({}) }
    const store = createDynamoAuthTransactionStore({
      tableName: 'auth-table',
      stage: 'staging',
      dynamo: client,
      loadKey,
    })
    const created = await store.create(draft())
    if (!isOk(created)) throw new Error('expected created transaction')
    const item = putItemFromCall(client.send.mock.calls[0]?.[0])
    item.ExpiresAt = { N: String(timestamp) }
    client.send.mockResolvedValueOnce({ Item: item })

    expect(await store.read(created.value, 'sign-in')).toEqual({
      ok: false,
      error: { kind: 'transaction-expired' },
    })
    expect(client.send.mock.calls.at(-1)?.[0]).toBeInstanceOf(GetItemCommand)
  })
})

function draft() {
  return {
    flow: 'sign-in' as const,
    challenge: 'software-token-mfa' as const,
    providerState: { session: 'cognito-challenge-session' },
    attempts: 0,
    expiresAt: timestamp + AUTH_TRANSACTION_LIFETIME_SECONDS,
  }
}

function putItemFromCall(value: unknown): Record<string, AttributeValue> {
  return putCommandFromCall(value).input.Item ?? fail('expected put item')
}

function putCommandFromCall(value: unknown): PutItemCommand {
  if (!(value instanceof PutItemCommand)) throw new Error('expected put item')
  return value
}

function updateCommandFromCall(value: unknown): UpdateItemCommand {
  if (!(value instanceof UpdateItemCommand)) throw new Error('expected conditional consume')
  return value
}

function fail(message: string): never {
  throw new Error(message)
}
