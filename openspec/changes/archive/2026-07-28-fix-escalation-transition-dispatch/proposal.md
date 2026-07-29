## Why

Canonical incident transitions reach SQS but cannot be consumed: dispatch acknowledgement occurs before consumption, and the consumer looks up the causal run ID rather than the canonical transition ID. This silently suppresses alerts and wastes Stream/SQS retries.

## What Changes

- Keep a canonical transition outbox record dispatch-pending until downstream transition handling succeeds.
- Carry and consume the canonical transition identity consistently from DynamoDB Stream through SQS and outbox lookup.
- Add end-to-end regression coverage for Stream insert, queue message, transition handling, and acknowledgement ordering.
- Keep existing DynamoDB Stream, SQS, DLQ, and sparse outbox recovery topology; add no resources.
- FinOps: eliminate retries and DLQ churn caused by acknowledged-but-unprocessed records. No new DynamoDB indexes, queues, Lambdas, metrics, or retained payloads.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `notification-delivery-assurance`: Canonical transition dispatch and acknowledgement must preserve a processable pending record until downstream handling succeeds.

## Impact

- `services/escalation-runtime/stream_dispatch.go`
- `services/escalation-runtime/handler.go`
- Stream/queue integration tests and existing notification DLQ behavior
- No public API, deployment resource, or dependency change
