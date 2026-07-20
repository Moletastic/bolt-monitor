## Context

Notification work currently crosses three durability boundaries without a complete delivery state machine. The prerequisite `make-check-execution-retry-safe` change corrects the first boundary by atomically creating one deterministic incident transition event/outbox record with the accepted recurring result and incident transition. This change consumes that record; it does not create a second transition outbox or retain the retry-safe design's direct post-commit publisher. `escalation-runtime` currently sends channels directly, advances escalation state only after all sends return, and has no record of a channel that already succeeded. Its SQS handler acknowledges malformed and unsupported messages, while provider senders return opaque errors that include raw response bodies and do not distinguish retryable failures. Delayed steps are implemented as annual EventBridge cron rules with per-rule Lambda permissions; neither the rules, targets, nor permissions are removed.

The existing primary DynamoDB table, execution queue, notification queue, and notification DLQ are sufficient. The table already stores incident activity and escalation state, and the notification queue already provides at-least-once delivery and max-receive redrive. The design must preserve the five current channel types and the single-tenant-compatible ownership model, while keeping records tenant-safe for future use.

Provider acceptance is the observable delivery boundary. A timeout after provider acceptance is inherently ambiguous unless that provider supports idempotency. The system can prevent resending records known to be delivered and can use stable provider idempotency keys, but cannot prove exactly-once external side effects for every provider.

## Goals / Non-Goals

**Goals:**

- Make every notification-relevant incident transition recoverable after enqueue failure.
- Persist and expose deterministic per-channel delivery state and safe diagnostics.
- Retry transient failures without resending channels already known to be delivered.
- Preserve poison work in the existing notification DLQ.
- Stop future escalation promptly after incident recovery.
- Replace annual rules and per-incident Lambda permissions with deterministic self-deleting one-time schedules.
- Provide incident-scoped inspection, controlled replay, runbook guidance, and regression tests.
- Reuse current queues and table with bounded incremental cost.
- Consume one canonical retry-safe transition/outbox contract and own its Stream dispatch, acknowledgement, reconciliation, and repair path.

**Non-Goals:**

- Adding or removing provider integration types.
- Building an on-call assignment, acknowledgement, rotation, or paging platform.
- Guaranteeing human receipt, reading, or action after provider acceptance.
- Guaranteeing exactly-once provider side effects when a provider lacks idempotency and an attempt has an ambiguous network result.
- Replaying successful deliveries or automatically retrying terminal configuration/provider rejections forever.
- Introducing a new queue, DynamoDB table, or always-on worker.
- Creating incident transitions, transition events, or a competing transition publication/outbox protocol.

## Decisions

### 1. Consume the retry-safe canonical outbox through one Stream dispatcher

This change explicitly depends on the corrected `make-check-execution-retry-safe` outbox contract. That prerequisite owns atomic creation of exactly one canonical record with deterministic `eventId`, causal `runId`, scheduled slot, incident ID, transition type/time, immutable payload fingerprint, and `dispatchStatus=pending`. Its `eventId` is also the incident activity identity and becomes `transitionId` throughout queue messages, escalation state, schedules, deliveries, API responses, and logs. This change neither calls incident result persistence nor conditionally creates another transition record.

The existing table's DynamoDB Stream invokes `escalation-runtime` for inserts of canonical dispatch records. The dispatcher sends the versioned canonical envelope to the existing notification queue and conditionally changes that same record from `pending` to `acknowledged` with SQS message metadata. An ambiguous SQS result leaves it pending. Stream filters select inserts only, so acknowledgement updates cannot loop. The retry-safe change's direct worker-to-notification-queue publication and recovery are replaced by this owner; there is only one transition dispatch path.

Stream retries are bounded and use per-record partial-batch failure. Exhaustion writes the Stream invocation envelope to the notification DLQ, but the canonical record remains `pending` and appears in a sparse, tenant-bucketed pending-dispatch access path keyed by `eventId` and creation-time bucket. A scheduled bounded reconciler reads at most configured pages from recent/explicit buckets and dispatches by canonical identity; an operator can reconcile one `eventId` directly. Neither path scans the table. Conditional acknowledgement makes Stream, reconciler, and manual repair races harmless. Pending records have no TTL; acknowledged records receive bounded retention.

