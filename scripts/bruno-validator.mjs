import fs from 'node:fs';
import path from 'node:path';
import { createRoute, normalizeRoutePath, routeKey } from './route-record.mjs';

function walk(directory) {
  return fs.readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const fullPath = path.join(directory, entry.name);
    return entry.isDirectory() ? walk(fullPath) : [fullPath];
  });
}

function block(source, key) {
  const lines = source.split('\n');
  const start = lines.findIndex((line) => line === `${key}:` || line.startsWith(`${key}: `));
  if (start < 0) return '';
  const end = lines.slice(start + 1).findIndex((line) => /^\S/.test(line));
  return lines.slice(start, end < 0 ? lines.length : start + 1 + end).join('\n');
}

export function parseBrunoRequest(source, filePath, collectionsDirectory) {
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
  const relativePath = path.relative(collectionsDirectory, filePath);
  const parts = relativePath.split(path.sep);

  return {
    ...createRoute({ method, path: rawUrl, source: filePath }),
    filePath,
    folder: parts.length > 2 ? parts[1] : '',
    name,
    variables,
    tags,
    docs,
    external: info.match(/^\s+external:\s*(\S+)\s*$/m)?.[1],
  };
}

function authMode(source) {
  if (
    /^\s+auth:\s*$/m.test(source) &&
    /^\s+type:\s+bearer\s*$/m.test(source) &&
    /^\s+token:\s+["']?\{\{accessToken\}\}["']?\s*$/m.test(source)
  ) return 'access-token';
  return /^\s+auth:\s*/m.test(source) ? 'other' : 'none';
}

export function readBrunoRequests(collectionsDirectory) {
  const files = walk(collectionsDirectory)
    .filter((filePath) => filePath.endsWith('.yml'))
    .map((filePath) => ({ filePath, source: fs.readFileSync(filePath, 'utf8') }));
  const folderAuth = new Map(files.filter(({ filePath }) => path.basename(filePath) === 'folder.yml').map(({ filePath, source }) => [path.dirname(filePath), authMode(source)]));

  return files.filter(({ source }) => /^\s*type:\s*http\s*$/m.test(source)).map(({ filePath, source }) => {
    const request = parseBrunoRequest(source, filePath, collectionsDirectory);
    let directory = path.dirname(filePath);
    let auth = authMode(source);
    while (auth === 'none' && directory.startsWith(collectionsDirectory)) {
      auth = folderAuth.get(directory) ?? 'none';
      directory = path.dirname(directory);
    }
    return { ...request, auth };
  });
}

function expectedVariables(routePath) {
  return [...routePath.matchAll(/\{([A-Za-z][A-Za-z0-9_]*)\}/g)].map((match) => match[1]);
}

export function validateBruno({ routes, requests, specRoutes = new Set() }) {
  const expected = new Map(routes.map((route) => [routeKey(route), route]));
  const errors = [];
  const requestKeys = new Set();
  for (const request of requests) {
    if (request.external !== undefined) {
      if (request.external !== "cognito" || !/^https:\/\/cognito-idp\.\{cognitoRegion\}\.amazonaws\.com\/$/.test(request.path))
        errors.push(`${request.filePath}: external requests must target the Cognito regional endpoint`);
      continue;
    }
    const key = routeKey(request);
    requestKeys.add(key);
    if (!expected.has(key)) {
      errors.push(`${request.filePath}: stale route ${key}`);
      continue;
    }
    const expectedRoute = expected.get(key);
    if (expectedRoute.auth === 'protected' && request.auth !== 'access-token') errors.push(`${request.filePath}: SST classifies ${key} as protected; use inherited Bearer {{accessToken}} authentication`);
    if (expectedRoute.auth === 'public' && request.auth !== 'none') errors.push(`${request.filePath}: SST classifies ${key} as public; do not send authentication`);
    const routeVariables = expectedVariables(expectedRoute.path);
    if (JSON.stringify(request.variables) !== JSON.stringify(routeVariables)) errors.push(`${request.filePath}: route variables must be ${routeVariables.join(', ') || 'none'}`);
    const domainTag = request.tags.filter((tag) => tag.startsWith('domain:'));
    const operationTag = request.tags.filter((tag) => tag.startsWith('operation:'));
    if (domainTag.length !== 1 || operationTag.length !== 1) errors.push(`${request.filePath}: requires exactly one domain:* and operation:* tag`);
    if (domainTag[0] && request.folder && domainTag[0] !== `domain:${request.folder}`) errors.push(`${request.filePath}: domain tag must match folder ${request.folder}`);
    if (!/^[A-Z][A-Za-z]+(?:\s+[A-Za-z][A-Za-z0-9-]*)+$/.test(request.name)) errors.push(`${request.filePath}: name must use Verb Resource form`);
    for (const section of ['Purpose:', 'Setup:', 'Expected result:']) if (!request.docs.includes(section)) errors.push(`${request.filePath}: docs missing ${section}`);
  }
  for (const key of expected.keys()) if (!requestKeys.has(key)) errors.push(`missing Bruno request: ${key}`);
  return { errors, specOnly: [...specRoutes].filter((key) => !expected.has(key)), routeCount: expected.size, requestCount: requests.filter((request) => request.external === undefined).length };
}

export { normalizeRoutePath };
