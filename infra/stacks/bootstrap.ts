import { lifecyclePolicy, type DeploymentTarget } from '../deployment-target'

export function createBootstrapStack(target: DeploymentTarget) {
  const policy = lifecyclePolicy(target)
  const disposableOptions = { retainOnDelete: false }
  const durableOptions = { retainOnDelete: policy.retainDurableResources }
  const authKeyParameterName = `/${target.service}/${target.stage}/auth/aes-256-gcm`
  const authEncryptionKey = aws.ssm.Parameter.get('AuthEncryptionKey', authKeyParameterName)

  const operatorUserPool = new aws.cognito.UserPool(
    'OperatorUserPool',
    {
      name: `${target.service}-${target.stage}-operators`,
      userPoolTier: 'ESSENTIALS',
      usernameAttributes: ['email'],
      autoVerifiedAttributes: ['email'],
      adminCreateUserConfig: { allowAdminCreateUserOnly: true },
      emailConfiguration: { emailSendingAccount: 'COGNITO_DEFAULT' },
      accountRecoverySetting: {
        recoveryMechanisms: [{ name: 'verified_email', priority: 1 }],
      },
      passwordPolicy: {
        minimumLength: 12,
        requireLowercase: true,
        requireNumbers: true,
        requireSymbols: true,
        requireUppercase: true,
        temporaryPasswordValidityDays: 7,
      },
      mfaConfiguration: 'OPTIONAL',
      softwareTokenMfaConfiguration: { enabled: true },
      deletionProtection: policy.retainDurableResources ? 'ACTIVE' : 'INACTIVE',
      tags: policy.tags,
    },
    durableOptions
  )

  const userPoolClientArgs = {
    userPoolId: operatorUserPool.id,
    explicitAuthFlows: ['ALLOW_USER_PASSWORD_AUTH', 'ALLOW_REFRESH_TOKEN_AUTH'],
    enableTokenRevocation: true,
    refreshTokenRotation: { feature: 'ENABLED', retryGracePeriodSeconds: 10 },
    accessTokenValidity: 60,
    idTokenValidity: 60,
    refreshTokenValidity: 12,
    tokenValidityUnits: {
      accessToken: 'minutes',
      idToken: 'minutes',
      refreshToken: 'hours',
    },
    preventUserExistenceErrors: 'ENABLED',
  }

  const dashboardUserPoolClient = new aws.cognito.UserPoolClient(
    'DashboardUserPoolClient',
    {
      name: `${target.service}-${target.stage}-dashboard`,
      generateSecret: true,
      ...userPoolClientArgs,
    },
    durableOptions
  )
  const directOperatorUserPoolClient = new aws.cognito.UserPoolClient(
    'DirectOperatorUserPoolClient',
    {
      name: `${target.service}-${target.stage}-operator`,
      generateSecret: false,
      ...userPoolClientArgs,
    },
    durableOptions
  )

  const authTable = new aws.dynamodb.Table(
    'AuthTable',
    {
      name: `${target.service}-${target.stage}-auth`,
      billingMode: 'PAY_PER_REQUEST',
      hashKey: 'PK',
      rangeKey: 'SK',
      attributes: [
        { name: 'PK', type: 'S' },
        { name: 'SK', type: 'S' },
      ],
      ttl: { attributeName: 'TTL', enabled: true },
      deletionProtectionEnabled: policy.retainDurableResources,
      pointInTimeRecovery: { enabled: policy.retainDurableResources },
      tags: policy.tags,
    },
    durableOptions
  )

  const bootstrapBucket = new sst.aws.Bucket(
    'BootstrapBucket',
    {
      lifecycle: [{ id: 'expire-objects', enabled: true, expiresIn: '30 days' }],
    },
    disposableOptions
  )
  const table = new sst.aws.Dynamo(
    'AppTable',
    {
      fields: {
        PK: 'string',
        SK: 'string',
        GSI1PK: 'string',
        GSI1SK: 'string',
        GSI2PK: 'string',
        GSI2SK: 'string',
        GSI3PK: 'string',
        GSI3SK: 'string',
      },
      primaryIndex: {
        hashKey: 'PK',
        rangeKey: 'SK',
      },
      globalIndexes: {
        OpenIncidentsIndex: {
          hashKey: 'GSI1PK',
          rangeKey: 'GSI1SK',
        },
        StatusByTenantIndex: {
          hashKey: 'GSI2PK',
          rangeKey: 'GSI2SK',
        },
        AuditByResourceIndex: {
          hashKey: 'GSI3PK',
          rangeKey: 'GSI3SK',
        },
      },
      ttl: 'TTL',
      deletionProtection: policy.retainDurableResources,
      transform: {
        table: (args) => {
          args.tags = { ...args.tags, ...policy.tags }
        },
      },
    },
    durableOptions
  )

  const executionQueueDLQ = new sst.aws.Queue(
    'ExecutionQueueDLQ',
    {
      fifo: false,
    },
    disposableOptions
  )

  const executionQueue = new sst.aws.Queue(
    'ExecutionQueue',
    {
      fifo: false,
      dlq: {
        queue: executionQueueDLQ.arn,
        retry: 3,
      },
    },
    disposableOptions
  )

  const notificationQueueDLQ = new sst.aws.Queue(
    'NotificationQueueDLQ',
    {
      fifo: false,
    },
    disposableOptions
  )

  const notificationQueue = new sst.aws.Queue(
    'NotificationQueue',
    {
      fifo: false,
      dlq: {
        queue: notificationQueueDLQ.arn,
        retry: 3,
      },
    },
    disposableOptions
  )

  new sst.aws.Cron(
    'SchedulerSchedule',
    {
      schedule: 'rate(1 minute)',
      function: {
        runtime: 'go',
        handler: '../services/check-runtime',
        link: [table, executionQueue],
        environment: {
          TABLE_NAME: table.name,
          RUNTIME_MODE: 'scheduler',
          EXECUTION_QUEUE_URL: executionQueue.url,
          ESCALATION_QUEUE_URL: notificationQueue.url,
        },
      },
    },
    disposableOptions
  )

  executionQueue.subscribe(
    {
      runtime: 'go',
      handler: '../services/check-runtime',
      link: [table],
      permissions: [
        {
          actions: ['sqs:SendMessage'],
          resources: [notificationQueue.arn],
        },
      ],
      environment: {
        TABLE_NAME: table.name,
        RUNTIME_MODE: 'worker',
        ESCALATION_QUEUE_URL: notificationQueue.url,
      },
    },
    {
      batch: {
        size: 1,
      },
    }
  )

  notificationQueue.subscribe(
    {
      runtime: 'go',
      handler: '../services/escalation-runtime',
      link: [table],
      environment: {
        TABLE_NAME: table.name,
      },
    },
    {
      batch: {
        size: 1,
      },
    }
  )

  const api = new sst.aws.ApiGatewayV2(
    'Api',
    { accessLog: { retention: '2 weeks' } },
    disposableOptions
  )

  api.route('GET /api/health', {
    runtime: 'go',
    handler: '../services/api-health',
  })

  const operatorAuthorizer = api.addAuthorizer({
    name: 'OperatorJwt',
    jwt: {
      issuer: $interpolate`https://cognito-idp.${aws.getRegionOutput().region}.amazonaws.com/${operatorUserPool.id}`,
      audiences: [dashboardUserPoolClient.id, directOperatorUserPoolClient.id],
    },
  })

  const monitorHandler = {
    runtime: 'go' as const,
    handler: '../services/monitor-api',
    link: [table],
    environment: {
      TABLE_NAME: table.name,
    },
  }

  const protectedV1Routes = [
    'GET /api/v1/search',
    'POST /api/v1/notification-channels',
    'GET /api/v1/notification-channels',
    'GET /api/v1/notification-channels/{channelId}',
    'PUT /api/v1/notification-channels/{channelId}',
    'DELETE /api/v1/notification-channels/{channelId}',
    'POST /api/v1/notification-channels/{channelId}/test',
    'POST /api/v1/escalation-policies',
    'GET /api/v1/escalation-policies',
    'GET /api/v1/escalation-policies/{policyId}',
    'PUT /api/v1/escalation-policies/{policyId}',
    'DELETE /api/v1/escalation-policies/{policyId}',
    'POST /api/v1/services',
    'GET /api/v1/services',
    'GET /api/v1/services/{serviceId}',
    'GET /api/v1/services/{serviceId}/escalation-policy',
    'PATCH /api/v1/services/{serviceId}',
    'DELETE /api/v1/services/{serviceId}',
    'POST /api/v1/services/{serviceId}/archive',
    'POST /api/v1/services/{serviceId}/reactivate',
    'POST /api/v1/services/{serviceId}/monitors',
    'GET /api/v1/services/{serviceId}/monitors',
    'GET /api/v1/services/{serviceId}/monitors/{monitorId}',
    'PATCH /api/v1/services/{serviceId}/monitors/{monitorId}',
    'DELETE /api/v1/services/{serviceId}/monitors/{monitorId}',
    'GET /api/v1/services/{serviceId}/monitors/{monitorId}/status',
    'GET /api/v1/services/{serviceId}/monitors/{monitorId}/runs',
    'POST /api/v1/services/{serviceId}/monitors/{monitorId}/run',
    'GET /api/v1/services/{serviceId}/monitors/{monitorId}/incidents',
    'GET /api/v1/services/{serviceId}/incidents',
    'GET /api/v1/services/{serviceId}/monitors/{monitorId}/audit',
    'GET /api/v1/services/{serviceId}/audit',
    'POST /api/v1/services/{serviceId}/monitors/{monitorId}/enable',
    'POST /api/v1/services/{serviceId}/monitors/{monitorId}/disable',
    'POST /api/v1/services/{serviceId}/monitors/{monitorId}/maintenance/enable',
    'POST /api/v1/services/{serviceId}/monitors/{monitorId}/maintenance/disable',
    'GET /api/v1/incidents',
    'GET /api/v1/incidents/{incidentId}',
    'GET /api/v1/incidents/{incidentId}/escalation-state',
    'GET /api/v1/incidents/{incidentId}/activities',
    'POST /api/v1/incidents/{incidentId}/ack',
    'POST /api/v1/incidents/{incidentId}/resolve',
    'GET /api/v1/admin/scheduler-config',
    'PATCH /api/v1/admin/scheduler-config',
  ]
  for (const route of protectedV1Routes)
    api.route(route, monitorHandler, {
      auth: {
        jwt: { authorizer: operatorAuthorizer.id, scopes: ['aws.cognito.signin.user.admin'] },
      },
    })

  const dashboard = new sst.aws.Nextjs(
    'Dashboard',
    {
      path: '../apps/dashboard',
      environment: {
        NEXT_PUBLIC_MONITOR_API_BASE_URL: api.url,
        DASHBOARD_ORIGIN: target.dashboardOrigin,
      },
    },
    disposableOptions
  )

  return {
    apiUrl: api.url,
    dashboardUrl: dashboard.url,
    appTableName: table.name,
    authTableName: authTable.name,
    operatorUserPoolId: operatorUserPool.id,
    dashboardUserPoolClientId: dashboardUserPoolClient.id,
    directOperatorUserPoolClientId: directOperatorUserPoolClient.id,
    authEncryptionKeyParameterName: authEncryptionKey.name,
    bootstrapBucket: bootstrapBucket.name,
    notificationQueueUrl: notificationQueue.url,
    lifecycleClass: target.lifecycle,
    deploymentIdentity: {
      service: target.service,
      stage: target.stage,
      owner: target.owner,
      accountId: target.accountId,
      region: target.region,
      credentialSource: target.credentialSource,
      dashboardOrigin: target.dashboardOrigin,
    },
    retainedResourceInventory: policy.retainDurableResources
      ? {
          version: 'v1',
          tables: [{ logicalName: 'AppTable', name: table.name, arn: table.arn }],
          identity: [
            { logicalName: 'OperatorUserPool', id: operatorUserPool.id, arn: operatorUserPool.arn },
          ],
          authTables: [{ logicalName: 'AuthTable', name: authTable.name, arn: authTable.arn }],
          parametersAndSecrets: [],
        }
      : { version: 'v1', tables: [], authTables: [], identity: [], parametersAndSecrets: [] },
  }
}
