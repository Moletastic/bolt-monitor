const routeMethods = new Set(['GET', 'POST', 'PUT', 'PATCH', 'DELETE']);

export function normalizeRoutePath(value) {
  return value
    .replace(/^\{\{apiUrl\}\}/, '')
    .split('?', 1)[0]
    .replace(/\{\{([A-Za-z][A-Za-z0-9_]*)\}\}/g, '{$1}');
}

export function createRoute({ method, path, target, source, auth }) {
  const normalizedMethod = method.toUpperCase();
  if (!routeMethods.has(normalizedMethod)) throw new Error(`unsupported HTTP method: ${method}`);

  const route = { method: normalizedMethod, path: normalizeRoutePath(path) };
  if (target !== undefined) route.target = target;
  if (source !== undefined) route.source = source;
  if (auth !== undefined) route.auth = auth;
  return route;
}

export function routeKey(route) {
  return `${route.method} ${route.path}`;
}
