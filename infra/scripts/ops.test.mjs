import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import test from 'node:test'

const persistent = {
  stage: 'staging',
  profile: 'bolt-monitor',
  lifecycle: 'persistent',
  owner: 'platform',
  service: 'bolt-monitor',
  accountId: '123456789012',
  region: 'us-east-1',
  dashboardOrigin: 'https://staging.example.com',
  approved: true,
}

async function withTarget(name, target, run) {
  const dir = mkdtempSync(join(tmpdir(), 'bolt-ops-'))
  const targetsDir = join(dir, 'infra', 'targets')
  mkdirSync(targetsDir, { recursive: true })
  const path = join(targetsDir, `${name}.target.json`)
  writeFileSync(path, JSON.stringify(target))
  const previousTarget = process.env.TARGET
  const previousFile = process.env.TARGET_FILE
  process.env.TARGET = name
  process.env.TARGET_FILE = path
  try {
    await run()
  } finally {
    if (previousTarget === undefined) delete process.env.TARGET
    else process.env.TARGET = previousTarget
    if (previousFile === undefined) delete process.env.TARGET_FILE
    else process.env.TARGET_FILE = previousFile
  }
}

test('status rejects account mismatch before mutation', async () => {
  await withTarget('staging', persistent, async () => {
    const ops = await import('./ops.mjs')
    await assert.rejects(
      () =>
        ops.status({
          run: () => {
            throw new Error('account mismatch')
          },
        }),
      /account mismatch/
    )
  })
})

test('persistent remove refuses without DESTROY=yes', async () => {
  await withTarget('staging', persistent, async () => {
    const ops = await import('./ops.mjs')
    await assert.rejects(() => ops.remove({ destroy: false }, { run: () => '' }), /DESTROY=yes/)
  })
})

test('deploy postflight fails when health endpoint unreachable', async () => {
  await withTarget('staging', persistent, async () => {
    const ops = await import('./ops.mjs')
    await assert.rejects(
      () =>
        ops.deploy({
          preflight: () => ({ accountId: '123456789012', region: 'us-east-1' }),
          run: (command, args) => {
            if (command === 'aws' && args[0] === 'dynamodb') {
              return JSON.stringify({ Table: { DeletionProtectionEnabled: true } })
            }
            if (command === 'curl') {
              throw new Error('health endpoint failed')
            }
            return ''
          },
          outputs: () => ({ apiUrl: 'https://example.com', appTableName: 'AppTable' }),
        }),
      /health endpoint failed/
    )
  })
})

test('invite-admin fails when deploy outputs are missing', async () => {
  await withTarget('staging', persistent, async () => {
    const ops = await import('./ops.mjs')
    await assert.rejects(
      () =>
        ops.inviteAdmin('jane@example.com', {
          run: () => '123456789012',
          outputs: () => {
            throw new Error('SST outputs not found at; deploy first')
          },
        }),
      /deploy outputs|outputs not found/
    )
  })
})
