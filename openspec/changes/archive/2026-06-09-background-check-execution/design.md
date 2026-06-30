## Context

The manual check execution change allows operators to trigger checks via API and get immediate results. Now we need automated background execution so monitors run on their configured schedules (e.g., every 60 seconds) without manual triggers.

This builds on the existing `check-runtime` code which already has scheduler and worker modes, but lacks:
1. EventBridge to trigger the scheduler
2. SQS queue for work distribution
3. SQS trigger for the worker
4. SQS client integration

Current flow (broken):
```
EventBridge (not configured) → check-runtime scheduler → DynamoDB RUN_REQUEST# → (no one processes)
```

Target flow (after this change):
```
EventBridge (rate 30s) → check-runtime scheduler → SQS Queue → check-runtime worker → Target URL
                          ↓                                                      ↓
                     DynamoDB (audit)                                    DynamoDB (results)
```

## Goals / Non-Goals

**Goals:**
- Automate monitor execution on configured schedules
- Use SQS for event-driven, reliable work distribution
- Maintain DynamoDB audit trail (RUN_REQUEST# items still written)
- Support DLQ for failed message handling
- Preserve the existing manual run capability (executes inline, not via SQS)
- Allow multiple workers to process in parallel (SQS fan-out)

**Non-Goals:**
- Multi-region or multi-probe-location execution (single "iad" location for now)
- SQS FIFO queues (standard queue is sufficient)
- Replacing DynamoDB writes with SQS-only (keep audit trail)
- Changing monitor-api Lambda (manual runs stay inline)

## Decisions

### Decision 1: Standard SQS queue (not FIFO)

**Choice:** Use standard SQS queue for execution work.

**Rationale:** Standard queue supports unlimited throughput, good for our scale. FIFO would add ordering guarantees we don't need (each execution is independent). Later we could migrate to FIFO per probe location if ordering becomes important.

### Decision 2: One message per (monitor × probeLocation)

**Choice:** Scheduler sends one SQS message per monitor per probe location.

**Rationale:** Simple, matches current behavior. Each message contains full ExecutionRequest as JSON. Worker processes one message at a time. If a monitor has 3 probe locations, we send 3 messages.

### Decision 3: DynamoDB audit trail kept (not replaced)

**Choice:** Scheduler writes both to DynamoDB (RUN_REQUEST#) and SQS.

**Rationale:** DynamoDB provides historical audit trail of what was enqueued. Even if SQS message is processed, we have record of the enqueue event. This is useful for debugging and compliance.

### Decision 4: SQS visibility timeout = 60 seconds

**Choice:** Set SQS queue visibility timeout to 60 seconds.

**Rationale:** HTTP checks have configurable timeout (default 5000ms). Visibility timeout should be greater than max expected execution time. 60 seconds provides buffer for slow targets and Lambda cold starts.

### Decision 5: DLQ after 3 receive attempts

**Choice:** Configure maxReceiveCount = 3, then move to DLQ.

**Rationale:** Transient failures (Lambda throttle, temporary network) should retry. After 3 failures, assume permanent failure and move to DLQ for manual inspection.

### Decision 6: Worker processes one message per Lambda invocation

**Choice:** SQS trigger invokes Lambda with one message at a time (batch size = 1).

**Rationale:** Simpler error handling, easier to trace, matches Lambda concurrency model. If we need higher throughput, configure reserved concurrency and increase batch size.

### Decision 7: Scheduler runs every 1 minute

**Choice:** EventBridge Schedule triggers scheduler every 1 minute.

**Rationale:** EventBridge rate expressions support minimum 1 minute. 1-minute interval is sufficient for near-real-time response while not overwhelming the system. Users with 60-second intervals get 1 check per cycle. Can be adjusted later.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    BACKGROUND EXECUTION ARCHITECTURE                         │
└─────────────────────────────────────────────────────────────────────────────┘

  ┌─────────────────────────────────────────────────────────────────────────┐
  │  SCHEDULER PATH (EventBridge → Scheduler → SQS + DynamoDB)              │
  │                                                                         │
  │  EventBridge Schedule (rate 1 minute)                                  │
  │         │                                                                │
  │         ▼                                                                │
  │  ┌─────────────────────────────────────────────────────────────────┐   │
  │  │  check-runtime Lambda                                            │   │
  │  │  RUNTIME_MODE=scheduler                                         │   │
  │  │                                                                 │   │
  │  │  1. Read SchedulerConfig from DynamoDB                         │   │
  │  │     • RecurringEnabled must be true                           │   │
  │  │                                                                 │   │
  │  │  2. Query all enabled monitors from DynamoDB                  │   │
  │  │     • Filter: monitors where Enabled = true                   │   │
  │  │                                                                 │   │
  │  │  3. For each monitor × probeLocation:                          │   │
  │  │     a. Build ExecutionRequest (monitor config + location)     │   │
  │  │     b. Send to SQS execution-queue                           │   │
  │  │     c. Write RUN_REQUEST# item to DynamoDB (audit trail)      │   │
  │  │                                                                 │   │
  │  └─────────────────────────────────────────────────────────────────┘   │
  │         │                                                                │
  │         ├────▶ SQS execution-queue (triggers worker)                   │
  │         │                                                                │
  │         └────▶ DynamoDB RUN_REQUEST# (audit trail)                      │
  │                                                                         │
  └─────────────────────────────────────────────────────────────────────────┘

  ┌─────────────────────────────────────────────────────────────────────────┐
  │  WORKER PATH (SQS → Worker → Target URL → DynamoDB)                    │
  │                                                                         │
  │  SQS execution-queue                                                    │
  │         │                                                                │
  │         │ SQS trigger (1 message per invocation)                       │
  │         ▼                                                                │
  │  ┌─────────────────────────────────────────────────────────────────┐   │
  │  │  check-runtime Lambda                                            │   │
  │  │  RUNTIME_MODE=worker                                            │   │
  │  │                                                                 │   │
  │  │  1. Receive SQS message (ExecutionRequest JSON)                │   │
  │  │                                                                 │   │
  │  │  2. Execute HTTP check                                          │   │
  │  │     • GET/POST/etc to target URL                               │   │
  │  │     • Respect timeout from monitor config                      │   │
  │  │     • Verify expected status codes                              │   │
  │  │     • Verify expected body if configured                        │   │
  │  │                                                                 │   │
  │  │  3. Record result to DynamoDB                                   │   │
  │  │     • Write RUN# item (CheckRun record)                        │   │
  │  │     • Update STATUS item (MonitorStatus)                        │   │
  │  │     • Handle incident creation/resolution                       │   │
  │  │     • Update service rollup                                    │   │
  │  │                                                                 │   │
  │  │  4. Delete SQS message (on success)                            │   │
  │  │     • On error: let visibility timeout expire (SQS retries)  │   │
  │  │                                                                 │   │
  │  └─────────────────────────────────────────────────────────────────┘   │
  │         │                                                                │
  │         ▼                                                                │
  │  ┌─────────────────────────────────────────────────────────────────┐   │
  │  │  Target URL (user's service)                                   │   │
  │  │  • HTTP response status + body                                  │   │
  │  └─────────────────────────────────────────────────────────────────┘   │
  │         │                                                                │
  │         ▼                                                                │
  │  ┌─────────────────────────────────────────────────────────────────┐   │
  │  │  DynamoDB                                                       │   │
  │  │  • RUN#<timestamp>#<runId> — CheckRun record                  │   │
  │  │  • STATUS — MonitorStatus (updated)                            │   │
  │  │  • INCIDENT#... — Incident (if failure)                       │   │
  │  │  • SERVICE#...#STATUS — Service rollup (updated)              │   │
  │  └─────────────────────────────────────────────────────────────────┘   │
  │                                                                         │
  └─────────────────────────────────────────────────────────────────────────┘

  ┌─────────────────────────────────────────────────────────────────────────┐
  │  ERROR HANDLING (DLQ FLOW)                                              │
  │                                                                         │
  │  Worker fails to process message (Lambda error, timeout, etc.)         │
  │         │                                                                │
  │         ▼                                                                │
  │  SQS does NOT delete message                                            │
  │         │                                                                │
  │         │ After visibility timeout (60s), message becomes visible       │
  │         │ again. Total attempts: 1 → 2                                 │
  │         │                                                                │
  │         ├───▶ After 3 attempts ──▶ SQS moves to DLQ                  │
  │         │                                                                │
  │         └───▶ Success ──▶ SQS deletes message                         │
  │                                                                         │
  └─────────────────────────────────────────────────────────────────────────┘
```

## SQS Message Format

```json
{
  "monitor": {
    "tenantId": "DEFAULT",
    "serviceId": "example-com",
    "monitorId": "homepage",
    "name": "Homepage check",
    "type": "http",
    "intervalSeconds": 60,
    "probeLocations": ["iad"],
    "enabled": true,
    "http": {
      "target": "https://example.com",
      "method": "GET",
      "headers": {},
      "timeoutMs": 5000,
      "expectedStatusCodes": [200],
      "expectedBodyContains": "healthy"
    }
  },
  "probeLocation": {
    "locationId": "iad",
    "displayName": "US East",
    "executionTarget": "worker-us-east",
    "enabled": true
  },
  "runId": "RUN_20260608_143000_ABC123",
  "trigger": "recurring"
}
```

## Infrastructure Changes (bootstrap.ts)

```typescript
// 1. SQS Queues
const executionQueue = new sst.aws.SQS('ExecutionQueue', {
  fifo: false,
  retentionPeriodHours: 1,
})

const executionQueueDLQ = new sst.aws.SQS('ExecutionQueueDLQ', {
  fifo: false,
  retentionPeriodHours: 24,
})

// Configure DLQ
executionQueue.addConsumer({
  queue: executionQueueDLQ,
  maxReceiveCount: 3,
})

// 2. EventBridge Schedule
const schedulerSchedule = new sst.aws.EventBridgeSchedule('SchedulerSchedule', {
  schedule: 'rate(1 minute)',
})

// 3. Scheduler Lambda
const schedulerHandler = {
  runtime: 'go',
  handler: '../services/check-runtime',
  link: [table],
  environment: {
    TABLE_NAME: table.name,
    RUNTIME_MODE: 'scheduler',
    EXECUTION_QUEUE_URL: executionQueue.url,
  },
}
schedulerSchedule.target(schedulerHandler)

// 4. Worker Lambda with SQS trigger
const workerHandler = {
  runtime: 'go',
  handler: '../services/check-runtime',
  link: [table],
  environment: {
    TABLE_NAME: table.name,
    RUNTIME_MODE: 'worker',
  },
}
executionQueue.addConsumer(workerHandler)
```

## Code Changes

### services/check-runtime/main.go

```go
func main() {
  ctx := context.Background()
  awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
  // ...

  sqsClient := sqs.NewFromConfig(awsCfg)
  handler := newRuntimeHandler(
    newDynamoRuntimeRepository(dynamodb.NewFromConfig(awsCfg), os.Getenv("TABLE_NAME")),
    sqsClient,
    os.Getenv("EXECUTION_QUEUE_URL"),
    defaultProbeLocationCatalog(),
    defaultTenantID,
    os.Getenv("RUNTIME_MODE"),
  )

  switch os.Getenv("RUNTIME_MODE") {
  case modeScheduler:
    lambda.Start(handler.handleCloudWatchEvent)
  case modeWorker:
    lambda.Start(handler.handleSQSEvent)
  default:
    log.Fatalf("unsupported runtime mode %q", os.Getenv("RUNTIME_MODE"))
  }
}
```

### services/check-runtime/sqs.go (NEW)

```go
type sqsClient interface {
  SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
  ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
  DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

func (c *awsSQSClient) SendMessage(ctx context.Context, queueURL string, body string) error {
  _, err := c.client.SendMessage(ctx, &sqs.SendMessageInput{
    QueueUrl:    aws.String(queueURL),
    MessageBody: aws.String(body),
  })
  return err
}
```

### services/check-runtime/runtime.go

Add to `runtimeHandler`:
```go
type runtimeHandler struct {
  repo       runtimeRepository
  sqsClient  sqsClient
  queueURL   string
  catalog    probelocationcatalog.Catalog
  tenantID   string
  mode       string
  now        func() time.Time
  newHTTP    func(time.Duration) *http.Client
}
```

Modify `runScheduler`:
```go
func (h runtimeHandler) runScheduler(ctx context.Context) (runtimeSummary, error) {
  // ... existing config and monitor reading ...

  // Send to SQS for each request
  for _, req := range requests {
    jsonReq, _ := json.Marshal(req)
    if err := h.sqsClient.SendMessage(ctx, h.queueURL, string(jsonReq)); err != nil {
      return summary, err
    }
  }

  // Still write to DynamoDB for audit trail (existing behavior)
  if err := h.repo.EnqueueExecutionRequests(ctx, requests, h.now()); err != nil {
    return summary, err
  }

  return summary, nil
}
```

Add `handleSQSEvent`:
```go
func (h runtimeHandler) handleSQSEvent(ctx context.Context, event events.SQSEvent) (runtimeSummary, error) {
  // SQS invokes one message at a time (batch size = 1)
  for _, msg := range event.Records {
    var req checkexecution.ExecutionRequest
    if err := json.Unmarshal([]byte(msg.Body), &req); err != nil {
      return runtimeSummary{}, err
    }

    result := checkexecution.ExecuteHTTP(ctx, h.newHTTP(time.Duration(req.Monitor.HTTP.TimeoutMs)*time.Millisecond), req)
    if err := h.repo.RecordExecutionResult(ctx, req.Monitor, req.RunID, req.ProbeLocation.LocationID, req.Trigger, result); err != nil {
      return runtimeSummary{}, err
    }
  }
  return runtimeSummary{Mode: modeWorker, Processed: len(event.Records)}, nil
}
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| SQS message lost after worker crash | SQS visibility timeout ensures message reappears; DLQ catches permanently failed messages |
| Scheduler enqueues work faster than worker processes | SQS buffers requests; Lambda concurrency can scale horizontally |
| DynamoDB write fails after SQS send | Scheduler writes to DynamoDB first, then SQS. If SQS fails, DynamoDB write rolls back. If DynamoDB fails, no SQS send. |
| Lambda cold start delays | Provisioned concurrency optional; acceptable for monitoring use case |
| Too many concurrent executions causing throttling | Configure reserved concurrency on worker Lambda if needed |

## Migration Plan

1. Deploy new `check-runtime` binary with SQS support (no SQS trigger yet)
2. Deploy infrastructure: SQS queues, EventBridge schedule, Lambda definitions (no triggers yet)
3. Test scheduler in isolation: Verify monitors are read and messages sent to SQS
4. Enable worker trigger (SQS → Lambda)
5. Verify end-to-end: monitor executes, result recorded, status updated
6. Monitor DLQ for any failures and adjust as needed

**Rollback:** Disable EventBridge schedule, remove SQS trigger from worker Lambda.

## Open Questions

1. **Should we use SQS batch size > 1 for worker?** Currently designed for batch size 1. Could increase for higher throughput but complicates error handling.

2. **Should scheduler delete old RUN_REQUEST# items?** Currently items persist forever. DynamoDB TTL could be added later.

3. **Should we add CloudWatch metrics for queue depth?** Useful for monitoring but not critical for initial implementation.