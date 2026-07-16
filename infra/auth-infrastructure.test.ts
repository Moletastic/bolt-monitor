import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const stackSource = readFileSync(
  fileURLToPath(new URL('./stacks/bootstrap.ts', import.meta.url)),
  'utf8'
)

function stackSection(resourceName: string) {
  const start = stackSource.indexOf(`'${resourceName}'`)
  assert.notEqual(start, -1, `${resourceName} resource must be declared`)
  const end = stackSource.indexOf('\n  )', start)
  assert.notEqual(end, -1, `${resourceName} resource declaration must end`)
  return stackSource.slice(start, end)
}

test('operator Cognito remains invite-only with custom authentication settings', () => {
  const userPool = stackSection('OperatorUserPool')
  const clients = `${stackSection('DashboardUserPoolClient')}\n${stackSection(
    'DirectOperatorUserPoolClient'
  )}`

  assert.match(userPool, /userPoolTier: 'ESSENTIALS'/)
  assert.match(userPool, /adminCreateUserConfig: \{ allowAdminCreateUserOnly: true \}/)
  assert.match(userPool, /emailConfiguration: \{ emailSendingAccount: 'COGNITO_DEFAULT' \}/)
  assert.match(userPool, /softwareTokenMfaConfiguration: \{ enabled: true \}/)
  assert.match(userPool, /tags: policy\.tags/)
  assert.match(
    stackSource,
    /refreshTokenRotation: \{ feature: 'ENABLED', retryGracePeriodSeconds: 10 \}/
  )
  assert.match(clients, /\.\.\.userPoolClientArgs/)
  assert.doesNotMatch(stackSource, /UserPoolDomain|managedLogin|Amplify/)
})

test('AuthTable has auth-specific lifecycle protection and required tags', () => {
  const authTable = stackSection('AuthTable')

  assert.match(authTable, /billingMode: 'PAY_PER_REQUEST'/)
  assert.match(authTable, /hashKey: 'PK'/)
  assert.match(authTable, /rangeKey: 'SK'/)
  assert.match(authTable, /ttl: \{ attributeName: 'TTL', enabled: true \}/)
  assert.match(authTable, /deletionProtectionEnabled: policy\.retainDurableResources/)
  assert.match(authTable, /pointInTimeRecovery: \{ enabled: policy\.retainDurableResources \}/)
  assert.match(authTable, /tags: policy\.tags/)
  assert.match(authTable, /durableOptions/)

  assert.match(
    stackSource,
    /const durableOptions = \{ retainOnDelete: policy\.retainDurableResources \}/
  )
  assert.match(stackSource, /const disposableOptions = \{ retainOnDelete: false \}/)
  assert.match(stackSource, /authTableName: authTable\.name/)
  assert.doesNotMatch(stackSource, /authTable[^\n]*AppTable|AppTable[^\n]*authTable/)
})

test('auth encryption uses only a non-secret reference and no customer-managed key', () => {
  assert.match(stackSource, /aws\.ssm\.Parameter\.get\('AuthEncryptionKey', authKeyParameterName\)/)
  assert.match(stackSource, /authEncryptionKeyParameterName: authEncryptionKey\.name/)
  assert.doesNotMatch(stackSource, /authEncryptionKey\.(value|valueString|secretValue)/)
  assert.doesNotMatch(stackSource, /aws\.kms|kms\.Key|customer.?managed/i)
})

test('auth configuration and permissions are scoped to monitor API and dashboard server', () => {
  const monitorHandler = stackSource.slice(
    stackSource.indexOf('const monitorHandler'),
    stackSource.indexOf('const protectedV1Routes')
  )
  const dashboard = stackSource.slice(
    stackSource.indexOf("'Dashboard'"),
    stackSource.indexOf('\n  return {')
  )
  const outputs = stackSource.slice(stackSource.indexOf('\n  return {'))

  assert.match(monitorHandler, /AUTH_TABLE_NAME: authTable\.name/)
  assert.match(
    monitorHandler,
    /COGNITO_CLIENT_IDS: \$interpolate`\$\{dashboardUserPoolClient\.id\},\$\{directOperatorUserPoolClient\.id\}`/
  )
  assert.match(monitorHandler, /actions: \['dynamodb:GetItem'\]/)
  assert.match(monitorHandler, /resources: \[authTable\.arn\]/)

  assert.match(dashboard, /DASHBOARD_ORIGIN: target\.dashboardOrigin/)
  assert.match(dashboard, /AUTH_STAGE: target\.stage/)
  assert.match(dashboard, /AUTH_TABLE_NAME: authTable\.name/)
  assert.match(dashboard, /COGNITO_USER_POOL_ID: operatorUserPool\.id/)
  assert.match(dashboard, /COGNITO_DASHBOARD_CLIENT_ID: dashboardUserPoolClient\.id/)
  assert.match(dashboard, /COGNITO_DASHBOARD_CLIENT_SECRET: dashboardUserPoolClient\.clientSecret/)
  assert.match(dashboard, /AUTH_ENCRYPTION_KEY_PARAMETER_NAME: authEncryptionKey\.name/)
  assert.match(dashboard, /actions: \['ssm:GetParameter'\]/)
  assert.doesNotMatch(dashboard, /NEXT_PUBLIC_(?:AUTH|COGNITO|DASHBOARD_ORIGIN)/)
  assert.doesNotMatch(dashboard, /NEXT_PUBLIC_COGNITO_DASHBOARD_CLIENT_SECRET/)
  assert.doesNotMatch(dashboard, /authEncryptionKey\.(?:value|valueString|secretValue)/)

  assert.match(outputs, /authTableName: authTable\.name/)
  assert.match(outputs, /operatorUserPoolId: operatorUserPool\.id/)
  assert.match(outputs, /authEncryptionKeyParameterName: authEncryptionKey\.name/)
  assert.match(
    outputs,
    /authTables: \[\{ logicalName: 'AuthTable', name: authTable\.name, arn: authTable\.arn \}\]/
  )
  assert.match(
    outputs,
    /logicalName: 'OperatorUserPool', id: operatorUserPool\.id, arn: operatorUserPool\.arn/
  )
  assert.match(outputs, /logicalName: 'AuthEncryptionKey', name: authEncryptionKey\.name/)
  assert.doesNotMatch(outputs, /dashboardUserPoolClientId|directOperatorUserPoolClientId/)
})

