import assert from 'node:assert/strict'
import test from 'node:test'
import { validateAuthRoutes } from './check-auth-routes.mjs'

test('requires shared JWT authorizer and scope for v1 routes', () => {
  const source = "api.route('GET /api/health', health)\nconst operatorAuthorizer = api.addAuthorizer({})\nconst protectedV1Routes = [\n  'GET /api/v1/items',\n  ]\napi.route(route, handler, { auth: { jwt: { authorizer: operatorAuthorizer.id, scopes: ['aws.cognito.signin.user.admin'] } } })"
  assert.deepEqual(validateAuthRoutes(source), [])
  assert.match(validateAuthRoutes(source.replace("scopes: ['aws.cognito.signin.user.admin']", 'scopes: []')).join('\n'), /missing aws.cognito.signin.user.admin scope/)
})