Replay uses the same dispatch record schema and Stream-to-SQS dispatcher with `sourceKind=delivery_replay`; it does not introduce another queue publisher or outbox protocol. The replay API transaction creates one replay command and replay-idempotency record while changing the eligible delivery to `pending`.

Alternatives considered:

- Returning the enqueue error from `check-runtime` alone is insufficient because the incident transition has already committed and a retried check may not emit the same transition.
- Unbounded polling of historical outbox records is prohibited; the sparse time-bucketed pending access path and point reconciliation provide bounded recovery after Stream exhaustion.
- Publishing before the incident transaction can produce notification work for an incident transition that never commits.

### 2. Persist an immutable escalation plan and deterministic delivery records

The first successful handling of an `incident.down` transition will create escalation state before provider I/O and snapshot the selected path's ordered step numbers, delays, and channel IDs. It will not snapshot channel credentials or provider configuration. This prevents policy edits or duplicate transition messages from changing the route for an in-flight incident, while a later channel configuration correction can be used by an explicit replay.

For each channel in a step, the runtime computes:

```text
deliveryId = sha256(tenantId + "\n" + transitionId + "\n" + stepNumber + "\n" + channelId)
```

The externally returned ID uses a fixed, lowercase encoded prefix of the digest; canonical inputs remain fields on the record. Delivery records live in the incident partition and sort by creation timestamp, step, channel, and delivery ID. Conditional writes enforce one record per identity. This handles the current channel reference and any existing multi-channel step data without increasing provider integration count.

An escalation step advances exactly once after all its channel deliveries are terminal (`delivered` or `terminal_failed`). A transient or ambiguous failure keeps only that delivery active while automatic retry budget remains; a terminal failure does not block subsequent policy steps. Conditional state advancement prevents duplicate workers from scheduling the next step twice.

Alternative considered: keying only by incident and channel would incorrectly collapse repeated use of the same channel in different policy steps.

### 3. Use one exact delivery state machine with conditional claims

Every delivery has exactly one state:

- `pending`: durable work exists and no provider attempt is active;
- `in_flight`: one fenced claim owns a provider attempt until `leaseUntil`;
- `retryable_failed`: the last confirmed outcome was retryable and may be claimed after `nextAttemptAt` while budget remains;
- `ambiguous`: the request may have reached the provider but acceptance was not confirmed; it may be claimed after `nextAttemptAt` while budget remains, with duplicate-side-effect risk disclosed;
- `delivered`: terminal provider-accepted success;
- `terminal_failed`: terminal rejection/configuration failure or exhausted automatic retry budget.

Recovery suppression is an escalation eligibility/state (`active` or `suppressed`), never a delivery state. Existing delivery history remains unchanged when escalation is suppressed.

Before provider I/O, a conditional update claims `pending`, eligible `retryable_failed`, eligible `ambiguous`, or lease-expired `in_flight`; sets `in_flight`; assigns a fencing token; increments `attemptCount`; and records `lastAttemptAt` and `leaseUntil`. Reclaiming an expired `in_flight` first classifies the abandoned outcome as `ambiguous` in the same conditional transition. A current lease, `delivered`, or `terminal_failed` record cannot be claimed.

On provider acceptance, the fenced worker writes `delivered`. A confirmed transient response writes `retryable_failed`; a timeout, connection loss after request transmission, or crash boundary writes or recovers as `ambiguous`; and a non-retryable response writes `terminal_failed`. Retryable and ambiguous results report the SQS record failure. At the automatic attempt or SQS receive limit, known unfinished deliveries become `terminal_failed` with `retry_exhausted` while preserving the last outcome class, and the record remains failed so SQS moves the source message to the notification DLQ.

