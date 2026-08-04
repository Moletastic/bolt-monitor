import assert from 'node:assert/strict';
import test from 'node:test';
import { readRepositoryFile, walkFiles } from './helpers.mjs';

test('reads repository files and recursively finds matching files', () => {
  assert.match(readRepositoryFile('package.json'), /@bolt-monitor\/root/);
  assert.ok(walkFiles(new URL('.', import.meta.url).pathname, '.test.mjs').some((filePath) => filePath.endsWith('helpers.test.mjs')));
});
