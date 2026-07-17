import assert from 'node:assert/strict';
import test from 'node:test';
import { extractOpenAPIRoutes } from './openapi-routes.mjs';

test('extracts operation lines and resolves inherited, public, and protected security', () => {
  const routes = extractOpenAPIRoutes(`openapi: 3.1.0
security:
  - Bearer: []
paths:
  /api/health:
    get:
      security: []
  /api/v1/items/{itemId}:
    get: {operationId: getItem}
    post:
      security:
        - AlternateBearer: []
`);
  assert.deepEqual(routes, [
    { method: 'GET', path: '/api/health', source: 'openapi/openapi.yaml:6', auth: 'public' },
    { method: 'GET', path: '/api/v1/items/{itemId}', source: 'openapi/openapi.yaml:9', auth: 'protected' },
    { method: 'POST', path: '/api/v1/items/{itemId}', source: 'openapi/openapi.yaml:10', auth: 'protected' },
  ]);
});
