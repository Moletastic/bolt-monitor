import assert from 'node:assert/strict';
import test from 'node:test';
import { extractSSTRoutes } from './sst-routes.mjs';

test('extracts direct and protected SST routes with target, auth, and locations', () => {
  const result = extractSSTRoutes(`api.route('GET /api/health', {
  handler: '../services/api-health',
})
const monitorHandler = {
  handler: '../services/monitor-api',
  }
const protectedV1Routes = [
  'GET /api/v1/items/{itemId}',
  ]
for (const route of protectedV1Routes)
  api.route(route, monitorHandler, {
    auth: { jwt: {} },
  })`);
  assert.deepEqual(result.routes, [
    { method: 'GET', path: '/api/health', target: '../services/api-health', source: 'infra/stacks/bootstrap.ts:1', auth: 'public' },
    { method: 'GET', path: '/api/v1/items/{itemId}', target: '../services/monitor-api', source: 'infra/stacks/bootstrap.ts:7', auth: 'protected' },
  ]);
  assert.deepEqual(result.diagnostics, []);
});

test('reports missing classifications when auth metadata is only partially present', () => {
  const result = extractSSTRoutes(`api.route('GET /api/health', { handler: '../services/api-health' })
const protectedV1Routes = [
  'GET /api/v1/items',
  ]
for (const route of protectedV1Routes)
  api.route(route, monitorHandler)`);
  assert.match(result.diagnostics.join('\n'), /route lacks public\/protected authentication classification/);
});

test('resolves direct variable registrations to their handler target metadata', () => {
  const result = extractSSTRoutes(`const monitorHandler = {
  handler: '../services/monitor-api',
  }
api.route('GET /api/v1/items', monitorHandler)`);
  assert.equal(result.routes[0].target, '../services/monitor-api');
});
