import assert from 'node:assert/strict'
import test from 'node:test'
import { lifecyclePolicy, type DeploymentTarget } from './deployment-target.ts'

const base: Omit<DeploymentTarget, 'lifecycle'> = {
  stage: 'staging',
  profile: 'bolt-monitor',
  owner: 'platform',
  service: 'bolt-monitor',
  accountId: '123456789012',
  region: 'us-east-1',
  dashboardOrigin: 'https://staging.example.com',
}

test('persistent policy retains only durable resources and protects stack', () => {
  const policy = lifecyclePolicy({ ...base, lifecycle: 'persistent', approved: true })
  assert.equal(policy.appProtect, true)
  assert.equal(policy.retainDurableResources, true)
  assert.deepEqual(policy.tags, {
    service: 'bolt-monitor',
    stage: 'staging',
    owner: 'platform',
    lifecycle: 'persistent',
  })
})

test('ephemeral policy is removable and identifies its expiry', () => {
  const policy = lifecyclePolicy({
    ...base,
    stage: 'dev-jane',
    lifecycle: 'ephemeral',
    disposable: true,
    expiresAt: '2099-01-01T00:00:00Z',
  })
  assert.equal(policy.appProtect, false)
  assert.equal(policy.retainDurableResources, false)
  assert.equal(policy.tags.expiresAt, '2099-01-01T00:00:00Z')
})
