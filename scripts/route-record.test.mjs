import assert from 'node:assert/strict';
import test from 'node:test';
import { createRoute, normalizeRoutePath } from './route-record.mjs';

test('normalizes query strings while preserving parameter names', () => {
  assert.equal(normalizeRoutePath('{{apiUrl}}/api/v1/services/{{serviceId}}/monitors/{monitorId}?cursor=next'), '/api/v1/services/{serviceId}/monitors/{monitorId}');
  assert.deepEqual(createRoute({ method: 'get', path: '/api/v1/items/{{itemId}}?limit=10', target: 'handler', source: 'fixture:2', auth: 'protected' }), {
    method: 'GET', path: '/api/v1/items/{itemId}', target: 'handler', source: 'fixture:2', auth: 'protected',
  });
});
