import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { validateOpenAPIAuth } from './check-openapi-auth.mjs';

const root = path.resolve(new URL('..', import.meta.url).pathname);
const source = fs.readFileSync(path.join(root, 'openapi/openapi.yaml'), 'utf8');

test('documents Cognito access-token security for v1 and keeps health public', () => {
  assert.deepEqual(validateOpenAPIAuth(source), []);
});

test('rejects a public versioned operation', () => {
  const publicV1 = source.replace(
    '/api/v1/search: {get:',
    '/api/v1/search:\n    security: []\n    get:'
  );
  assert.match(validateOpenAPIAuth(publicV1).join('\n'), /\/api\/v1\/search overrides required CognitoBearer security/);
});

test('rejects undocumented Gateway and envelope error semantics', () => {
  const undocumented = source.replace('API Gateway may return a non-envelope 401 before Lambda', 'Gateway may reject credentials');
  assert.match(validateOpenAPIAuth(undocumented).join('\n'), /Gateway non-envelope 401/);
});
