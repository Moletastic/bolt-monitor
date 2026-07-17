import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { execFileSync } from 'node:child_process';
import { extractSSTRoutes } from './sst-routes.mjs';
import { extractMergedOpenSpecRoutes } from './openspec-routes.mjs';
import { normalizeRoutePath, parseBrunoRequest, readBrunoRequests, validateBruno } from './bruno-validator.mjs';

const root = path.resolve(new URL('..', import.meta.url).pathname);

export { normalizeRoutePath as normalizePath, parseBrunoRequest, readBrunoRequests };

export function extractBootstrapRoutes(source) {
  return extractSSTRoutes(source).routes;
}

export function validate({ bootstrapSource, requests, specRoutes = new Set() }) {
  return validateBruno({ routes: extractBootstrapRoutes(bootstrapSource), requests, specRoutes });
}

const authSecretKey = /(?:access|id|refresh)[_-]?token|(?:cognito[_-]?)?(?:new|temporary)?password|challenge[_-]?session|recovery[_-]?code|(?:dashboard[_-]?)?session(?:[_-]?(?:id|value))?|client[_-]?secret/i;
const placeholder = /^(?:\{\{[A-Za-z][A-Za-z0-9_]*\}\}|replace-with-[\w-]+|<[^>]+>)$/;

export function validateNoCommittedAuthSecrets(files) {
  const errors = [];
  for (const { filePath, source } of files) {
    for (const [lineNumber, line] of source.split('\n').entries()) {
      const match = line.match(/^\s*([A-Za-z][A-Za-z0-9_-]*)\s*:\s*["']?([^"'#]+)["']?\s*(?:#.*)?$/);
      if (!match || !authSecretKey.test(match[1].trim())) continue;
      const value = match[2].trim();
      if (value && !placeholder.test(value)) errors.push(`${filePath}:${lineNumber + 1}: committed auth secret in ${match[1].trim()}`);
    }
    for (const [lineNumber, line] of source.split('\n').entries()) {
      if (/authorization\s*:\s*bearer\s+(?!\{\{accessToken\}\}\s*$)\S+/i.test(line)) errors.push(`${filePath}:${lineNumber + 1}: committed auth secret in Authorization header`);
    }
  }
  return errors;
}

function readTrackedBrunoFiles() {
  return execFileSync('git', ['ls-files', '-z', '--', '.bruno'], { cwd: root, encoding: 'utf8' })
    .split('\0')
    .filter((filePath) => filePath.endsWith('.yml') || filePath.endsWith('.md'))
    .map((filePath) => {
      const absolutePath = path.join(root, filePath);
      return { filePath: absolutePath, source: fs.readFileSync(absolutePath, 'utf8') };
    });
}

function main() {
  const result = validate({
    bootstrapSource: fs.readFileSync(path.join(root, 'infra/stacks/bootstrap.ts'), 'utf8'),
    requests: readBrunoRequests(path.join(root, '.bruno/collections')),
    specRoutes: new Set(extractMergedOpenSpecRoutes(path.join(root, 'openspec/specs')).map((route) => `${route.method} ${route.path}`)),
  });
  result.errors.push(...validateNoCommittedAuthSecrets(readTrackedBrunoFiles()));
  console.log(`Bruno route coverage: ${result.requestCount}/${result.routeCount}`);
  for (const route of result.specOnly) console.warn(`OpenSpec route not wired in bootstrap: ${route}`);
  for (const error of result.errors) console.error(`ERROR ${error}`);
  if (result.errors.length > 0) process.exitCode = 1;
}

if (process.argv[1] === new URL(import.meta.url).pathname) main();
