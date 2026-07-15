import assert from 'node:assert/strict';
import test from 'node:test';
import { cleanupEphemeral, residualInventory } from './sst-cleanup.mjs';

const target = { stage: 'dev-jane', lifecycle: 'ephemeral', disposable: true, service: 'bolt-monitor' };
const empty = () => JSON.stringify({ ResourceTagMappingList: [] });

test('cleanup rejects persistent targets and succeeds idempotently with zero residuals', () => {
  assert.throws(() => cleanupEphemeral({ ...target, lifecycle: 'persistent' }, {}, () => {}, empty), /refuses persistent/);
  assert.deepEqual(cleanupEphemeral(target, {}, () => {}, empty).residuals, []);
  assert.deepEqual(cleanupEphemeral(target, {}, () => {}, empty).residuals, []);
});

test('cleanup preserves provider failure and reports exact owned residual identifiers', () => {
  const query = () => JSON.stringify({ ResourceTagMappingList: [{ ResourceARN: 'arn:aws:s3:::bolt-monitor-dev-jane' }] });
  assert.throws(() => cleanupEphemeral(target, {}, () => { throw new Error('bucket version deletion failed'); }, query), /bucket version deletion failed.*arn:aws:s3/);
});

test('residual inventory does not select similarly named foreign resources', () => {
  const seen = [];
  residualInventory(target, {}, (_, args) => { seen.push(args); return JSON.stringify({ ResourceTagMappingList: [] }); });
  assert.deepEqual(seen[0].slice(-2), ['Key=service,Values=bolt-monitor', 'Key=stage,Values=dev-jane']);
});
