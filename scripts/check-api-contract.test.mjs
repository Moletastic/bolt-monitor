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
    bootstrapSource: `const monitorHandler = {
  handler: '../services/monitor-api',
  }
api.route('GET /api/health', health)
api.route('GET /api/v1/items/{itemId}', monitorHandler)`,
    brunoRequests: [],
    openapiSource: 'paths:\n  /api/old: {get: {}}',
    handlerSource: 'var monitorRoutes = []routeDefinition{{"GET", "/api/v1/other", "other"}}',
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
    handlerSource: 'var monitorRoutes = []routeDefinition{{"GET", "/api/v1/items/{itemId}", "getItem"}}',
    openSpecRoutes: [{ method: 'GET', path: '/api/v1/required', source: 'openspec/specs/example/spec.md:4' }],
  });
  const errors = result.errors.join('\n');
  assert.match(errors, /Bruno missing GET \/api\/v1\/items\/\{itemId\}; add route matching SST/);
  assert.match(errors, /OpenAPI stale GET \/api\/v1\/items\/\{id\}.*remove it or wire it in SST/);
  assert.match(errors, /SST missing GET \/api\/v1\/required; add route matching merged OpenSpec/);
});

test('reports handler-only and SST-only monitor routes using resolved SST target metadata', () => {
  const result = validateContract({
    bootstrapSource: `const monitorHandler = {
  handler: '../services/monitor-api',
  }
api.route('GET /api/v1/deployed', monitorHandler)
api.route('GET /api/v1/other', otherHandler)`,
    brunoRequests: [
      { method: 'GET', path: '/api/v1/deployed', filePath: 'deployed.yml' },
      { method: 'GET', path: '/api/v1/other', filePath: 'other.yml' },
    ],
    openapiSource: 'paths:\n  /api/v1/deployed: {get: {}}\n  /api/v1/other: {get: {}}',
    handlerSource: 'var monitorRoutes = []routeDefinition{{"GET", "/api/v1/handled", "getHandled"}}',
  });
  const errors = result.errors.join('\n');
  assert.match(errors, /handler inventory missing GET \/api\/v1\/deployed; add route matching SST monitor routes/);
  assert.match(errors, /handler inventory stale GET \/api\/v1\/handled .*wire it in SST monitor routes/);
  assert.doesNotMatch(errors, /GET \/api\/v1\/other.*handler inventory/);
});

test('reports unsupported dynamic monitor inventory entries without inferring routes', () => {
  const result = validateContract({
    bootstrapSource: '',
    brunoRequests: [],
    openapiSource: 'paths: {}',
    handlerSource: 'var monitorRoutes = []routeDefinition{{http.MethodGet, "/api/v1/dynamic", "getDynamic"}}',
  });
  assert.match(result.errors.join('\n'), /services\/monitor-api\/routes.go:1: unsupported dynamic monitorRoutes entry; use a literal method, path, and handler/);
  assert.equal(result.counts.handler, 0);
});

test('reports Bruno and OpenAPI authentication conflicts with the SST route and source', () => {
  const result = validateContract({
    bootstrapSource: "api.route('GET /api/health', { handler: '../services/api-health' })\nconst protectedV1Routes = [\n  'GET /api/v1/items',\n  ]\napi.route(route, monitorHandler, { auth: { jwt: {} } })",
    brunoRequests: [
      { method: 'GET', path: '/api/health', filePath: 'health.yml', auth: 'access-token' },
      { method: 'GET', path: '/api/v1/items', filePath: 'items.yml', auth: 'none' },
    ],
    openapiSource: 'security:\n  - Bearer: []\npaths:\n  /api/health:\n    get: {}\n  /api/v1/items:\n    get:\n      security: []',
    handlerSource: 'var monitorRoutes = []routeDefinition{{"GET", "/api/v1/items", "getItems"}}',
  });
  const errors = result.errors.join('\n');
  assert.match(errors, /Bruno authentication conflict GET \/api\/health: SST is public, Bruno is protected; update Bruno metadata to match SST/);
  assert.match(errors, /Bruno authentication conflict GET \/api\/v1\/items: SST is protected, Bruno is public; update Bruno metadata to match SST/);
  assert.match(errors, /OpenAPI authentication conflict GET \/api\/v1\/items: SST is protected, OpenAPI is public; update OpenAPI metadata to match SST/);
});
