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
  assert.doesNotMatch(outputs, /dashboardUserPoolClientId|directOperatorUserPoolClientId/)
})
