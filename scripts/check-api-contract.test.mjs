import assert from 'node:assert/strict';
import test from 'node:test';
import { extractOpenAPIRoutes, validateContract } from './check-api-contract.mjs';

test('extracts block and inline OpenAPI operations', () => {
  assert.deepEqual(extractOpenAPIRoutes('paths:\n  /api/health:\n    get:\n  /api/v1/things: {post: {}}'), [
    { method: 'GET', path: '/api/health', source: 'openapi/openapi.yaml:3' },
    { method: 'POST', path: '/api/v1/things', source: 'openapi/openapi.yaml:4' },
  ]);
});

test('reports stale OpenAPI, missing Bruno, and handler drift', () => {
  const result = validateContract({
    bootstrapSource: "api.route('GET /api/health', health)\napi.route('GET /api/v1/items/{itemId}', monitorHandler)",
    brunoRequests: [],
    openapiSource: 'paths:\n  /api/old: {get: {}}',
    handlerSource: 'var monitorRoutes = []routeDefinition{{"GET", "/api/v1/other"}}',
    openSpecRoutes: [{ method: 'GET', path: '/api/v1/required', source: 'openspec/specs/example/spec.md:4' }],
  });
  assert.match(result.errors.join('\n'), /Bruno missing GET \/api\/health/);
  assert.match(result.errors.join('\n'), /OpenAPI stale GET \/api\/old/);
  assert.match(result.errors.join('\n'), /handler inventory missing GET \/api\/v1\/items\/\{itemId\}/);
  assert.match(result.errors.join('\n'), /SST missing GET \/api\/v1\/required/);
});

test('reports parameter mismatches and missing merged requirements with remedies', () => {
  const result = validateContract({
    bootstrapSource: "api.route('GET /api/v1/items/{itemId}', monitorHandler)",
    brunoRequests: [{ method: 'GET', path: '/api/v1/items/{id}', filePath: 'items/get.yml' }],
    openapiSource: 'paths:\n  /api/v1/items/{id}: {get: {}}',
    handlerSource: 'var monitorRoutes = []routeDefinition{{"GET", "/api/v1/items/{itemId}"}}',
    openSpecRoutes: [{ method: 'GET', path: '/api/v1/required', source: 'openspec/specs/example/spec.md:4' }],
  });
  const errors = result.errors.join('\n');
  assert.match(errors, /Bruno missing GET \/api\/v1\/items\/\{itemId\}; add route matching SST/);
  assert.match(errors, /OpenAPI stale GET \/api\/v1\/items\/\{id\}.*remove it or wire it in SST/);
  assert.match(errors, /SST missing GET \/api\/v1\/required; add route matching merged OpenSpec/);
});
