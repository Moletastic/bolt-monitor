import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { extractOpenAPIRoutes } from './openapi-routes.mjs';
import { extractSSTRoutes } from './sst-routes.mjs';

const root = path.resolve(new URL('..', import.meta.url).pathname);
const requiredScope = 'aws.cognito.signin.user.admin';

export function validateOpenAPIAuth(source, bootstrapSource) {
  const errors = [];
  if (!/\nsecurity:\n  - CognitoBearer: \[\]\npaths:/.test(source)) {
    errors.push('missing global CognitoBearer security requirement for versioned operations');
  }
  if (!/CognitoBearer:\n      type: http\n      scheme: bearer\n      bearerFormat: JWT/.test(source)) {
    errors.push('missing CognitoBearer HTTP Bearer JWT security scheme');
  }
  if (!new RegExp(`x-required-cognito-scope: ${requiredScope}`).test(source)) {
    errors.push(`CognitoBearer missing ${requiredScope} required scope`);
  }
  if (!/\/api\/health:\n    get:[\s\S]*?security: \[\]/.test(source)) {
    errors.push('health must explicitly declare security: []');
  }
  for (const match of source.matchAll(/^  (\/api\/v1\/[^:]+):\n([\s\S]*?)(?=^  \/api\/|^components:)/gm)) {
    if (/security:\s*\[\]/.test(match[2])) errors.push(`${match[1]} overrides required CognitoBearer security`);
  }
  if (!/API Gateway may return a non-envelope 401 before Lambda/.test(source)) {
    errors.push('missing documented Gateway non-envelope 401 behavior');
  }
  if (!/API Gateway may return a non-envelope 403 before Lambda/.test(source)) {
    errors.push('missing documented Gateway non-envelope 403 behavior');
  }
  if (!/reason\.code `AUTHENTICATION_REQUIRED`/.test(source) || !/reason\.code `AUTHORIZATION_DENIED`/.test(source)) {
    errors.push('missing documented Lambda envelope authentication and authorization behavior');
  }
  if (bootstrapSource !== undefined) {
    const sstRoutes = new Map(extractSSTRoutes(bootstrapSource).routes.map((route) => [`${route.method} ${route.path}`, route]));
    for (const route of extractOpenAPIRoutes(source)) {
      const sstRoute = sstRoutes.get(`${route.method} ${route.path}`);
      if (sstRoute?.auth !== undefined && route.auth !== sstRoute.auth) {
        errors.push(`OpenAPI authentication conflict ${route.method} ${route.path}: SST is ${sstRoute.auth}, OpenAPI is ${route.auth ?? 'unclassified'}; update OpenAPI security to match SST (${sstRoute.source})`);
      }
    }
  }
  return errors;
}

if (process.argv[1] === new URL(import.meta.url).pathname) {
  const errors = validateOpenAPIAuth(
    fs.readFileSync(path.join(root, 'openapi/openapi.yaml'), 'utf8'),
    fs.readFileSync(path.join(root, 'infra/stacks/bootstrap.ts'), 'utf8')
  );
  for (const error of errors) console.error(`ERROR ${error}`);
  if (errors.length) process.exitCode = 1;
}