If a worker crashes after provider acceptance but before recording `delivered`, a later lease retry can duplicate the external request. Senders will pass `deliveryId` as the provider idempotency/deduplication key where supported. This is the unavoidable trade-off for providers without such a contract and is called out in operator text and the runbook.

### 4. Return typed, sanitized sender outcomes

The notification sender boundary will return a structured result/error containing:

- normalized class: `accepted`, `timeout`, `transport`, `throttled`, `provider_5xx`, `provider_4xx`, `invalid_config`, or `unsupported_channel`;
- retryable boolean;
- safe HTTP status code or class;
- allowlisted provider request ID when documented safe;
- parsed and bounded retry-after time when valid.

Senders will validate configuration before request construction. HTTP helpers will consume and close response bodies but will not include raw bodies, headers, credentials, targets, or full URLs in returned or persisted errors. Logs use delivery, incident, transition, step, channel ID/type, attempt, and normalized class only. The test-send API will adapt the same typed result to its existing sanitized error contract without creating durable incident delivery records.

Status policy is centralized: timeout/transport/`429`/`5xx` retry; invalid configuration/unsupported type/other `4xx` fail terminally. Provider acceptance remains each sender's documented successful response range.

Alternative considered: parsing strings from existing errors is brittle and risks persisting raw response bodies.

### 5. Use source-specific partial batches and DLQ repair

`escalation-runtime` distinguishes DynamoDB Stream events from notification SQS events at the Lambda boundary and returns the source's `batchItemFailures` identifiers. Stream records use DynamoDB sequence numbers; SQS records use message IDs. For either source, malformed records, unknown versions/kinds, retryable repository failures, and retryable or ambiguous provider outcomes are reported as failed. Successfully completed and terminally classified records are omitted. A mixed batch therefore never retries a successful sibling because another record is poison.

Infra enables `ReportBatchItemFailures` for both event sources and retains the current notification queue redrive count and DLQ. Although queue batch size is currently one, partial responses prevent future batch-size changes from coupling poison and successful messages.

Every DLQ envelope carries `sourceKind` (`dynamodb_stream`, `notification_sqs`, or `scheduler_target`) and safe canonical identity when parseable. Redrive is source-aware: a Stream failure envelope is never sent to the notification queue and instead triggers point reconciliation of its canonical dispatch record; a Scheduler target envelope may enqueue only the validated scheduled-step message after current incident/escalation eligibility is rechecked; an SQS envelope may return to the queue only if its canonical version/kind/identity validates. Malformed or unknown envelopes are quarantined. Bulk blind DLQ redrive is prohibited.

### 6. Assert timeout, lease, visibility, and retry ordering

Named configuration constants and infrastructure tests enforce:

```text
ProviderRequestTimeout + ProviderCompletionBuffer < NotificationLambdaTimeout
NotificationLambdaTimeout + LambdaTerminationBuffer < DeliveryAttemptLease
ClaimStartBudget + DeliveryAttemptLease + RedeliveryBuffer <= NotificationQueueVisibilityTimeout
DeliveryAutomaticAttemptLimit <= NotificationQueueMaxReceiveCount
NotificationRetryBackoffMax < DeliveryAttemptLease
SchedulerTargetRetryAge >= SchedulerTargetRetryBackoffMax
```

`ProviderCompletionBuffer` covers conditional outcome persistence, and `ClaimStartBudget` bounds event receipt, parsing, reads, and the claim before provider I/O. The lease outlives a valid Lambda invocation but expires before SQS redelivery becomes visible, preventing both concurrent sends and active-lease receive-count burn. Retry-after/backoff is bounded by `NotificationRetryBackoffMax`; it cannot extend a claim lease or exceed the remaining automatic retry budget. On the last allowed provider attempt or queue receive, known unfinished deliveries terminalize before the SQS record is failed for DLQ redrive. Scheduler retry attempts and maximum event age apply only to delivery of a scheduled-step message to SQS; after exhaustion its source-tagged envelope goes to the DLQ and does not consume a provider attempt. Scheduler constants are finite and asserted with its DLQ configuration.

### 7. Target the notification queue with EventBridge Scheduler one-time schedules

