## Why

The check execution path currently crosses DynamoDB, SQS, HTTP, incident state, and notification delivery without one stable identity or conditional state machine, so retries and partial failures can duplicate observations, lose work, regress status, or emit duplicate incident notifications. The pipeline needs explicit at-least-once invariants and recovery behavior before recurring execution can be treated as operationally reliable.

## What Changes

- Assign every accepted execution one stable `runId` before any side effect; recurring identity includes immutable `scheduleDefinitionVersion` and UTC `scheduledFor`, and is reused across scheduler retries.
- Carry execution identity and trigger metadata through persisted work, SQS messages, `CheckRun`, `MonitorStatus`, incident transitions/activity, and canonical notification outbox items.
- Define conditional create, claim, lease recovery, complete, and skip transitions for execution work, with duplicate deliveries becoming no-ops after a terminal result.
- Make scheduler monitor discovery paginated and retry-safe, preserving per-monitor progress when one persistence or enqueue operation fails.
- Persist work before queue publication and maintain directly queryable, bounded publication/work recovery markers that are conditionally removed at terminal or publication-complete states; duplicate queue publication remains safe under standard SQS at-least-once delivery.
- Revalidate current monitor existence, enabled state, maintenance state, and execution-relevant configuration before an HTTP side effect.
- Commit each canonical result atomically with terminal work state and condition status/incident changes on observation ordering. A notification-relevant transition atomically creates one deterministic activity/transition and one canonical notification outbox item; notification assurance exclusively owns dispatch, acknowledgement, delivery claims, and per-channel outcomes.
- Give manual and recurring runs one result-recording contract while retaining their trigger distinction. The service-scoped manual command requires `Idempotency-Key`, maps it deterministically to a request fingerprint and `runId` for bounded replay, and never silently advances recurring health, rollup, or incident semantics.
- Add typed runtime failure classification, explicit timeout/visibility/lease/concurrency ordering, partial-batch failure behavior, and deterministic fault-injection tests for every owned DynamoDB, execution-SQS, and HTTP boundary.
- Reuse the existing DynamoDB table and standard execution queue. This change does not directly send notification SQS messages or own notification dispatch acknowledgement, delivery claims, or per-channel outcomes; it does not claim FIFO ordering or exactly-once execution, add an execution protocol, calculate SLOs, or introduce new infrastructure resources.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `execution-sqs-queue`: Require stable execution identity, retry-safe publication, and duplicate-delivery semantics on the existing standard queue.
- `check-runtime-scheduler-mode`: Define deterministic recurring schedule-definition/time identities, paginated discovery, durable partial progress, and persistence/publication recovery.
- `check-runtime-worker-mode`: Define conditional leasing, stale-claim recovery, current-config revalidation, terminal duplicate suppression, and typed retry behavior.
- `check-execution-pipeline`: Define the end-to-end at-least-once state machine and canonical-observation invariants.
- `check-result-status-model`: Protect canonical run, status, and incident projections from duplicate and out-of-order results while preserving trigger semantics.
- `manual-run-api`: Add service-scoped request idempotency, unify manual result persistence with the pipeline, and isolate manual runs from recurring health progression.
- `incident-management-api`: Make execution-driven incident transitions, activity, and canonical notification outbox creation idempotent and causally tied to the accepted recurring run.

## Impact

Implementation will affect the shared execution and result models, DynamoDB record/key mappings and conditional repository operations, scheduler and SQS worker handlers, synchronous manual-run persistence, incident transition/activity/outbox construction, AWS facade inputs, infrastructure runtime settings, and tests. The current manual route remains `POST /api/v1/services/{serviceId}/monitors/{monitorId}/run` and gains the `Idempotency-Key` header; response envelopes remain unchanged. Stored records and execution messages gain identity, schedule-definition, lease, publication, ordering, recovery-marker, and failure metadata. Rollout is coordinated with `assure-notification-and-escalation-delivery`: its sole dispatcher must be enabled before outbox-producing paths, after which canonical pending outbox records safely bridge temporary dispatcher unavailability without a direct-send fallback or silent notification gap.
