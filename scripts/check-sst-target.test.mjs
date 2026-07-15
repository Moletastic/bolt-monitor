import assert from 'node:assert/strict';
import test from 'node:test';
import { resolveSSTTarget } from './check-sst-target.mjs';

test('uses portable defaults', () => assert.deepEqual(resolveSSTTarget({}), { profile: 'bolt-monitor', stage: 'staging' }));
test('rejects production targets without AWS access', () => assert.throws(() => resolveSSTTarget({ SST_STAGE: 'production' }), /refusing production/));
