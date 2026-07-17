import assert from 'node:assert/strict';
import test from 'node:test';
import { extractBootstrapRoutes, normalizePath, validate, validateNoCommittedAuthSecrets } from './check-bruno.mjs';

test('normalizes Bruno variables and removes query strings', () => {
  assert.equal(normalizePath('{{apiUrl}}/api/v1/services/{{serviceId}}?status=open'), '/api/v1/services/{serviceId}');
});

test('extracts SST method and path declarations', () => {
  assert.deepEqual(extractBootstrapRoutes("api.route('GET /api/health', handler)").map(({ method, path }) => ({ method, path })), [
    { method: 'GET', path: '/api/health' },
  ]);
});

test('extracts shared protected v1 route declarations', () => {
  assert.deepEqual(
    extractBootstrapRoutes("const protectedV1Routes = [\n  'GET /api/v1/items',\n  'POST /api/v1/items',\n  ]").map(({ method, path }) => ({ method, path })),
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

test('uses SST authentication classification without storing credentials', () => {
  const result = validate({
    bootstrapSource: "api.route('GET /api/health', handler)\napi.route('GET /api/v1/items', { handler: '../services/monitor-api', auth: { jwt: {} } })",
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

  assert.match(result.errors[0], /SST classifies GET \/api\/health as public; do not send authentication/);
  assert.match(result.errors[1], /SST classifies GET \/api\/v1\/items as protected; use inherited Bearer \{\{accessToken\}\} authentication/);
});

test('permits only explicitly marked Cognito authentication helpers outside SST routes', () => {
  const result = validate({
    bootstrapSource: "api.route('GET /api/health', handler)",
    requests: [
      {
        filePath: 'auth/login.yml',
        method: 'POST',
        path: 'https://cognito-idp.{cognitoRegion}.amazonaws.com/',
        external: 'cognito',
      },
    ],
  });
  assert.match(result.errors[0], /missing Bruno request: GET \/api\/health/);
  assert.equal(result.errors.length, 1);
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
