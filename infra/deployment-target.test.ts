import assert from 'node:assert/strict'
import { mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import test from 'node:test'
import {
  loadDeploymentTargetFromPath,
  parseTarget,
  validateDeploymentTarget,
  type DeploymentTarget,
} from './deployment-target.ts'

const persistent: DeploymentTarget = {
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

const ephemeral: DeploymentTarget = {
  stage: 'dev-jane-20260801',
  profile: 'bolt-monitor',
  lifecycle: 'ephemeral',
  owner: 'jane',
  service: 'bolt-monitor',
  accountId: '123456789012',
  region: 'us-east-1',
  dashboardOrigin: 'https://dev-jane.example.com',
  disposable: true,
  expiresAt: '2099-01-01T00:00:00Z',
}

function writeTarget(target: DeploymentTarget): string {
  const path = join(mkdtempSync(join(tmpdir(), 'bolt-target-')), `${target.stage}.target.json`)
  writeFileSync(path, JSON.stringify(target))
  return path
}

test('parses a complete target file', () => {
  const target = parseTarget({
    stage: 'staging',
    profile: 'bolt-monitor',
    lifecycle: 'persistent',
    owner: 'platform',
    service: 'bolt-monitor',
    accountId: '123456789012',
    region: 'us-east-1',
    dashboardOrigin: 'https://staging.example.com',
    approved: true,
  })
  assert.equal(target.profile, 'bolt-monitor')
  assert.equal(target.lifecycle, 'persistent')
  assert.equal(target.approved, true)
})

test('rejects target without AWS profile', () => {
  assert.throws(() => validateDeploymentTarget({ ...persistent, profile: '' }), /profile/)
})

test('rejects ephemeral target without disposable expiry', () => {
  assert.throws(() => validateDeploymentTarget({ ...ephemeral, expiresAt: undefined }), /expiresAt/)
  assert.throws(
    () => validateDeploymentTarget({ ...ephemeral, expiresAt: '2000-01-01T00:00:00Z' }),
    /expiresAt/
  )
})

test('rejects persistent target without approval', () => {
  assert.throws(() => validateDeploymentTarget({ ...persistent, approved: false }), /approved=true/)
})

test('rejects persistent smoke stage names', () => {
  assert.throws(
    () => validateDeploymentTarget({ ...persistent, stage: 'smoke-20260801' }),
    /smoke stage name/
  )
})

test('rejects ephemeral protected name and persistent with disposal fields', () => {
  assert.throws(
    () => validateDeploymentTarget({ ...ephemeral, stage: 'PROD-uction' }),
    /protected name/
  )
  assert.throws(
    () => validateDeploymentTarget({ ...persistent, disposable: true }),
    /cannot declare disposable/
  )
})

test('rejects non-HTTPS dashboard origins and bad account/region shapes', () => {
  assert.throws(
    () =>
      validateDeploymentTarget({ ...persistent, dashboardOrigin: 'http://staging.example.com' }),
    /HTTPS origin/
  )
  assert.throws(
    () => validateDeploymentTarget({ ...persistent, accountId: '1234' }),
    /12-digit AWS account ID/
  )
  assert.throws(() => validateDeploymentTarget({ ...persistent, region: 'east' }), /AWS region/)
})

test('loads target file from path', () => {
  const path = writeTarget(persistent)
  const loaded = loadDeploymentTargetFromPath(path)
  assert.equal(loaded.stage, 'staging')
  assert.equal(loaded.profile, 'bolt-monitor')
  assert.equal(loaded.lifecycle, 'persistent')
})

test('fails closed when target file is unreadable or malformed', () => {
  const dir = mkdtempSync(join(tmpdir(), 'bolt-target-'))
  const missing = join(dir, 'missing.target.json')
  assert.throws(() => loadDeploymentTargetFromPath(missing), /unable to read/)
  const malformed = join(dir, 'malformed.target.json')
  writeFileSync(malformed, '{not-json')
  assert.throws(() => loadDeploymentTargetFromPath(malformed), /unable to read/)
})

test('rejects ephemeral target with unknown lifecycle in parse', () => {
  assert.throws(() => parseTarget({ ...persistent, lifecycle: 'temporary' }), /unknown lifecycle/)
})
