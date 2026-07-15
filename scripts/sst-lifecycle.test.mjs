import assert from 'node:assert/strict';
import { mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import test from 'node:test';
import { confirmationFor } from '../infra/deployment-target.ts';
import { preflight, sstArgs, verifyPersistentDeployment } from './sst-lifecycle.mjs';

const target = {
  stage: 'dev-jane-20260715', lifecycle: 'ephemeral', owner: 'jane', service: 'bolt-monitor',
  accountId: '123456789012', region: 'us-east-1', credentialSource: 'AWS profile bolt-monitor', dashboardOrigin: 'https://dev-jane.example.com',
  disposable: true, expiresAt: '2099-01-01T00:00:00Z',
};
const config = join(mkdtempSync(join(tmpdir(), 'bolt-lifecycle-')), 'target.json');
writeFileSync(config, JSON.stringify({ targets: [target] }));

function environment(overrides = {}) {
  return {
    SST_TARGET_CONFIG: config, SST_STAGE: target.stage, SST_LIFECYCLE_ACTION: 'deploy',
    AWS_REGION: target.region, SST_TARGET_CONFIRMATION: confirmationFor(target, 'deploy'), ...overrides,
  };
}

test('preflight rejects account and region mismatches before mutation', () => {
  assert.throws(() => preflight(environment(), () => '000000000000'), /account mismatch/);
  assert.throws(() => preflight(environment({ AWS_REGION: 'eu-west-1' }), () => target.accountId), /region mismatch/);
});

test('preflight rejects stale confirmation before mutation', () => {
  assert.throws(() => preflight(environment({ SST_TARGET_CONFIRMATION: 'stale' }), () => target.accountId), /confirmation required/);
});

test('preflight summary is secret-safe', () => {
  const result = preflight(environment(), () => target.accountId);
  assert.doesNotMatch(result.summary, /AKIA|secret|token/i);
  assert.match(result.summary, /credential-source=AWS profile bolt-monitor/);
});

test('persistent removal needs separate destructive intent', () => {
  const persistent = { ...target, stage: 'staging', lifecycle: 'persistent', approved: true, disposable: undefined, expiresAt: undefined };
  writeFileSync(config, JSON.stringify({ targets: [persistent] }));
  const env = environment({ SST_STAGE: 'staging', SST_LIFECYCLE_ACTION: 'remove', SST_TARGET_CONFIRMATION: confirmationFor(persistent, 'remove') });
  assert.throws(() => preflight(env, () => persistent.accountId), /destruction requires/);
  writeFileSync(config, JSON.stringify({ targets: [target] }));
});

test('preview fails closed because SST 4.14.1 has no safe preview command', () => {
  assert.throws(() => sstArgs('preview', target), /no safe preview command/);
});

test('persistent deployment verification rejects missing protection and tags', () => {
  const persistent = { ...target, lifecycle: 'persistent', stage: 'staging', approved: true };
  const responses = [
    JSON.stringify({ Table: { TableArn: 'arn:table', DeletionProtectionEnabled: true } }),
    JSON.stringify({ Tags: [{ Key: 'service', Value: 'bolt-monitor' }, { Key: 'stage', Value: 'staging' }] }),
  ];
  assert.throws(() => verifyPersistentDeployment(persistent, 'AppTable', {}, () => responses.shift()), /tag mismatch: owner/);
  assert.throws(() => verifyPersistentDeployment(persistent, 'AppTable', {}, () => JSON.stringify({ Table: { DeletionProtectionEnabled: false } })), /lacks deletion protection/);
});

test('persistent deployment verification accepts matching retained identity', () => {
  const persistent = { ...target, lifecycle: 'persistent', stage: 'staging', owner: 'Moletastic', approved: true };
  const responses = [
    JSON.stringify({ Table: { TableArn: 'arn:table', DeletionProtectionEnabled: true } }),
    JSON.stringify({ Tags: [
      { Key: 'service', Value: 'bolt-monitor' },
      { Key: 'stage', Value: 'staging' },
      { Key: 'owner', Value: 'Moletastic' },
    ] }),
  ];
  assert.doesNotThrow(() => verifyPersistentDeployment(persistent, 'AppTable', {}, () => responses.shift()));
});
