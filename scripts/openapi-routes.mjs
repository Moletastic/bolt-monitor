import { createRoute } from './route-record.mjs';

const methods = /\b(get|post|put|patch|delete):/gi;

function securityFor(operation, globalSecurity) {
  if (/\bsecurity:\s*\[\s*\]/.test(operation)) return 'public';
  if (/\bsecurity:\s*\[/.test(operation) || /\bsecurity:\s*\n\s*-\s+\S/.test(operation)) return 'protected';
  return globalSecurity;
}

export function extractOpenAPIRoutes(source, filePath = 'openapi/openapi.yaml') {
  const lines = source.split('\n');
  const globalSecurity = /^security:\n(?:  -\s+\S.*\n?)+/m.test(source) ? 'protected' : undefined;
  const paths = [];
  for (const [index, line] of lines.entries()) {
    const match = line.match(/^  (\/api\/[^:]+):/);
    if (match) paths.push({ path: match[1], line: index, inline: line.slice(match[0].length) });
  }
  const routes = [];
  for (const [pathIndex, pathEntry] of paths.entries()) {
    const end = paths[pathIndex + 1]?.line ?? lines.length;
    for (const match of pathEntry.inline.matchAll(methods)) {
      routes.push(createRoute({ method: match[1], path: pathEntry.path, source: `${filePath}:${pathEntry.line + 1}`, auth: securityFor(pathEntry.inline, globalSecurity) }));
    }
    for (let index = pathEntry.line + 1; index < end; index += 1) {
      const method = lines[index].match(/^    (get|post|put|patch|delete):/i);
      if (!method) continue;
      const operationEnd = lines.slice(index + 1, end).findIndex((line) => /^    (get|post|put|patch|delete):|^  \/api\/|^\S/.test(line));
      const operation = lines.slice(index, operationEnd < 0 ? end : index + 1 + operationEnd).join('\n');
      routes.push(createRoute({ method: method[1], path: pathEntry.path, source: `${filePath}:${index + 1}`, auth: securityFor(operation, globalSecurity) }));
    }
  }
  return routes;
}
