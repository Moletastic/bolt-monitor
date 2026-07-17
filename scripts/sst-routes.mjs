import { createRoute } from './route-record.mjs';

function lineNumber(source, index) {
  return source.slice(0, index).split('\n').length;
}

function handlerForObject(source, objectName) {
  const match = source.match(new RegExp(`const ${objectName} = \\{([\\s\\S]*?)\\n  \\}`));
  return match?.[1].match(/handler:\s*['"]([^'"]+)['"]/)?.[1];
}

export function validateSSTAuthClassification(routes) {
  const classified = routes.filter((route) => route.auth !== undefined);
  if (classified.length > 0 && classified.length !== routes.length) {
    return routes.filter((route) => route.auth === undefined).map((route) => `${route.source}: route lacks public/protected authentication classification`);
  }
  return [];
}

export function extractSSTRoutes(source, filePath = 'infra/stacks/bootstrap.ts') {
  const routes = [];
  for (const match of source.matchAll(/api\.route\(\s*['"]([A-Z]+)\s+([^'"]+)['"]\s*,\s*\{([\s\S]*?)\}\s*\)/g)) {
    const routeSource = `${filePath}:${lineNumber(source, match.index)}`;
    const target = match[3].match(/handler:\s*['"]([^'"]+)['"]/)?.[1];
    routes.push(createRoute({ method: match[1], path: match[2], target, source: routeSource, auth: /\bauth\s*:/.test(match[3]) ? 'protected' : 'public' }));
  }
  for (const match of source.matchAll(/api\.route\(\s*['"]([A-Z]+)\s+([^'"]+)['"]\s*,\s*([A-Za-z][A-Za-z0-9_]*)\s*\)/g)) {
    routes.push(createRoute({
      method: match[1],
      path: match[2],
      target: handlerForObject(source, match[3]) ?? match[3],
      source: `${filePath}:${lineNumber(source, match.index)}`,
      auth: 'public',
    }));
  }
  const protectedMatch = source.match(/const protectedV1Routes = \[([\s\S]*?)\n  \]/);
  if (!protectedMatch) return { routes, diagnostics: validateSSTAuthClassification(routes) };
  const target = handlerForObject(source, 'monitorHandler') ?? 'monitorHandler';
  const registration = source.match(/api\.route\(route,\s*monitorHandler(?:,\s*\{([\s\S]*?)\}\s*)?\)/)?.[1];
  const auth = registration === undefined ? undefined : /\bauth\s*:/.test(registration) ? 'protected' : 'public';
  for (const match of protectedMatch[1].matchAll(/['"]([A-Z]+)\s+(\/api\/v1\/[^'"]+)['"]/g)) {
    const offset = protectedMatch.index + match.index;
    routes.push(createRoute({ method: match[1], path: match[2], target, source: `${filePath}:${lineNumber(source, offset)}`, auth }));
  }
  return { routes, diagnostics: validateSSTAuthClassification(routes) };
}
