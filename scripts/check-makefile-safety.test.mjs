import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const root = new URL('..', import.meta.url);
const makefile = readFileSync(new URL('Makefile', root), 'utf8');

test('Makefile defaults to documented non-mutating help', () => {
  const output = execFileSync('make', [], { cwd: root, encoding: 'utf8' });

  assert.match(makefile, /^\.DEFAULT_GOAL := help$/m);
  assert.match(output, /^  help\s+Show documented Make targets$/m);
  assert.match(output, /^  setup\s+Install JavaScript dependencies/m);
  assert.doesNotMatch(output, /pnpm --dir|go work sync/);
});

test('Makefile enables strict Bash execution and failed-target cleanup', () => {
  assert.match(makefile, /^SHELL := \/bin\/bash$/m);
  assert.match(makefile, /^\.SHELLFLAGS := -euo pipefail -c$/m);
  assert.match(makefile, /^\.DELETE_ON_ERROR:$/m);
});

test('persistent removal receives caller-supplied destructive intent only', () => {
  assert.match(makefile, /remove-infra:.*\n\tnode .* remove DESTROY=\$\(DESTROY\)/);
  assert.doesNotMatch(makefile, /remove DESTROY=yes/);
});

test('targeted formatting documents FILES path constraints', () => {
  assert.match(makefile, /format-dashboard-files: ## .*whitespace-delimited; no spaces or single quotes/);
  assert.match(makefile, /format-infra-files: ## .*whitespace-delimited; no spaces or single quotes/);
});
