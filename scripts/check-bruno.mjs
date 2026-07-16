import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { execFileSync } from 'node:child_process';

const root = path.resolve(new URL('..', import.meta.url).pathname);

function walk(directory) {
  return fs.readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const fullPath = path.join(directory, entry.name);
    return entry.isDirectory() ? walk(fullPath) : [fullPath];
  });
}

export function normalizePath(rawUrl) {
  const withoutBase = rawUrl.replace(/^\{\{apiUrl\}\}/, '');
  const withoutQuery = withoutBase.split('?', 1)[0];
  return withoutQuery.replace(/\{\{([A-Za-z][A-Za-z0-9_]*)\}\}/g, '{$1}');
}

export function extractBootstrapRoutes(source) {
  const routes = [];
  const routePattern = /api\.route\(\s*['"]([A-Z]+)\s+([^'"]+)['"]/g;
  for (const match of source.matchAll(routePattern)) {
    routes.push({ method: match[1], path: match[2] });
  }
  const protectedRoutes = source.match(/const protectedV1Routes = \[([\s\S]*?)\n  \]/)?.[1] ?? '';
  for (const match of protectedRoutes.matchAll(/['"]([A-Z]+)\s+(\/api\/v1\/[^'"]+)['"]/g)) {
    routes.push({ method: match[1], path: match[2] });
  }
  return routes;
}

function block(source, key) {
  const lines = source.split('\n');
  const start = lines.findIndex((line) => line === `${key}:` || line.startsWith(`${key}: `));
  if (start < 0) return '';
  const end = lines.slice(start + 1).findIndex((line) => /^\S/.test(line));
  return lines.slice(start, end < 0 ? lines.length : start + 1 + end).join('\n');
}

export function parseBrunoRequest(source, filePath) {
  const info = block(source, 'info');
  const http = block(source, 'http');
  const docs = source.match(/^docs:\s*\|-\s*\n([\s\S]*)$/m)?.[1] ?? '';
  const tagsBlock = info.match(/^\s+tags:\s*\n((?:\s+-\s+.*\n?)*)/m)?.[1] ?? '';
  const tags = [...tagsBlock.matchAll(/^\s+-\s+(.+)$/gm)].map((match) => match[1].trim());
  const name = info.match(/^\s+name:\s+(.+)$/m)?.[1]?.trim() ?? '';
  const method = http.match(/^\s+method:\s+([A-Z]+)$/m)?.[1] ?? '';
  const rawUrl = http.match(/^\s+url:\s+["']([^"']+)["']$/m)?.[1] ?? '';
  const variables = [...rawUrl.matchAll(/\{\{([A-Za-z][A-Za-z0-9_]*)\}\}/g)]
    .map((match) => match[1])
    .filter((variable) => variable !== 'apiUrl');
  const relativePath = path.relative(path.join(root, '.bruno', 'collections'), filePath);
  const parts = relativePath.split(path.sep);

  return {
    filePath,
    folder: parts.length > 2 ? parts[1] : '',
    name,
    method,
    path: normalizePath(rawUrl),
    variables,
    tags,
    docs,
  };
}

function authMode(source) {
  if (
    /^\s+auth:\s*$/m.test(source) &&
    /^\s+mode:\s+bearer\s*$/m.test(source) &&
    /^\s+token:\s+["']?\{\{accessToken\}\}["']?\s*$/m.test(source)
  ) {
    return 'access-token';
  }
  return /^\s+auth:\s*/m.test(source) ? 'other' : 'none';
}

export function readBrunoRequests(collectionsDirectory) {
  const files = walk(collectionsDirectory)
    .filter((filePath) => filePath.endsWith('.yml'))
    .map((filePath) => ({ filePath, source: fs.readFileSync(filePath, 'utf8') }));
  const folderAuth = new Map(
    files
      .filter(({ filePath }) => path.basename(filePath) === 'folder.yml')
      .map(({ filePath, source }) => [path.dirname(filePath), authMode(source)])
  );

  return files
    .filter(({ source }) => /^\s*type:\s*http\s*$/m.test(source))
    .map(({ filePath, source }) => {
      const request = parseBrunoRequest(source, filePath);
      let directory = path.dirname(filePath);
      let auth = authMode(source);
      while (auth === 'none' && directory.startsWith(collectionsDirectory)) {
        auth = folderAuth.get(directory) ?? 'none';
        directory = path.dirname(directory);
      }
      return { ...request, auth };
    });
}

function routeKey(route) {
  return `${route.method} ${route.path}`;
}

function expectedVariables(routePath) {
  return [...routePath.matchAll(/\{([A-Za-z][A-Za-z0-9_]*)\}/g)].map((match) => match[1]);
}

function openSpecRoutes(specDirectory) {
  const routes = new Set();
  for (const filePath of walk(specDirectory).filter((file) => file.endsWith('.md'))) {
    const source = fs.readFileSync(filePath, 'utf8');
    for (const match of source.matchAll(/`(GET|POST|PUT|PATCH|DELETE)\s+(\/api\/[^`\s]+)`/g)) {
      routes.add(`${match[1]} ${match[2].split('?', 1)[0]}`.replace(/\{\{([^}]+)\}\}/g, '{$1}'));
    }
  }
  return routes;
}

export function validate({ bootstrapSource, requests, specRoutes = new Set() }) {
  const bootstrapRoutes = extractBootstrapRoutes(bootstrapSource);
  const expected = new Map(bootstrapRoutes.map((route) => [routeKey(route), route]));
  const errors = [];
  const requestKeys = new Set();

  for (const request of requests) {
    const key = `${request.method} ${request.path}`;
    requestKeys.add(key);
    if (!expected.has(key)) {
      errors.push(`${request.filePath}: stale route ${key}`);
      continue;
    }
    const expectedRoute = expected.get(key);
    if (expectedRoute.path.startsWith('/api/v1/') && request.auth !== 'access-token') {
      errors.push(`${request.filePath}: versioned routes require Bearer {{accessToken}} authentication`);
    }
    if (expectedRoute.path === '/api/health' && request.auth !== 'none') {
      errors.push(`${request.filePath}: health must not send authentication`);
    }
    const routeVariables = expectedVariables(expectedRoute.path);
    if (JSON.stringify(request.variables) !== JSON.stringify(routeVariables)) {
      errors.push(`${request.filePath}: route variables must be ${routeVariables.join(', ') || 'none'}`);
    }
    const domainTag = request.tags.filter((tag) => tag.startsWith('domain:'));
    const operationTag = request.tags.filter((tag) => tag.startsWith('operation:'));
    if (domainTag.length !== 1 || operationTag.length !== 1) {
      errors.push(`${request.filePath}: requires exactly one domain:* and operation:* tag`);
    }
    if (domainTag[0] && request.folder && domainTag[0] !== `domain:${request.folder}`) {
      errors.push(`${request.filePath}: domain tag must match folder ${request.folder}`);
    }
    if (!/^[A-Z][A-Za-z]+(?:\s+[A-Za-z][A-Za-z0-9-]*)+$/.test(request.name)) {
      errors.push(`${request.filePath}: name must use Verb Resource form`);
    }
    for (const section of ['Purpose:', 'Setup:', 'Expected result:']) {
      if (!request.docs.includes(section)) errors.push(`${request.filePath}: docs missing ${section}`);
    }
  }

  for (const key of expected.keys()) {
    if (!requestKeys.has(key)) errors.push(`missing Bruno request: ${key}`);
  }

  const specOnly = [...specRoutes].filter((key) => !expected.has(key));
  return { errors, specOnly, routeCount: expected.size, requestCount: requests.length };
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
      if (value && !placeholder.test(value)) {
        errors.push(`${filePath}:${lineNumber + 1}: committed auth secret in ${match[1].trim()}`);
      }
    }
    for (const [lineNumber, line] of source.split('\n').entries()) {
      if (/authorization\s*:\s*bearer\s+(?!\{\{accessToken\}\}\s*$)\S+/i.test(line)) {
        errors.push(`${filePath}:${lineNumber + 1}: committed auth secret in Authorization header`);
      }
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
  const bootstrapPath = path.join(root, 'infra', 'stacks', 'bootstrap.ts');
  const collectionsPath = path.join(root, '.bruno', 'collections');
  const specsPath = path.join(root, 'openspec', 'specs');
  const result = validate({
    bootstrapSource: fs.readFileSync(bootstrapPath, 'utf8'),
    requests: readBrunoRequests(collectionsPath),
    specRoutes: openSpecRoutes(specsPath),
  });
  result.errors.push(
    ...validateNoCommittedAuthSecrets(readTrackedBrunoFiles())
  );

  console.log(`Bruno route coverage: ${result.requestCount}/${result.routeCount}`);
  for (const route of result.specOnly) console.warn(`OpenSpec route not wired in bootstrap: ${route}`);
  for (const error of result.errors) console.error(`ERROR ${error}`);
  if (result.errors.length > 0) process.exitCode = 1;
}

if (process.argv[1] === new URL(import.meta.url).pathname) main();
