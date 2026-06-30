## Why

The manual check execution change verifies that monitors work end-to-end via API. Now we need automated background execution so monitors run on their configured schedules without manual triggers. This requires scheduling infrastructure (EventBridge), work distribution (SQS), and a worker that processes execution requests.

## What Changes

- **New**: EventBridge Schedule triggers scheduler Lambda every 30 seconds
- **New**: Scheduler Lambda mode reads enabled monitors from DynamoDB, sends work to SQS queue
- **New**: SQS queue (execution-queue) holds execution requests
- **New**: SQS Dead Letter Queue (execution-queue-dlq) for failed messages
- **New**: Worker Lambda mode receives SQS events and executes HTTP checks
- **New**: DynamoDB `RUN_REQUEST#` items serve as audit trail (not consumed by worker)
- **Modified**: Existing `check-runtime` binary operates in two modes based on `RUNTIME_MODE` env

## Capabilities

### New Capabilities
- `scheduler-eventbridge-trigger`: EventBridge Schedule triggers scheduler Lambda on configured interval to enqueue work
- `execution-sqs-queue`: SQS queue distributes execution work to worker Lambdas with visibility timeout and DLQ support
- `check-runtime-scheduler-mode`: Scheduler mode reads monitors and enqueues execution requests to SQS
- `check-runtime-worker-mode`: Worker mode receives SQS events and executes HTTP checks against target URLs
- `execution-work-item`: DynamoDB stores RUN_REQUEST# items as audit trail for manual and scheduled executions

### Modified Capabilities
- `check-execution-pipeline`: Extend to support event-driven SQS-based execution in addition to manual execution
- `manual-run-api`: After this change, manual runs still execute inline; scheduled runs go through SQS

## Impact

- **Infra**: `infra/stacks/bootstrap.ts` — add SQS queues, EventBridge schedule, two Lambda functions
- **Code**: `services/check-runtime/main.go` — add SQS event handler alongside existing CloudWatch handler
- **Code**: `services/check-runtime/runtime.go` — add SQS send in scheduler mode, SQS receive in worker mode
- **Code**: `services/check-runtime/sqs.go` (NEW) — SQS client wrapper for sending/receiving messages
- **Code**: `services/check-runtime/repository.go` — already has `EnqueueExecutionRequests` writing to DynamoDB (keep for audit), need to add SQS send
- **Dependencies**: Add AWS SQS SDK to `check-runtime/go.mod`