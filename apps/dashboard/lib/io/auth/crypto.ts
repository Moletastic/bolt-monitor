import 'server-only'

import { GetParameterCommand, SSMClient } from '@aws-sdk/client-ssm'
import { createCipheriv, createDecipheriv, createHash, randomBytes } from 'node:crypto'

type ParameterClient = Pick<SSMClient, 'send'>

export interface EncryptionKey {
  readonly generation: string
  readonly value: Buffer
}

export function createSsmKeyLoader(
  client?: ParameterClient,
  parameterName = ''
): () => Promise<EncryptionKey> {
  const ssm = client ?? new SSMClient({})
  return async () => {
    if (!parameterName) throw new Error('missing auth encryption key parameter')
    const response = await ssm.send(
      new GetParameterCommand({ Name: parameterName, WithDecryption: true })
    )
    const value = response.Parameter?.Value
    const version = response.Parameter?.Version
    if (!value || !version) throw new Error('missing auth encryption key')
    const key = Buffer.from(value, 'base64')
    if (key.length !== 32) throw new Error('invalid auth encryption key')
    return { generation: String(version), value: key }
  }
}

export function encrypt(value: unknown, aad: string, key: Buffer): string {
  const plaintext = JSON.stringify(value)
  if (plaintext === undefined) throw new Error('unserializable encrypted value')
  const iv = randomBytes(12)
  const cipher = createCipheriv('aes-256-gcm', key, iv)
  cipher.setAAD(Buffer.from(aad))
  const ciphertext = Buffer.concat([cipher.update(plaintext, 'utf8'), cipher.final()])
  return `${iv.toString('base64url')}.${ciphertext.toString('base64url')}.${cipher.getAuthTag().toString('base64url')}`
}

export function tryDecrypt(value: string, aad: string, key: Buffer): unknown | undefined {
  try {
    const [iv, ciphertext, tag, ...rest] = value.split('.')
    if (!iv || !ciphertext || !tag || rest.length > 0) return undefined
    const decipher = createDecipheriv('aes-256-gcm', key, Buffer.from(iv, 'base64url'))
    decipher.setAAD(Buffer.from(aad))
    decipher.setAuthTag(Buffer.from(tag, 'base64url'))
    return JSON.parse(
      Buffer.concat([
        decipher.update(Buffer.from(ciphertext, 'base64url')),
        decipher.final(),
      ]).toString('utf8')
    )
  } catch {
    return undefined
  }
}

export function encryptionContext(
  application: string,
  stage: string,
  recordKind: string,
  generation: string,
  recordHash: string
): string {
  return `${application}|${stage}|${recordKind}|${generation}|${recordHash}`
}

export function hashOpaqueReference(reference: string): string {
  return createHash('sha256').update(reference).digest('hex')
}

export function createOpaqueReference(): string {
  return randomBytes(32).toString('base64url')
}