test('Gateway rejects protected-route requests before Lambda and keeps health public', () => {
  const healthRoute = stackSource.slice(
    stackSource.indexOf("api.route('GET /api/health'"),
    stackSource.indexOf('const operatorAuthorizer')
  )
  const protectedRoutes = stackSource.slice(
    stackSource.indexOf('const protectedV1Routes'),
    stackSource.indexOf('const dashboard = new sst.aws.Nextjs')
  )

  assert.match(healthRoute, /handler: '\.\.\/services\/api-health'/)
  assert.doesNotMatch(healthRoute, /auth:\s*\{/)
  assert.match(protectedRoutes, /api\.route\(route, monitorHandler, \{\s*auth:/)
  assert.match(protectedRoutes, /authorizer: operatorAuthorizer\.id/)
  assert.match(protectedRoutes, /scopes: \['aws\.cognito\.signin\.user\.admin'\]/)
  assert.match(
    stackSource,
    /audiences: \[dashboardUserPoolClient\.id, directOperatorUserPoolClient\.id\]/
  )
})

test('static deployment configuration keeps the auth boundary complete and dependency-free', () => {
  const protectedRoutes = stackSource.slice(
    stackSource.indexOf('const protectedV1Routes'),
    stackSource.indexOf('const dashboard = new sst.aws.Nextjs')
  )
  const routes = [...protectedRoutes.matchAll(/'([A-Z]+ \/api\/v1\/[^']+)'/g)].map(
    ([, route]) => route
  )

  assert.ok(routes.length > 0, 'at least one protected v1 route must be declared')
  assert.ok(routes.every((route) => route.includes(' /api/v1/')))
  assert.match(
    protectedRoutes,
    /for \(const route of protectedV1Routes\)\s+api\.route\(route, monitorHandler, \{\s+auth: \{\s+jwt: \{ authorizer: operatorAuthorizer\.id, scopes: \['aws\.cognito\.signin\.user\.admin'\] \}/
  )

  assert.doesNotMatch(
    stackSource,
    /Amplify|UserPoolDomain|managedLogin|aws\.ses|aws\.route53|customEmailSender/i
  )
  assert.doesNotMatch(stackSource, /new aws\.ssm\.Parameter\(/)
  assert.doesNotMatch(stackSource, /authEncryptionKey\.(?:value|valueString|secretValue)/)

  const userPool = stackSection('OperatorUserPool')
  const authTable = stackSection('AuthTable')
  const clients = `${stackSection('DashboardUserPoolClient')}\n${stackSection(
    'DirectOperatorUserPoolClient'
  )}`
  for (const resource of [userPool, authTable, clients]) {
    assert.match(resource, /durableOptions/)
  }
  assert.match(stackSource, /deletionProtection: policy\.retainDurableResources/)
  assert.match(authTable, /pointInTimeRecovery: \{ enabled: policy\.retainDurableResources \}/)
  assert.match(stackSource, /const disposableOptions = \{ retainOnDelete: false \}/)
})

test('auth logs and alarms have finite retention, bounded metrics, tags, and thresholds', () => {
  const monitorHandler = stackSource.slice(
    stackSource.indexOf('const monitorHandler'),
    stackSource.indexOf('const protectedV1Routes')
  )
  const dashboard = stackSource.slice(
    stackSource.indexOf("'Dashboard'"),
    stackSource.indexOf('\n  return {')
  )

  assert.match(stackSource, /const authLogRetention = '2 weeks'/)
  assert.match(monitorHandler, /logging: \{ retention: authLogRetention \}/)
  assert.match(
    dashboard,
    /transform: \{ server: \{ logging: \{ retention: authLogRetention \} \} \}/
  )
  assert.match(stackSource, /const authMetricNamespace = 'BoltMonitor\/Auth'/)
  assert.match(stackSource, /const authMetricName = 'AuthenticationEvents'/)
  assert.match(stackSource, /const authAlarmPeriodSeconds = 300/)
  assert.match(stackSource, /const authAlarmEvaluationPeriods = 3/)
  assert.match(stackSource, /const authRefreshFailureThreshold = 5/)
  assert.match(stackSource, /const authInfrastructureErrorThreshold = 3/)
  assert.match(stackSource, /new aws\.cloudwatch\.MetricAlarm\('AuthRefreshFailureAlarm'/)
  assert.match(stackSource, /\['AuthStorageFailureAlarm', 'storage'\]/)
  assert.match(stackSource, /\['AuthKeyLoadingFailureAlarm', 'key_loading'\]/)
  assert.match(stackSource, /datapointsToAlarm: authAlarmEvaluationPeriods/)
  assert.match(stackSource, /treatMissingData: 'notBreaching'/)
  assert.match(stackSource, /tags: policy\.tags/)
  assert.match(
    stackSource,
    /dimensions: \{\s+stage: target\.stage,\s+component: 'dashboard-auth',\s+operation: 'refresh',\s+outcome: 'failure',\s+\}/
  )
  assert.match(
    stackSource,
    /dimensions: \{\s+stage: target\.stage,\s+component: 'dashboard-auth',\s+operation,\s+outcome: 'failure',\s+\}/
  )
})
