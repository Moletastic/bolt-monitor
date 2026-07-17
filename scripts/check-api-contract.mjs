import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { readBrunoRequests } from './check-bruno.mjs';
import { normalizeRoutePath, routeKey } from './route-record.mjs';
import { extractOpenAPIRoutes as extractOperations } from './openapi-routes.mjs';
import { extractSSTRoutes } from './sst-routes.mjs';
import { extractMergedOpenSpecRoutes } from './openspec-routes.mjs';
import { extractHandlerRouteInventory } from './handler-routes.mjs';

const root = path.resolve(new URL('..', import.meta.url).pathname);

export { normalizeRoutePath as normalizePath, extractMergedOpenSpecRoutes };
export function extractOpenAPIRoutes(source, filePath) { return extractOperations(source, filePath); }
export function extractHandlerRoutes(source, filePath = 'services/monitor-api/routes.go') {
  return extractHandlerRouteInventory(source, filePath).routes;
}

function keys(routes) { return new Map(routes.map((route) => [routeKey(route), route])); }
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

function compareAuth(expected, actual, actualName) {
  const errors = [];
  const expectedByKey = keys(expected);
  for (const route of actual) {
    const sstRoute = expectedByKey.get(routeKey(route));
    if (sstRoute?.auth !== undefined && route.auth !== sstRoute.auth) {
      errors.push(`${actualName} authentication conflict ${routeKey(route)}: SST is ${sstRoute.auth}, ${actualName} is ${route.auth ?? 'unclassified'}; update ${actualName} metadata to match SST (${sstRoute.source})`);
    }
  }
  return errors;
}

export function validateContract({ bootstrapSource, brunoRequests, openapiSource, handlerSource, openSpecRoutes = [] }) {
  const extractedSst = extractSSTRoutes(bootstrapSource);
  const sst = extractedSst.routes;
  const bruno = brunoRequests.filter((route) => route.external === undefined).map((route) => ({ method: route.method, path: route.path, source: route.filePath, auth: route.auth === 'access-token' ? 'protected' : route.auth === 'none' ? 'public' : undefined }));
  const openapi = extractOperations(openapiSource);
  const extractedHandler = extractHandlerRouteInventory(handlerSource);
  const handler = extractedHandler.routes;
  const sstMap = keys(sst);
  return { errors: [...extractedSst.diagnostics, ...extractedHandler.diagnostics, ...compare(sstMap, keys(bruno), 'SST', 'Bruno'), ...compare(sstMap, keys(openapi), 'SST', 'OpenAPI'), ...compareAuth(sst, bruno, 'Bruno'), ...compareAuth(sst, openapi, 'OpenAPI'), ...compare(keys(sst.filter((route) => route.target === '../services/monitor-api')), keys(handler), 'SST monitor routes', 'handler inventory'), ...requireRoutes(keys(openSpecRoutes), sstMap, 'merged OpenSpec', 'SST')], counts: { sst: sst.length, bruno: bruno.length, openapi: openapi.length, handler: handler.length, openSpec: openSpecRoutes.length } };
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
