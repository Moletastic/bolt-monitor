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
