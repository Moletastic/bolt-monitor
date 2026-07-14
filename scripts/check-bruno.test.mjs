import assert from 'node:assert/strict';
import test from 'node:test';
import { extractBootstrapRoutes, normalizePath, validate } from './check-bruno.mjs';

test('normalizes Bruno variables and removes query strings', () => {
  assert.equal(normalizePath('{{apiUrl}}/api/v1/services/{{serviceId}}?status=open'), '/api/v1/services/{serviceId}');
});

test('extracts SST method and path declarations', () => {
  assert.deepEqual(extractBootstrapRoutes("api.route('GET /api/health', handler)"), [
    { method: 'GET', path: '/api/health' },
  ]);
});

test('reports missing, stale, and metadata violations', () => {
  const result = validate({
    bootstrapSource: "api.route('GET /api/health', handler)",
    requests: [
      {
        filePath: 'stale.yml',
        folder: 'health',
        name: 'Read Health',
        method: 'GET',
        path: '/api/old',
        variables: [],
        tags: ['domain:health', 'operation:read'],
        docs: 'Purpose: x\nSetup: x\nExpected result: x',
      },
    ],
  });
  assert.equal(result.errors.length, 2);
  assert.match(result.errors[0], /stale route/);
  assert.match(result.errors[1], /missing Bruno request/);
});

test('reports spec routes missing from bootstrap separately', () => {
  const result = validate({
    bootstrapSource: "api.route('GET /api/health', handler)",
    requests: [],
    specRoutes: new Set(['GET /api/probe-locations']),
  });
  assert.deepEqual(result.specOnly, ['GET /api/probe-locations']);
  assert.match(result.errors[0], /missing Bruno request/);
});
