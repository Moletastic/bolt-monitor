import assert from 'node:assert/strict';
import test from 'node:test';
import { cleanupEphemeral, residualInventory, stageStateIsDeployed, stateInventory } from './sst-cleanup.mjs';

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

test('state inventory records non-secret logical identifiers and tolerates missing state', () => {
  const execute = () => JSON.stringify({ deployment: { resources: [
    { urn: 'urn:pulumi:dev::app::aws:s3/bucket:Bucket::owned', type: 'aws:s3/bucket:Bucket' },
    { urn: 'urn:pulumi:dev::app::random:index:RandomBytes::secret', type: 'random:index:RandomBytes' },
  ] } });
  assert.deepEqual(stateInventory(target, {}, execute), ['urn:pulumi:dev::app::aws:s3/bucket:Bucket::owned']);
  assert.deepEqual(stateInventory(target, {}, () => { throw new Error('state missing'); }), []);
});

test('cleanup supports interrupted and versioned bucket retry fixtures', () => {
  const providerFailure = () => { throw new Error('versioned bucket deletion interrupted'); };
  assert.throws(() => cleanupEphemeral(target, {}, providerFailure, empty), /versioned bucket deletion interrupted/);
  assert.deepEqual(cleanupEphemeral(target, {}, () => {}, empty).residuals, []);
});

test('state verification covers SST-managed resources after taggable resources are gone', () => {
  assert.equal(stageStateIsDeployed(target, {}, () => 'Stages: staging\nsmoke (not deployed)\ndev-jane (not deployed)'), false);
  assert.equal(stageStateIsDeployed(target, {}, () => 'Stages:\n  dev-jane'), true);
  assert.throws(
    () => cleanupEphemeral(target, {}, () => 'Stages:\n  dev-jane', empty),
    /state=stage remains deployed/
  );
});