Delayed steps use AWS EventBridge Scheduler `CreateSchedule` with:

- a deterministic name based on a bounded hash of tenant, transition, and step;
- a dedicated managed schedule group;
- `at(<UTC timestamp>)` and flexible time window `OFF`;
- `ActionAfterCompletion: DELETE`;
- the notification queue ARN as target;
- the existing notification DLQ ARN as Scheduler target DLQ;
- bounded target retry age and attempts;
- a dedicated execution role that can send only to those queues.

The message contains version, kind, tenant, incident, transition, and step. Runtime permissions permit schedule create/get/update only in the managed group and `iam:PassRole` only for the dedicated execution role. The runtime no longer needs EventBridge rule, STS account lookup, Lambda permission, or direct Lambda scheduling clients. A deterministic name plus input comparison makes create retries idempotent; an existing schedule with conflicting immutable identity is treated as an error rather than overwritten.

Every scheduled message re-reads the incident and escalation state before creating or claiming deliveries. Resolution suppresses the escalation even if the `incident.up` outbox event is delayed. The recovery event still persists `SUPPRESSED` for visibility. Pending schedules need not be explicitly deleted because they self-delete after invocation and become harmless when the state check suppresses them.

Alternative considered: disabling/deleting EventBridge rules manually after invocation leaves crash windows, per-rule Lambda policy growth, and annual replay risk. Direct Scheduler-to-Lambda would retain a second processing path and broader invocation permissions; Scheduler-to-SQS uses the existing retry/DLQ/idempotency boundary.

### 8. Add incident-scoped read and idempotent replay operations

The monitor API adds:

```text
GET  /api/v1/incidents/{incidentId}/deliveries
POST /api/v1/incidents/{incidentId}/deliveries/{deliveryId}/replay
```

The read verifies incident ownership, queries only the incident partition, and returns safe fields in the standard envelope. The replay operation requires an `Idempotency-Key`, verifies incident and delivery ownership, current incident/escalation eligibility, and `terminal_failed` state, then conditionally transacts the same delivery to `pending`, increments `replayCount`, records audit metadata, and creates a `delivery_replay` canonical dispatch record consumed by the same Stream dispatcher. A successful sibling remains untouched.

The idempotency record is scoped to tenant, incident, delivery, operation, and key and stores a canonical request fingerprint plus the replay result identity. The same key with the same request returns the original result and creates no additional replay; the same key with a different fingerprint returns a typed idempotency conflict. Concurrent first requests converge through the transaction. Records use a named bounded retention duration longer than the maximum replay dispatch/retry window; after expiry the key may be reused as a new request only if the delivery is again eligible.

The incident escalation tab will group deliveries by transition and step, show each channel independently, and use “accepted by provider” language. Replay uses a server action and inline pending/result state. No imperative router APIs are added.

Alternative considered: direct API-to-provider replay bypasses queue durability and would create a second delivery path.

### 9. Preserve observability without persisting sensitive data

Structured logs and metrics will cover outbox dispatch failures, delivery attempts by normalized class/state/channel type, queue retries, DLQ movement indicators, schedule create conflicts/failures, suppression, and replay. Dimensions exclude tenant IDs where they create high cardinality or disclosure risk and exclude incident/delivery IDs from metrics; IDs remain in structured logs for correlation.

The incident activity endpoint remains backward compatible. Its `activityId` correlates with `transitionId` on the new deliveries endpoint rather than embedding a potentially large delivery collection into every activity response.

### 10. Keep incremental cost bounded

The retry-safe prerequisite already adds one canonical outbox write per notification-relevant incident transition. This design adds one Stream invocation per canonical dispatch insert, one delivery record plus a few conditional updates per channel attempt, bounded replay/idempotency records, bounded pending-index reconciliation reads, and one short-lived Scheduler schedule per delayed step. Acknowledged dispatch and replay-idempotency records receive bounded TTL; pending dispatch records remain durable until acknowledged or manually quarantined. Stream filters avoid invocations for unrelated table writes. API reads are incident-partition queries and repair uses sparse bounded queries or point reads, never scans. Existing queues, DLQ, table, Lambda, and provider integrations are reused.

