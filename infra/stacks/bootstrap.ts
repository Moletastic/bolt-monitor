import { lifecyclePolicy, type DeploymentTarget } from '../deployment-target'

export function createBootstrapStack(target: DeploymentTarget) {
  const policy = lifecyclePolicy(target)
  const disposableOptions = { retainOnDelete: false }
  const durableOptions = { retainOnDelete: policy.retainDurableResources }
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

  const monitorHandler = {
    runtime: 'go' as const,
    handler: '../services/monitor-api',
    link: [table],
    environment: {
      TABLE_NAME: table.name,
    },
  }

  api.route('GET /api/v1/search', monitorHandler)
  api.route('POST /api/v1/notification-channels', monitorHandler)
  api.route('GET /api/v1/notification-channels', monitorHandler)
  api.route('GET /api/v1/notification-channels/{channelId}', monitorHandler)
  api.route('PUT /api/v1/notification-channels/{channelId}', monitorHandler)
  api.route('DELETE /api/v1/notification-channels/{channelId}', monitorHandler)
  api.route('POST /api/v1/notification-channels/{channelId}/test', monitorHandler)
  api.route('POST /api/v1/escalation-policies', monitorHandler)
  api.route('GET /api/v1/escalation-policies', monitorHandler)
  api.route('GET /api/v1/escalation-policies/{policyId}', monitorHandler)
  api.route('PUT /api/v1/escalation-policies/{policyId}', monitorHandler)
  api.route('DELETE /api/v1/escalation-policies/{policyId}', monitorHandler)
  api.route('POST /api/v1/services', monitorHandler)
  api.route('GET /api/v1/services', monitorHandler)
  api.route('GET /api/v1/services/{serviceId}', monitorHandler)
  api.route('GET /api/v1/services/{serviceId}/escalation-policy', monitorHandler)
  api.route('PATCH /api/v1/services/{serviceId}', monitorHandler)
  api.route('DELETE /api/v1/services/{serviceId}', monitorHandler)
  api.route('POST /api/v1/services/{serviceId}/archive', monitorHandler)
  api.route('POST /api/v1/services/{serviceId}/reactivate', monitorHandler)
  api.route('POST /api/v1/services/{serviceId}/monitors', monitorHandler)
  api.route('GET /api/v1/services/{serviceId}/monitors', monitorHandler)
  api.route('GET /api/v1/services/{serviceId}/monitors/{monitorId}', monitorHandler)
  api.route('PATCH /api/v1/services/{serviceId}/monitors/{monitorId}', monitorHandler)
  api.route('DELETE /api/v1/services/{serviceId}/monitors/{monitorId}', monitorHandler)
  api.route('GET /api/v1/services/{serviceId}/monitors/{monitorId}/status', monitorHandler)
  api.route('GET /api/v1/services/{serviceId}/monitors/{monitorId}/runs', monitorHandler)
  api.route('POST /api/v1/services/{serviceId}/monitors/{monitorId}/run', monitorHandler)
  api.route('GET /api/v1/services/{serviceId}/monitors/{monitorId}/incidents', monitorHandler)
  api.route('GET /api/v1/services/{serviceId}/incidents', monitorHandler)
  api.route('GET /api/v1/services/{serviceId}/monitors/{monitorId}/audit', monitorHandler)
  api.route('GET /api/v1/services/{serviceId}/audit', monitorHandler)
  api.route('POST /api/v1/services/{serviceId}/monitors/{monitorId}/enable', monitorHandler)
  api.route('POST /api/v1/services/{serviceId}/monitors/{monitorId}/disable', monitorHandler)
  api.route(
    'POST /api/v1/services/{serviceId}/monitors/{monitorId}/maintenance/enable',
    monitorHandler
  )
  api.route(
    'POST /api/v1/services/{serviceId}/monitors/{monitorId}/maintenance/disable',
    monitorHandler
  )
  api.route('GET /api/v1/incidents', monitorHandler)
  api.route('GET /api/v1/incidents/{incidentId}', monitorHandler)
  api.route('GET /api/v1/incidents/{incidentId}/escalation-state', monitorHandler)
  api.route('GET /api/v1/incidents/{incidentId}/activities', monitorHandler)
  api.route('POST /api/v1/incidents/{incidentId}/ack', monitorHandler)
  api.route('POST /api/v1/incidents/{incidentId}/resolve', monitorHandler)
  api.route('GET /api/v1/admin/scheduler-config', monitorHandler)
  api.route('PATCH /api/v1/admin/scheduler-config', monitorHandler)

  const dashboard = new sst.aws.Nextjs(
    'Dashboard',
    {
      path: '../apps/dashboard',
      environment: {
        NEXT_PUBLIC_MONITOR_API_BASE_URL: api.url,
      },
    },
    disposableOptions
  )

  return {
    apiUrl: api.url,
    dashboardUrl: dashboard.url,
    appTableName: table.name,
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
    },
    retainedResourceInventory: policy.retainDurableResources
      ? {
          version: 'v1',
          tables: [{ logicalName: 'AppTable', name: table.name, arn: table.arn }],
          identity: [],
          parametersAndSecrets: [],
        }
      : { version: 'v1', tables: [], identity: [], parametersAndSecrets: [] },
  }
}
