import assert from 'node:assert/strict'
import { mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import test from 'node:test'

import { rotateAuthKey } from './rotate-auth-key.mjs'

test('rotates a 256-bit key without printing it', () => {
  const configPath = join(mkdtempSync(join(tmpdir(), 'bolt-auth-key-')), 'target.json')
  writeFileSync(
    configPath,
    JSON.stringify({
      targets: [
        {
          stage: 'test',
          lifecycle: 'ephemeral',
          owner: 'test',
          service: 'bolt-monitor',
          accountId: '123456789012',
          region: 'us-east-1',
          credentialSource: 'test',
          dashboardOrigin: 'https://test.example.com',
          disposable: true,
          expiresAt: '2099-01-01T00:00:00Z',
        },
      ],
    })
  )
  let invocation
  const result = rotateAuthKey(
    { ...process.env, SST_STAGE: 'test', SST_TARGET_CONFIG: configPath },
    (command, args, options) => {
      invocation = { command, args, options }
    }
  )

  assert.deepEqual(result, { stage: 'test', parameterName: '/bolt-monitor/test/auth/aes-256-gcm' })
  assert.equal(invocation.command, 'aws')
  assert.equal(invocation.options.stdio, 'pipe')
  const value = invocation.args[invocation.args.indexOf('--value') + 1]
  assert.match(value, /^[A-Za-z0-9_-]{43}$/)
})
