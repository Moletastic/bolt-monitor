import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { extractBootstrapRoutes, readBrunoRequests } from './check-bruno.mjs';

const root = path.resolve(new URL('..', import.meta.url).pathname);
const methods = /\b(get|post|put|patch|delete):/gi;

export function normalizePath(value) {
  return value.split('?', 1)[0].replace(/\{\{([^}]+)\}\}/g, '{$1}');
}

export function extractOpenAPIRoutes(source, filePath = 'openapi/openapi.yaml') {
  const routes = [];
  let currentPath = '';
  for (const [index, line] of source.split('\n').entries()) {
    const pathMatch = line.match(/^  (\/api\/[^:]+):/);
    if (pathMatch) {
      currentPath = normalizePath(pathMatch[1]);
      for (const match of line.matchAll(methods)) routes.push({ method: match[1].toUpperCase(), path: currentPath, source: `${filePath}:${index + 1}` });
      continue;
    }
    const methodMatch = line.match(/^    (get|post|put|patch|delete):/i);
    if (currentPath && methodMatch) routes.push({ method: methodMatch[1].toUpperCase(), path: currentPath, source: `${filePath}:${index + 1}` });
  }
  return routes;
}

export function extractHandlerRoutes(source, filePath = 'services/monitor-api/routes.go') {
  return [...source.matchAll(/\{"(GET|POST|PUT|PATCH|DELETE)",\s*"(\/api\/[^\"]+)"\}/g)].map((match) => ({ method: match[1], path: match[2], source: filePath }));
}

export function extractMergedOpenSpecRoutes(specDirectory) {
  const files = fs.readdirSync(specDirectory, { withFileTypes: true }).flatMap((entry) => {
    const filePath = path.join(specDirectory, entry.name);
    if (!entry.isDirectory()) return [filePath];
    const nested = extractMergedOpenSpecRoutes(filePath);
    return nested.map((route) => route.source);
  });
  return [...new Set(files)].filter((filePath) => filePath.endsWith('.md')).flatMap((filePath) => {
    const source = fs.readFileSync(filePath, 'utf8');
    return [...source.matchAll(/`(GET|POST|PUT|PATCH|DELETE)\s+(\/api\/[^`\s?]+)(?:\?[^`]*)?`/g)].map((match) => ({ method: match[1], path: normalizePath(match[2]), source: filePath }));
  });
}

function keys(routes) { return new Map(routes.map((route) => [`${route.method} ${route.path}`, route])); }

function compare(expected, actual, expectedName, actualName) {
  const errors = [];
  for (const [key, route] of expected) if (!actual.has(key)) errors.push(`${actualName} missing ${key}; add route matching ${expectedName} (${route.source ?? 'source'})`);
  for (const [key, route] of actual) if (!expected.has(key)) errors.push(`${actualName} stale ${key} (${route.source ?? 'source'}); remove it or wire it in ${expectedName}`);
  return errors;
}

function requireRoutes(required, actual, requiredName, actualName) {
  const errors = [];
  for (const [key, route] of required) if (!actual.has(key)) errors.push(`${actualName} missing ${key}; add route matching ${requiredName} (${route.source ?? 'source'})`);
  return errors;
}

export function validateContract({ bootstrapSource, brunoRequests, openapiSource, handlerSource, openSpecRoutes = [] }) {
  const sst = extractBootstrapRoutes(bootstrapSource).map((route) => ({ ...route, source: 'infra/stacks/bootstrap.ts' }));
  const bruno = brunoRequests.map((route) => ({ method: route.method, path: route.path, source: route.filePath }));
  const openapi = extractOpenAPIRoutes(openapiSource);
  const handler = extractHandlerRoutes(handlerSource);
  const sstMap = keys(sst);
  return { errors: [...compare(sstMap, keys(bruno), 'SST', 'Bruno'), ...compare(sstMap, keys(openapi), 'SST', 'OpenAPI'), ...compare(keys(sst.filter((route) => route.path !== '/api/health')), keys(handler), 'SST monitor routes', 'handler inventory'), ...requireRoutes(keys(openSpecRoutes), sstMap, 'merged OpenSpec', 'SST')], counts: { sst: sst.length, bruno: bruno.length, openapi: openapi.length, handler: handler.length, openSpec: openSpecRoutes.length } };
}

function main() {
  const result = validateContract({
    bootstrapSource: fs.readFileSync(path.join(root, 'infra/stacks/bootstrap.ts'), 'utf8'),
    brunoRequests: readBrunoRequests(path.join(root, '.bruno/collections')),
    openapiSource: fs.readFileSync(path.join(root, 'openapi/openapi.yaml'), 'utf8'),
    handlerSource: fs.readFileSync(path.join(root, 'services/monitor-api/routes.go'), 'utf8'),
    openSpecRoutes: extractMergedOpenSpecRoutes(path.join(root, 'openspec/specs')),
  });
  console.log(`API contract routes: SST ${result.counts.sst}, Bruno ${result.counts.bruno}, OpenAPI ${result.counts.openapi}, handler ${result.counts.handler}, OpenSpec ${result.counts.openSpec}`);
  for (const error of result.errors) console.error(`ERROR ${error}`);
  if (result.errors.length) process.exitCode = 1;
}

if (process.argv[1] === new URL(import.meta.url).pathname) main();
