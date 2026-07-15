import fs from 'node:fs'
import path from 'node:path'
import process from 'node:process'
import { extractBootstrapRoutes } from './check-bruno.mjs'

const root = path.resolve(new URL('..', import.meta.url).pathname)
const requiredScope = 'aws.cognito.signin.user.admin'

export function validateAuthRoutes(source) {
  const routes = extractBootstrapRoutes(source)
  const errors = []
  const v1Routes = routes.filter((route) => route.path.startsWith('/api/v1/'))
  if (v1Routes.length === 0) errors.push('no protected v1 routes declared')
  if (!source.includes('const operatorAuthorizer = api.addAuthorizer')) errors.push('missing operator JWT authorizer')
  if (!source.includes("authorizer: operatorAuthorizer.id")) errors.push('v1 routes do not use operator JWT authorizer')
  if (!source.includes(`scopes: ['${requiredScope}']`)) errors.push(`v1 routes missing ${requiredScope} scope`)
  if (!routes.some((route) => route.method === 'GET' && route.path === '/api/health')) errors.push('missing public health route')
  return errors
}

if (process.argv[1] === new URL(import.meta.url).pathname) {
  const errors = validateAuthRoutes(fs.readFileSync(path.join(root, 'infra/stacks/bootstrap.ts'), 'utf8'))
  for (const error of errors) console.error(`ERROR ${error}`)
  if (errors.length) process.exitCode = 1
}
