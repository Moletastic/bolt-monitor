import 'server-only'

import {
  DeleteItemCommand,
  DynamoDBClient,
  GetItemCommand,
  PutItemCommand,
  UpdateItemCommand,
  type AttributeValue,
} from '@aws-sdk/client-dynamodb'
import { getUnixTime } from 'date-fns'

import { now } from '@/lib/clock'
import type {
  AuthChallenge,
  AuthFlow,
  AuthResult,
  AuthTransactionDraft,
  AuthTransactionReference,
  AuthTransactionStore,
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

const AUTH_TRANSACTION_LIFETIME_SECONDS = 10 * 60
const AUTH_TRANSACTION_TTL_CLEANUP_SECONDS = 60
const MAX_AUTH_TRANSACTION_ATTEMPTS = 5
const AUTH_TRANSACTION_KIND = 'AUTH_TX'
const AUTH_TRANSACTION_SK = 'META'

type DynamoClient = Pick<DynamoDBClient, 'send'>
export interface DynamoAuthTransactionStoreOptions {
  readonly tableName: string
  readonly stage: string
  readonly application?: string
  readonly keyParameterName?: string
  readonly dynamo?: DynamoClient
  readonly ssm?: Pick<import('@aws-sdk/client-ssm').SSMClient, 'send'>
  /** Test-only injection for an already validated active key. */
  readonly loadKey?: () => Promise<EncryptionKey>
}

/** The only transaction lifetime accepted by the server-side storage boundary. */
export { AUTH_TRANSACTION_LIFETIME_SECONDS, MAX_AUTH_TRANSACTION_ATTEMPTS }

/**
 * Creates the authoritative AuthTable transaction store. Raw browser references
 * never enter DynamoDB; only their SHA-256 digests form primary keys.
 */
export function createDynamoAuthTransactionStore(
  options: DynamoAuthTransactionStoreOptions
): AuthTransactionStore {
  const dynamo = options.dynamo ?? new DynamoDBClient({})
  const application = options.application ?? 'bolt-monitor'
  const loadKey = options.loadKey ?? createSsmKeyLoader(options.ssm, options.keyParameterName)

  return {
    async create(draft) {
      const timestamp = unixNow()
      if (!isValidDraft(draft, timestamp)) return err({ kind: 'transaction-invalid' })

      return inIoBoundary(async () => {
        const reference = createOpaqueReference() as AuthTransactionReference
        const recordHash = hashOpaqueReference(reference)
        const key = await loadKey()
        const encryptedState = encrypt(
          draft.providerState,
          encryptionContext(
            application,
            options.stage,
            AUTH_TRANSACTION_KIND,
            key.generation,
            recordHash
          ),
          key.value
        )
        await dynamo.send(
          new PutItemCommand({
            TableName: options.tableName,
            Item: {
              PK: stringValue(`${AUTH_TRANSACTION_KIND}#${recordHash}`),
              SK: stringValue(AUTH_TRANSACTION_SK),
              EncryptedState: stringValue(encryptedState),
              Flow: stringValue(draft.flow),
              Challenge: stringValue(draft.challenge),
              Attempts: numberValue(draft.attempts),
              ExpiresAt: numberValue(draft.expiresAt),
              TTL: numberValue(draft.expiresAt + AUTH_TRANSACTION_TTL_CLEANUP_SECONDS),
              KeyGeneration: stringValue(key.generation),
              Version: numberValue(1),
            },
            ConditionExpression: 'attribute_not_exists(PK)',
          })
        )
        return ok(reference)
      })
    },

    async read(reference, flow) {
      return inIoBoundary(async () => {
        const recordHash = hashOpaqueReference(reference)
        const response = await dynamo.send(
          new GetItemCommand({
            TableName: options.tableName,
            Key: transactionKey(recordHash),
          })
        )
        if (!response.Item) return ok(null)
        const parsed = parseRecord(response.Item)
        if (!parsed) return err({ kind: 'transaction-invalid' })
        if (parsed.flow !== flow) return err({ kind: 'transaction-flow-mismatch' })
        if (parsed.consumed || parsed.expiresAt <= unixNow())
          return err({ kind: parsed.consumed ? 'transaction-consumed' : 'transaction-expired' })

        const key = await loadKey()
        if (key.generation !== parsed.generation) return err({ kind: 'transaction-invalid' })
        const providerState = tryDecrypt(
          parsed.encryptedState,
          encryptionContext(
            application,
            options.stage,
            AUTH_TRANSACTION_KIND,
            key.generation,
            recordHash
          ),
          key.value
        )
        if (providerState === undefined) return err({ kind: 'transaction-invalid' })
        return ok({
          reference,
          flow: parsed.flow,
          challenge: parsed.challenge,
          providerState,
          attempts: parsed.attempts,
          expiresAt: parsed.expiresAt,
        })
      })
    },

    async consume(reference, flow) {
      return inIoBoundary(async () => {
        const recordHash = hashOpaqueReference(reference)
        try {
          await dynamo.send(
            new UpdateItemCommand({
              TableName: options.tableName,
              Key: transactionKey(recordHash),
              UpdateExpression: 'SET ConsumedAt = :now',
              ConditionExpression:
                'attribute_exists(PK) AND attribute_not_exists(ConsumedAt) AND ExpiresAt > :now AND Flow = :flow',
              ExpressionAttributeValues: {
                ':now': numberValue(unixNow()),
                ':flow': stringValue(flow),
              },
            })
          )
          return ok(undefined)
        } catch (cause) {
          if (isConditionalFailure(cause)) return err({ kind: 'transaction-consumed' })
          return err({ kind: 'storage-unavailable' })
        }
      })
    },

    async invalidate(reference) {
      return inIoBoundary(async () => {
        await dynamo.send(
          new DeleteItemCommand({
            TableName: options.tableName,
            Key: transactionKey(hashOpaqueReference(reference)),
          })
        )
        return ok(undefined)
      })
    },
  }
}

export function createDynamoAuthTransactionStoreFromEnv(): AuthTransactionStore {
  return createDynamoAuthTransactionStore({
    tableName: process.env.AUTH_TABLE_NAME ?? '',
    stage: process.env.AUTH_STAGE ?? '',
    keyParameterName: process.env.AUTH_ENCRYPTION_KEY_PARAMETER_NAME,
  })
}

function isValidDraft(draft: AuthTransactionDraft, timestamp: number): boolean {
  return (
    isFlowChallengePair(draft.flow, draft.challenge) &&
    Number.isInteger(draft.attempts) &&
    draft.attempts >= 0 &&
    draft.attempts < MAX_AUTH_TRANSACTION_ATTEMPTS &&
    draft.expiresAt === timestamp + AUTH_TRANSACTION_LIFETIME_SECONDS
  )
}

function isFlowChallengePair(flow: AuthFlow, challenge: AuthChallenge['kind']): boolean {
  return flow === 'sign-in' || challenge === 'new-password-required'
}

function parseRecord(item: Record<string, AttributeValue>): ParsedRecord | null {
  const flow = item.Flow?.S
  const challenge = item.Challenge?.S
  const encryptedState = item.EncryptedState?.S
  const generation = item.KeyGeneration?.S
  const attempts = Number(item.Attempts?.N)
  const expiresAt = Number(item.ExpiresAt?.N)
  if (
    (flow !== 'sign-in' && flow !== 'password-recovery') ||
    (challenge !== 'new-password-required' &&
      challenge !== 'software-token-mfa' &&
      challenge !== 'software-token-setup') ||
    !encryptedState ||
    !generation ||
    !Number.isInteger(attempts) ||
    attempts < 0 ||
    attempts >= MAX_AUTH_TRANSACTION_ATTEMPTS ||
    !Number.isInteger(expiresAt)
  )
    return null
  return {
    flow,
    challenge,
    encryptedState,
    generation,
    attempts,
    expiresAt,
    consumed: Boolean(item.ConsumedAt?.N),
  }
}

interface ParsedRecord {
  readonly flow: AuthFlow
  readonly challenge: AuthChallenge['kind']
  readonly encryptedState: string
  readonly generation: string
  readonly attempts: number
  readonly expiresAt: number
  readonly consumed: boolean
}

function transactionKey(recordHash: string): Record<string, AttributeValue> {
  return {
    PK: stringValue(`${AUTH_TRANSACTION_KIND}#${recordHash}`),
    SK: stringValue(AUTH_TRANSACTION_SK),
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

function isConditionalFailure(cause: unknown): boolean {
  return cause instanceof Error && cause.name === 'ConditionalCheckFailedException'
}

async function inIoBoundary<T>(operation: () => Promise<AuthResult<T>>): Promise<AuthResult<T>> {
  try {
    return await operation()
  } catch {
    return err({ kind: 'storage-unavailable' })
  }
}
