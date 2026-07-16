import { describe, expect, it, vi } from 'vitest'

vi.mock('server-only', () => ({}))

import { GetParameterCommand } from '@aws-sdk/client-ssm'

import { createSsmKeyLoader, encryptionContext, encrypt, tryDecrypt } from './crypto'

describe('auth encryption key loading', () => {
  it('loads only a non-empty 256-bit decrypted parameter with a generation', async () => {
    const client = {
      send: vi.fn().mockResolvedValue({
        Parameter: { Value: Buffer.alloc(32, 7).toString('base64'), Version: 42 },
      }),
    }

    await expect(createSsmKeyLoader(client, '/bolt/auth-key')()).resolves.toEqual({
      generation: '42',
      value: Buffer.alloc(32, 7),
    })
    expect(client.send).toHaveBeenCalledWith(
      expect.objectContaining({ input: { Name: '/bolt/auth-key', WithDecryption: true } })
    )
    expect(client.send.mock.calls[0]?.[0]).toBeInstanceOf(GetParameterCommand)
  })

  it.each([
    ['parameter name is absent', undefined],
    ['parameter is absent', { Parameter: {} }],
    [
      'key length is invalid',
      { Parameter: { Value: Buffer.alloc(31).toString('base64'), Version: 1 } },
    ],
  ])('fails closed when %s', async (_description, response) => {
    const client = { send: vi.fn().mockResolvedValue(response) }
    const loader = createSsmKeyLoader(client, response === undefined ? '' : '/bolt/auth-key')

    await expect(loader()).rejects.toThrow(/auth encryption key/)
  })

  it('requires the exact authenticated context and does not expose plaintext after tampering', () => {
    const key = Buffer.alloc(32, 7)
    const context = encryptionContext('bolt-monitor', 'staging', 'SESSION', '42', 'record-hash')
    const ciphertext = encrypt({ refreshToken: 'refresh-token' }, context, key)

    expect(tryDecrypt(ciphertext, context, key)).toEqual({ refreshToken: 'refresh-token' })
    expect(
      tryDecrypt(
        ciphertext,
        encryptionContext('bolt-monitor', 'staging', 'SESSION', '42', 'other-record-hash'),
        key
      )
    ).toBeUndefined()
    expect(tryDecrypt(`${ciphertext}tampered`, context, key)).toBeUndefined()
  })
})
