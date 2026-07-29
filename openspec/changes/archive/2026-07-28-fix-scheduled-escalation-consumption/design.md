## Context

EventBridge Scheduler targets notification SQS using canonical `scheduled_step` envelopes. Runtime SQS parsing recognizes transition and legacy event payloads only.

## Goals / Non-Goals

**Goals:** Execute validated delayed-step messages through existing SQS consumer and preserve partial failures.

**Non-Goals:** New Scheduler resources, direct Lambda Scheduler targets, schedule format migration, or new retry policy.

## Decisions

- Decode canonical envelope before legacy notification payload. Dispatch `scheduled_step` using `incidentId`, tenant identity, and `stepNumber` into existing scheduled invocation handler.
- Reject missing/invalid step identity as per-record SQS failure. This preserves existing DLQ behavior.
- Reuse current SQS and Scheduler setup. Fix removes retries/DLQ writes; no new service, metric, or storage cost.

## Risks / Trade-offs

- Duplicate Scheduler delivery -> Existing escalation state checks make stale/terminal work no-op.
- Malformed envelope -> DLQ traffic remains intentional and bounded.

## Migration Plan

Deploy handler and tests. Create a short-delay staging policy step, verify one SQS message processes, then scheduler deletes completed schedule. Rollback leaves queued canonical work retryable after forward deploy.
