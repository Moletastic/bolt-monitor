export function createBootstrapStack() {
  const bootstrapBucket = new sst.aws.Bucket('BootstrapBucket')
  const table = new sst.aws.Dynamo('AppTable', {
    fields: {
      PK: 'string',
      SK: 'string',
      GSI1PK: 'string',
      GSI1SK: 'string',
      GSI2PK: 'string',
      GSI2SK: 'string',
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
    },
    ttl: 'TTL',
  })

  const executionQueueDLQ = new sst.aws.Queue('ExecutionQueueDLQ', {
    fifo: false,
  })

  const executionQueue = new sst.aws.Queue('ExecutionQueue', {
    fifo: false,
    dlq: {
      queue: executionQueueDLQ.arn,
      retry: 3,
    },
  })

  const notificationQueueDLQ = new sst.aws.Queue('NotificationQueueDLQ', {
    fifo: false,
  })

  const notificationQueue = new sst.aws.Queue('NotificationQueue', {
    fifo: false,
    dlq: {
      queue: notificationQueueDLQ.arn,
      retry: 3,
    },
  })

  new sst.aws.Cron('SchedulerSchedule', {
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
  })

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

  const api = new sst.aws.ApiGatewayV2('Api')

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

  api.route('GET /api/v1/probe-locations', monitorHandler)
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
  api.route('GET /api/v1/incidents', monitorHandler)
  api.route('GET /api/v1/incidents/{incidentId}', monitorHandler)
  api.route('GET /api/v1/incidents/{incidentId}/activities', monitorHandler)
  api.route('POST /api/v1/incidents/{incidentId}/ack', monitorHandler)
  api.route('POST /api/v1/incidents/{incidentId}/resolve', monitorHandler)
  api.route('GET /api/v1/admin/scheduler-config', monitorHandler)
  api.route('PATCH /api/v1/admin/scheduler-config', monitorHandler)

  const dashboard = new sst.aws.Nextjs('Dashboard', {
    path: '../apps/dashboard',
    environment: {
      NEXT_PUBLIC_MONITOR_API_BASE_URL: api.url,
    },
  })

  return {
    apiUrl: api.url,
    dashboardUrl: dashboard.url,
    appTableName: table.name,
    bootstrapBucket: bootstrapBucket.name,
    notificationQueueUrl: notificationQueue.url,
  }
}
