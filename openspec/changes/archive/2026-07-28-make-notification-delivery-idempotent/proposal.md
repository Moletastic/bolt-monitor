## Why

Notification provider I/O occurs before durable delivery claim and completion. SQS retries can resend accepted alerts, while existing delivery lifecycle persistence is unused.

## What Changes

- Persist route plan and deterministic per-channel deliveries before provider I/O.
- Conditionally claim delivery identity before each provider attempt and complete it with sanitized outcome data.
- Skip terminal/delivered deliveries on duplicate SQS or Scheduler work.
- Define handling for provider acceptance followed by persistence failure without promising impossible exactly-once external delivery.
- Add retry, concurrency, and partial multi-channel regression coverage.
- FinOps: reuse existing DynamoDB delivery items and notification queue. Fewer duplicate provider calls, Lambda retries, and logs; no new GSI, queue, Lambda, or high-cardinality metric.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `notification-delivery-assurance`: Delivery claims and terminal outcomes must fence provider calls across at-least-once processing.

## Impact

- `services/escalation-runtime/handler.go`
- `services/escalation-runtime/delivery_orchestration.go`
- `services/escalation-runtime/delivery_repository.go`
- Provider sender tests and notification delivery state behavior
- No public API, deployment resource, or dependency change
