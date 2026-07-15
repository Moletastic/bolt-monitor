import assert from 'node:assert/strict'
import { mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import test from 'node:test'
import {
  loadDeploymentTarget,
  validateDeploymentTarget,
  type DeploymentTarget,
} from './deployment-target.ts'

const persistent: DeploymentTarget = {
  stage: 'staging',
  lifecycle: 'persistent',
  owner: 'platform',
  service: 'bolt-monitor',
  accountId: '123456789012',
  region: 'us-east-1',
  credentialSource: 'AWS profile bolt-monitor',
  dashboardOrigin: 'https://staging.example.com',
  approved: true,
}

const ephemeral: DeploymentTarget = {
  stage: 'dev-jane-20260715',
  lifecycle: 'ephemeral',
  owner: 'jane',
  service: 'bolt-monitor',
  accountId: '123456789012',
  region: 'us-east-1',
  credentialSource: 'AWS profile bolt-monitor',
  dashboardOrigin: 'https://dev-jane.example.com',
  disposable: true,
  expiresAt: '2099-01-01T00:00:00Z',
}

function withTargetConfig(targets: DeploymentTarget[], run: () => void) {
  const path = join(mkdtempSync(join(tmpdir(), 'bolt-target-')), 'target.json')
  writeFileSync(path, JSON.stringify({ targets }))
  const previous = process.env.SST_TARGET_CONFIG
  process.env.SST_TARGET_CONFIG = path
  try {
    run()
  } finally {
    if (previous === undefined) delete process.env.SST_TARGET_CONFIG
    else process.env.SST_TARGET_CONFIG = previous
  }
}

test('loads explicit valid persistent and ephemeral targets', () => {
  withTargetConfig([persistent, ephemeral], () => {
    assert.equal(loadDeploymentTarget('staging').lifecycle, 'persistent')
    assert.equal(loadDeploymentTarget('dev-jane-20260715').lifecycle, 'ephemeral')
  })
})

test('fails closed when stage configuration is absent', () => {
  assert.throws(() => loadDeploymentTarget(undefined), /SST stage is required/)
})

test('rejects incomplete target identity', () => {
  assert.throws(() => validateDeploymentTarget({ ...persistent, owner: '' }, ['staging']), /owner/)
  assert.throws(
    () => validateDeploymentTarget({ ...persistent, accountId: '' }, ['staging']),
    /accountId/
  )
  assert.throws(
    () => validateDeploymentTarget({ ...persistent, region: '' }, ['staging']),
    /region/
  )
  assert.throws(
    () => validateDeploymentTarget({ ...persistent, dashboardOrigin: '' }, ['staging']),
    /dashboardOrigin/
  )
  assert.throws(
    () =>
      validateDeploymentTarget({ ...persistent, dashboardOrigin: 'http://staging.example.com' }, [
        'staging',
      ]),
    /HTTPS origin/
  )
})

test('rejects conflicting persistent and ephemeral fields', () => {
  assert.throws(
    () => validateDeploymentTarget({ ...persistent, disposable: true }, ['staging']),
    /cannot declare disposable/
  )
  assert.throws(
    () => validateDeploymentTarget({ ...ephemeral, approved: true }, ['staging']),
    /cannot declare approved/
  )
  assert.throws(
    () => validateDeploymentTarget({ ...ephemeral, expiresAt: undefined }, ['staging']),
    /expiresAt/
  )
})

test('rejects unapproved persistent and normalized protected ephemeral names', () => {
  assert.throws(
    () => validateDeploymentTarget({ ...persistent, approved: false }, ['staging']),
    /not approved/
  )
  assert.throws(
    () => validateDeploymentTarget({ ...persistent, stage: 'smoke-20260715' }, ['smoke-20260715']),
    /cannot use a smoke stage name/
  )
  assert.throws(
    () => validateDeploymentTarget({ ...ephemeral, stage: 'PROD-uction' }, ['staging']),
    /protected name/
  )
  assert.throws(
    () => validateDeploymentTarget({ ...ephemeral, stage: 'Stag_ing' }, ['staging']),
    /protected name/
  )
})
