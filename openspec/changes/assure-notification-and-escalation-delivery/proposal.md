## Why

Incident transitions and escalation notifications can currently be lost, duplicated, or replayed long after they are relevant because enqueue errors are discarded, delivery attempts are not persisted per channel, and one-time escalation rules are not cleaned up. Operators also lack a safe way to inspect or replay failed deliveries, making provider and queue failures difficult to diagnose and recover.

## What Changes

- Depend on the corrected `make-check-execution-retry-safe` contract for the single canonical transition event/outbox record created atomically with the recurring result and incident transition; do not create or publish a competing transition outbox.
- Own the one dispatch path from canonical DynamoDB Stream records to the existing notification queue, including conditional dispatch acknowledgement, a durable pending access path, bounded identity-based reconciliation, and manual repair after Stream retry exhaustion.
- Give each notification delivery a durable identity derived from the incident transition, policy step, and channel, with exactly `pending`, `in_flight`, `retryable_failed`, `ambiguous`, `delivered`, or `terminal_failed` state and sanitized provider metadata. Recovery suppression remains escalation eligibility/state, not a delivery state.
- Make retries idempotent at the delivery-record boundary so completed channels are not resent when another channel in the same step fails.
- Classify timeouts, provider throttling (`429`), and provider `5xx` responses as retryable while treating invalid channel configuration and terminal provider `4xx` responses as failed without automatic retry.
- Preserve poison messages with partial-batch responses and source-kind metadata, and restrict notification-DLQ redrive so Stream failures reconcile by outbox identity, Scheduler failures revalidate scheduled-step identity, and only valid SQS delivery messages return to the notification queue.
- Replace persistent annual cron rules and per-rule Lambda permissions with deterministic, self-cleaning one-time escalation schedules that have bounded retries, a DLQ, least-privilege invocation permissions, and action-after-completion cleanup.
- Re-check incident and escalation state before every scheduled step so recovery suppresses all future escalation work.
- Add incident-scoped API and dashboard visibility for delivery outcomes plus a controlled replay action for eligible terminal failures. Replay requires an `Idempotency-Key`: the same key and request returns one replay result, a changed request conflicts, and the replay-key record expires after bounded retention.
- Define and assert the ordering and retry budgets among provider timeout, notification Lambda timeout, delivery attempt lease, SQS visibility/max receives and backoff, and EventBridge Scheduler retry age/attempts and DLQ behavior.
- Document DLQ inspection/replay and failed-delivery recovery, and add unit, integration, infrastructure, API, and dashboard coverage for loss, retry, partial-success, suppression, poison-message, and schedule-cleanup behavior.
- Keep the existing five integration types and existing DynamoDB table and SQS queues; this change does not introduce an on-call platform or guarantee human receipt beyond provider acceptance.

## Capabilities

### New Capabilities
- `notification-delivery-assurance`: Canonical outbox dispatch and reconciliation, exact per-channel delivery states, retry classification, idempotent partial-success handling, source-aware poison handling, and recovery suppression.
- `escalation-one-time-scheduling`: Deterministic self-cleaning one-time escalation schedules with least privilege, bounded retries, DLQ handling, and no annual replay or leaked rule permissions.
- `notification-delivery-operations`: Incident-scoped API and dashboard delivery visibility, idempotency-key-controlled replay, sanitized metadata, and bounded operational repair requirements.

### Modified Capabilities
- `dynamodb-single-table-storage`: Consume the retry-safe canonical transition outbox and add its bounded pending-dispatch access path plus tenant-safe notification delivery, replay-command, and replay-idempotency records in the existing table.
- `incident-activity-read-api`: Correlate incident activity transitions with notification delivery outcomes while preserving chronological incident history.
- `api-documentation`: Document the incident delivery read and replay operations in the source-controlled API contract.

## Impact

- Affects `services/check-runtime`, `services/escalation-runtime`, `services/monitor-api`, shared notification/error/storage models, the incident detail dashboard, Bruno coverage, OpenAPI, and `infra/stacks/bootstrap.ts`.
- Reuses the primary DynamoDB table, execution queue, notification queue, and notification DLQ. EventBridge Scheduler schedules add short-lived per-step schedule operations; delivery, dispatch-state, and bounded replay-idempotency records add bounded DynamoDB reads and writes. No new always-on service, queue, table, provider integration, or on-call product is introduced.
- Requires AWS Scheduler execution-role permissions scoped to the escalation target and notification DLQ, and removal of runtime EventBridge rule/Lambda permission mutation.
- Provider acceptance is the terminal success boundary. The system does not claim that a person read or received a notification, and ambiguous network failures may still depend on provider-supported idempotency semantics.
- Implementation is blocked until `make-check-execution-retry-safe` exposes the corrected canonical outbox contract: one deterministic `eventId`, immutable causal payload, pending/acknowledged dispatch metadata, and atomic creation with the accepted result and incident transition. This change replaces any direct post-commit notification publisher with its Stream dispatcher rather than adding another producer protocol.
