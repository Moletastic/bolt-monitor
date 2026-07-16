import assert from 'node:assert/strict';
import test from 'node:test';
import { extractBootstrapRoutes, normalizePath, validate, validateNoCommittedAuthSecrets } from './check-bruno.mjs';

test('normalizes Bruno variables and removes query strings', () => {
  assert.equal(normalizePath('{{apiUrl}}/api/v1/services/{{serviceId}}?status=open'), '/api/v1/services/{serviceId}');
});

test('extracts SST method and path declarations', () => {
  assert.deepEqual(extractBootstrapRoutes("api.route('GET /api/health', handler)"), [
    { method: 'GET', path: '/api/health' },
  ]);
});

test('extracts shared protected v1 route declarations', () => {
  assert.deepEqual(
    extractBootstrapRoutes("const protectedV1Routes = [\n  'GET /api/v1/items',\n  'POST /api/v1/items',\n  ]"),
    [
      { method: 'GET', path: '/api/v1/items' },
      { method: 'POST', path: '/api/v1/items' },
    ]
  );
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

test('requires the access token for versioned requests and keeps health public', () => {
  const result = validate({
    bootstrapSource: "api.route('GET /api/health', handler)\nconst protectedV1Routes = [\n  'GET /api/v1/items',\n  ]",
    requests: [
      {
        filePath: 'health.yml',
        folder: 'health',
        name: 'Health Check',
        method: 'GET',
        path: '/api/health',
        variables: [],
        tags: ['domain:health', 'operation:read'],
        docs: 'Purpose: x\nSetup: x\nExpected result: x',
        auth: 'access-token',
      },
      {
        filePath: 'items.yml',
        folder: 'items',
        name: 'List Items',
        method: 'GET',
        path: '/api/v1/items',
        variables: [],
        tags: ['domain:items', 'operation:list'],
        docs: 'Purpose: x\nSetup: x\nExpected result: x',
        auth: 'none',
      },
    ],
  });

  assert.match(result.errors[0], /health must not send authentication/);
  assert.match(result.errors[1], /require Bearer \{\{accessToken\}\} authentication/);
});

test('rejects literal auth secrets while allowing local placeholders', () => {
  const errors = validateNoCommittedAuthSecrets([
    {
      filePath: 'example.yml',
      source: 'accessToken: eyJhbGciOiJIUzI1NiJ9.payload.signature\npassword: hunter2\nAuthorization: Bearer secret-token',
    },
    {
      filePath: 'safe.yml',
      source: 'accessToken: "{{accessToken}}"\ncognitoPassword: replace-with-password\nchallengeSession: <returned-session>',
    },
  ]);

  assert.equal(errors.length, 3);
  assert.match(errors[0], /accessToken/);
  assert.match(errors[1], /password/);
  assert.match(errors[2], /Authorization header/);
});
