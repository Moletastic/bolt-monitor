import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, realpathSync, writeFileSync } from 'node:fs'
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
  budgetAlertsOptOut: true,
  budgetAlertsOptOutReason: 'The account disallows AWS Budgets for this test fixture.',
}

const expiredEphemeral = {
  stage: 'dev-jane-expired',
  profile: 'bolt-monitor',
  lifecycle: 'ephemeral',
  owner: 'jane',
  service: 'bolt-monitor',
  accountId: '123456789012',
  region: 'us-east-1',
  dashboardOrigin: 'https://dev-jane.example.com',
  disposable: true,
  expiresAt: '2000-01-01T00:00:00Z',
}

async function withTarget(name, target, run) {
  const dir = mkdtempSync(join(tmpdir(), 'bolt-ops-'))
  const targetsDir = join(dir, 'infra', 'targets')
  mkdirSync(targetsDir, { recursive: true })
  const path = join(targetsDir, `${name}.target.json`)
  writeFileSync(path, JSON.stringify(target))
  const previousFile = process.env.TARGET_FILE
  delete process.env.TARGET
  process.env.TARGET_FILE = path
  try {
    await run()
  } finally {
    if (previousFile === undefined) delete process.env.TARGET_FILE
    else process.env.TARGET_FILE = previousFile
  }
}

test('preflight and SST receive the same explicit target path', async () => {
  await withTarget('staging', persistent, async () => {
    const ops = await import('./ops.mjs')
    const expectedPath = realpathSync(process.env.TARGET_FILE)
    let receivedEnvironment

    await ops.status({
      run: (_command, _args, environment) => {
        receivedEnvironment = environment
        return persistent.accountId
      },
    })

    assert.equal(receivedEnvironment.SST_TARGET_FILE, expectedPath)
  })
})

test('conflicting target name and explicit path fail before preflight', async () => {
  await withTarget('staging', persistent, async () => {
    const ops = await import('./ops.mjs')
    const previousTarget = process.env.TARGET
    process.env.TARGET = 'local'
    try {
      assert.throws(
        () => ops.resolveTarget(),
        /conflicting target selection.*TARGET=local.*TARGET_FILE=/
      )
    } finally {
      if (previousTarget === undefined) delete process.env.TARGET
      else process.env.TARGET = previousTarget
    }
  })
})

test('expired ephemeral targets run verified cleanup for removal', async () => {
  await withTarget('expired', expiredEphemeral, async () => {
    const ops = await import('./ops.mjs')
    let cleanupTarget
    let cleanupEnvironment

    await ops.remove(
      {},
      {
        preflight: () => ({
          accountId: expiredEphemeral.accountId,
          region: expiredEphemeral.region,
        }),
        cleanup: (target, environment) => {
          cleanupTarget = target
          cleanupEnvironment = environment
          return { coveredResourceKinds: [] }
        },
      }
    )

    assert.equal(cleanupTarget.stage, expiredEphemeral.stage)
    assert.equal(cleanupEnvironment.SST_OPERATION, 'remove')
  })
})

test('expired ephemeral targets reject deployment before preflight', async () => {
  await withTarget('expired', expiredEphemeral, async () => {
    const ops = await import('./ops.mjs')

    await assert.rejects(
      () =>
        ops.deploy({
          preflight: () => {
            throw new Error('preflight must not run')
          },
        }),
      /expiresAt must be in the future/
    )
  })
})

test('expired ephemeral targets reject every non-removal operation', async () => {
  await withTarget('expired', expiredEphemeral, async () => {
    const ops = await import('./ops.mjs')
    const operations = [
      () => ops.status(),
      () => ops.dev(),
      () => ops.inviteAdmin('jane@example.com'),
      () => ops.rotateAuthKey(),
    ]

    for (const operation of operations) {
      await assert.rejects(operation, /expiresAt must be in the future/)
    }
  })
})

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

test('target summary presents non-secret target details one field per line', async () => {
  const ops = await import('./ops.mjs')

  assert.equal(
    ops.targetSummary(persistent, persistent.accountId),
    [
      '  target: staging',
      '  stage: staging',
      '  class: persistent',
      '  owner: platform',
      '  service: bolt-monitor',
      '  account: 123456789012',
      '  region: us-east-1',
      '  profile: bolt-monitor',
      '  dashboard origin: https://staging.example.com',
    ].join('\n')
  )
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
          outputs: () => ({
            apiUrl: 'https://example.com',
            dashboardUrl: 'https://dashboard.example.com',
            appTableName: 'AppTable',
            authTableName: 'AuthTable',
          }),
        }),
      /health endpoint failed/
    )
  })
})

test('deploy outputs accept flat and selected stage-scoped shapes', async () => {
  const ops = await import('./ops.mjs')
  const output = {
    apiUrl: 'https://api.example.com',
    dashboardUrl: 'https://dashboard.example.com',
    appTableName: 'AppTable',
    authTableName: 'AuthTable',
  }

  assert.deepEqual(ops.resolveDeployOutputs(output, 'staging'), output)
  assert.deepEqual(ops.resolveDeployOutputs({ staging: output }, 'staging'), output)
})

test('deploy outputs reject missing, malformed, and stage-mismatched shapes', async () => {
  const ops = await import('./ops.mjs')
  const output = {
    apiUrl: 'https://api.example.com',
    dashboardUrl: 'https://dashboard.example.com',
    appTableName: 'AppTable',
    authTableName: 'AuthTable',
  }

  assert.throws(
    () => ops.resolveDeployOutputs({ ...output, authTableName: '' }, 'staging'),
    /missing required fields.*authTableName/
  )
  assert.throws(() => ops.resolveDeployOutputs([], 'staging'), /outputs are malformed/)
  assert.throws(
    () => ops.resolveDeployOutputs({ development: output }, 'staging'),
    /do not contain selected stage: staging/
  )
})

test('persistent deploy verifies health before protection and PITR', async () => {
  await withTarget('staging', persistent, async () => {
    const ops = await import('./ops.mjs')
    const calls = []

    await ops.deploy({
      preflight: () => ({ accountId: persistent.accountId, region: persistent.region }),
      run: (command, args) => {
        calls.push([command, args])
        if (command === 'aws' && args[0] === 'dynamodb' && args[1] === 'describe-table') {
          return JSON.stringify({ Table: { DeletionProtectionEnabled: true } })
        }
        if (command === 'aws' && args[0] === 'dynamodb') {
          return JSON.stringify({
            ContinuousBackupsDescription: {
              PointInTimeRecoveryDescription: { PointInTimeRecoveryStatus: 'ENABLED' },
            },
          })
        }
        return ''
      },
      outputs: () => ({
        apiUrl: 'https://api.example.com',
        dashboardUrl: 'https://dashboard.example.com',
        appTableName: 'AppTable',
        authTableName: 'AuthTable',
      }),
    })

    assert.equal(calls[1][0], 'curl')
    assert.deepEqual(calls[1][1], ['-fsS', 'https://api.example.com/api/health'])
    assert.equal(calls[2][0], 'aws')
    assert.equal(calls[3][0], 'aws')
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
