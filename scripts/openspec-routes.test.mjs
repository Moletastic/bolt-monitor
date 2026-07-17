import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { extractMergedOpenSpecRoutes } from './openspec-routes.mjs';

test('extracts only merged specification routes with source lines', () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'openspec-routes-'));
  const specs = path.join(root, 'openspec/specs/capability');
  const changes = path.join(root, 'openspec/changes/future');
  fs.mkdirSync(specs, { recursive: true });
  fs.mkdirSync(changes, { recursive: true });
  fs.writeFileSync(path.join(specs, 'spec.md'), 'Requirement\n\n`GET /api/v1/items/{itemId}?cursor=next`\n');
  fs.writeFileSync(path.join(changes, 'spec.md'), '`POST /api/future`\n');
  assert.deepEqual(extractMergedOpenSpecRoutes(path.join(root, 'openspec/specs')), [{
    method: 'GET', path: '/api/v1/items/{itemId}', source: `${path.join(specs, 'spec.md')}:3`,
  }]);
  fs.rmSync(root, { recursive: true, force: true });
});