This costs more than silent best-effort sending but avoids a new table, queue, polling service, or permanent rule per escalation.

## Risks / Trade-offs

- [Provider accepted a request before a timeout or worker crash] -> Reuse stable provider idempotency keys where supported, expose the ambiguity class, and require operator judgment before replay; do not claim universal exactly-once delivery.
- [DynamoDB Stream is unavailable or retries exhaust] -> Keep the canonical record pending in its sparse bucketed access path, route the source-tagged Stream envelope to the existing notification DLQ, alarm, run bounded recent-bucket reconciliation, and support point repair by `eventId`; never scan or silently acknowledge loss.
- [Notification DLQ contains queue, Scheduler, and stream failure envelopes] -> Include/version source-kind metadata and document source-specific inspection and replay procedures.
- [Policy or channel changes during an active incident] -> Snapshot step order/delay/channel IDs at escalation start, but resolve channel configuration at an attempt so corrected configuration can be replayed; never alter a delivered record.
- [Attempt lease expires while a slow provider request is still running] -> Set the lease beyond HTTP timeout plus Lambda buffer and use conditional completion writes; monitor lease conflicts.
- [Terminal failure allows escalation to continue] -> Preserve the failed outcome visibly and continue later policy steps so one bad integration does not suppress the rest of the escalation route.
- [Existing annual rules fire during migration] -> Keep a temporary legacy scheduled-payload adapter that re-enqueues canonical step work, where delivery identity suppresses duplicates, until legacy rules and permissions are inventoried and removed.
- [Replay follows an ambiguity that exhausted retries] -> Restrict replay to `terminal_failed`, preserve the prior ambiguous outcome class and identity history, require `Idempotency-Key`, and explain duplicate risk in the UI/runbook.
- [Additional writes increase DynamoDB cost] -> Use compact records, bounded metadata, TTL only for dispatched outbox records, no new GSI, and incident-partition queries.

## Migration Plan

1. Land the corrected `make-check-execution-retry-safe` canonical outbox contract and remove/disable its direct notification publisher so this change has sole dispatch ownership.
2. Add delivery, replay-command/idempotency, pending-dispatch access-path mappings, typed sender outcomes, repository conditional operations, and backward-compatible readers. Existing incidents without records return empty delivery lists.
3. Provision the Scheduler group, dedicated least-privilege execution role, table stream filter/subscription, partial batch response settings, sparse pending access path, reconciler bounds, named timeout/retry constants, and existing-DLQ destinations before enabling dispatch.
4. Deploy `escalation-runtime` support for Stream dispatch, canonical SQS messages, durable deliveries, one-time schedules, suppression checks, and a temporary adapter for already-created legacy direct scheduled invocations.
5. Deploy monitor API, OpenAPI, Bruno requests, dashboard visibility/replay, observability, and the source-aware operations runbook.
6. Verify new delayed steps create Scheduler schedules that delete after completion and no new `esc-*-step-*` EventBridge rules or `allow-events-*` Lambda statements appear.
7. Inventory and remove legacy escalation rule targets, rules, and matching Lambda permission statements using the runbook. Remove the temporary direct-schedule adapter in a follow-up after the maximum legacy schedule horizon or confirmed cleanup.

Rollback keeps new table item types and IAM resources because old code ignores them. Producers can be returned to the prior path only with explicit acceptance of its loss/leak behavior; pending outbox and delivery records must not be deleted. Do not remove Scheduler roles or stream subscriptions until no new schedules/outbox records depend on them. If API/dashboard rollout fails, roll those consumers back independently while retaining delivery processing.

## Open Questions

None. Named timeout, lease, visibility, retry/backoff, Scheduler, reconciliation-page/bucket, acknowledged-dispatch TTL, and replay-idempotency retention constants SHALL be chosen within AWS limits, documented beside infrastructure configuration, and covered by ordering/bound assertions; they do not alter the behavioral contract above.
