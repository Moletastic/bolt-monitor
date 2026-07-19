## Context

Recurring execution currently derives due work from mutable `LastExecutionAt`, sends an SQS message before writing its work record, and may generate different run IDs in the message and record. Standard SQS can redeliver, Lambda can fail after the HTTP request, and result persistence can be retried. Notification assurance defines the sole transition-outbox dispatcher and downstream delivery protocol; this change must produce its canonical outbox item atomically without creating a competing direct-send path. The worker also executes the monitor snapshot from the message rather than reloading current enabled, maintenance, and configuration state.

The existing DynamoDB table, execution queue, and execution DLQ are sufficient for this change; notification assurance owns use of the existing notification queue and DLQ. DynamoDB conditional operations and transactions can make persisted state canonical, but they cannot make an external HTTP request exactly once. The design therefore guarantees at-least-once attempts and exactly one canonical accepted observation per run/eligible recurring schedule definition/time, not exactly-once network side effects.

## Goals / Non-Goals

**Goals:**

- Establish immutable execution identity before persistence, queue publication, or HTTP execution.
- Make scheduler retries, SQS redelivery, Lambda failure, and stale worker recovery converge on one terminal work record and one canonical `CheckRun`.
- Ensure recurring status counters, incidents, service rollups, and transition outbox creation move forward in `scheduledFor` order and never regress because a result finishes late.
- Recover both persisted-but-not-enqueued work and successfully-enqueued work whose publication acknowledgement was not persisted through directly queryable bounded markers.
- Keep manual and recurring runs on one commit path while making their different projection semantics explicit.
- Make manual command retries deterministic through a bounded `Idempotency-Key` record and request fingerprint.
- Configure and test strict timeout, visibility, lease, partial-batch, and event-source concurrency relationships.
- Classify expected conflicts and operational failures with typed internal errors and verify partial failures through deterministic fault injection.

**Non-Goals:**

- FIFO ordering, exactly-once HTTP execution, or an exactly-once SQS claim.
- New tables, queues, GSIs, workflow services, or a new execution protocol.
- Probe-location or region routing.
- SLO, uptime, or availability-window calculations.
- Replaying every missed wall-clock schedule time after a scheduler outage; an invocation materializes only the current eligible `scheduledFor`.
- Sending notification SQS messages, acknowledging dispatch, claiming deliveries, or recording per-channel outcomes; `assure-notification-and-escalation-delivery` solely owns those behaviors.

## Decisions

### 1. Identity is assigned before side effects

An execution envelope carries `runId`, `trigger`, `acceptedAt`, and, for recurring work, immutable `scheduleDefinitionVersion` plus UTC `scheduledFor`. `scheduledFor` is the UTC start instant obtained by flooring the scheduler's captured invocation time to the versioned effective schedule definition. Stable recurring identity is the normalized tuple `(tenantId, serviceId, monitorId, scheduleDefinitionVersion, scheduledFor)`, and `runId` is deterministically derived from that tuple before any write, queue send, or HTTP call. Including the version prevents a cadence edit from aliasing old and new schedule definitions at the same timestamp.

For a manual command, the API validates `Idempotency-Key`, canonicalizes the current service-scoped route and command payload, and computes a deterministic request fingerprint. A deterministic mapping of `(tenantId, serviceId, monitorId, Idempotency-Key)` selects one idempotency-record address; its first conditional creation stores one `runId` assigned before effects, the fingerprint, replay state/result reference, and bounded TTL. The same key and fingerprint resumes or returns that same run/result; the same key with a different fingerprint is a typed conflict while the record is retained. `MANUAL_IDEMPOTENCY_RETENTION` is named configuration, bounded and tested, and manual work has neither schedule field.

The scheduler captures one UTC invocation time and uses it for every page. Re-evaluating the same monitor under the same schedule definition for the same `scheduledFor` therefore derives the same key and `runId`. A schedule-affecting monitor change creates a new immutable `scheduleDefinitionVersion`; it can supersede unexecuted old work during worker revalidation but cannot mutate or alias an existing identity.

Alternative considered: continue using unrelated random IDs plus `LastExecutionAt`. This cannot correlate retries after an ambiguous write or queue timeout and can create two observations for one intended cadence time.

### 2. DynamoDB work is the authority; SQS is a wake-up signal

The work record is keyed so `runId` is uniquely addressable and conditionally created with `attribute_not_exists`. It stores identity, trigger, schedule fields, accepted time, status, publication state, current claim fencing token, lease expiry, attempt count, and terminal metadata. Re-creating an identical identity is success/no-op; finding the same key with conflicting immutable fields is a typed conflict.

Work creation also creates directly queryable marker items in bounded tenant-scoped, time-bucketed/sharded recovery partitions: an execution-publication marker while queue acceptance is unacknowledged and a work-recovery marker while work is nonterminal. Recovery uses DynamoDB `Query` against configured current/overlap buckets with limits and cursors, never an unbounded tenant/table scan. Publication acknowledgement conditionally removes the publication marker. Terminal commit/skip conditionally removes the work marker; lease changes conditionally move its due bucket under the current fencing token. Stale markers are harmless because recovery strongly loads canonical work and conditionally deletes a marker only when canonical state proves it obsolete.

The scheduler first creates work and markers, then sends an execution envelope containing the same identity to the existing execution queue, then conditionally marks publication acknowledged and removes its publication marker. If sending fails, the marker remains. If sending succeeds but acknowledgement persistence fails, retry may send a duplicate. Scheduler retries and a bounded recovery pass republish only query-returned pending work; duplicate messages are harmless because claim and terminal transitions are conditional.

Alternative considered: send first and persist second. A successful send followed by a failed write leaves a worker message with no canonical work record; a failed/ambiguous send also cannot be recovered safely.

### 3. Scheduler traversal preserves partial progress

Monitor and nested service queries consume every DynamoDB page using `LastEvaluatedKey`; processing is bounded so the invocation can stop before Lambda timeout. Each monitor is independently materialized and published. A later monitor failure does not roll back earlier durable work, and the scheduler returns a typed retryable error with page/monitor context. EventBridge retry repeats the captured event time where available; if the event lacks a usable time, the handler captures one invocation time before any side effect.

Due evaluation uses `(scheduleDefinitionVersion, scheduledFor)` rather than elapsed time since a queue send. Existing `LastExecutionAt` may remain for display/migration but is not the idempotency authority. At most one work record is accepted for an enabled, non-maintenance monitor for that identity. The scheduler does not backfill all slots missed during an outage.

Alternative considered: one transaction for a whole page. DynamoDB transaction limits, queue publication, and per-item failures make that both brittle and unable to preserve useful progress.

### 4. Claims use leases and fencing tokens

A worker conditionally claims `pending` work, or reclaims `in_progress` work whose lease is expired. Each successful claim writes a fresh unguessable fencing token, `leaseUntil`, `startedAt`, and incremented attempt count. Completion and skip require the matching fencing token. A superseded worker may finish its HTTP call but cannot commit after another worker reclaims the work.

Named configuration obeys strict ordering: `WORKER_LAMBDA_TIMEOUT > MAX_OUTBOUND_EXECUTION + RESULT_COMMIT_BUFFER`; `EXECUTION_QUEUE_VISIBILITY_TIMEOUT > WORKER_LAMBDA_TIMEOUT + VISIBILITY_MARGIN`; and `WORK_LEASE_DURATION > MAX_OUTBOUND_EXECUTION + RESULT_COMMIT_BUFFER`. Values are selected within platform limits, validated at deployment, documented beside configuration, and boundary-tested during implementation. Lease expiry enables stale recovery after worker death. Because a crash can happen after the target responds but before result commit, a reclaimed attempt can repeat the HTTP request; this is the unavoidable at-least-once side-effect boundary.

Alternative considered: permanent `in_progress` claims. They suppress retries but strand work after Lambda termination. SQS visibility alone is also insufficient because DynamoDB is the result authority and visibility can expire independently.

### 5. Workers reload current runnable state before HTTP

After claim and immediately before `ExecuteHTTP`, the worker strongly re-reads the monitor and current status. Missing, disabled, maintenance, invalid, or no-longer-eligible recurring monitors are conditionally marked `skipped` with a typed terminal reason and produce no `CheckRun`, status change, incident/activity, or outbox item. The worker executes the current persisted monitor configuration, not the untrusted SQS snapshot. Identity and trigger come from work, not message-provided mutable monitor fields.

For recurring work, a current interval/configuration change does not rewrite accepted schedule identity. The current monitor is executed only if the accepted `scheduleDefinitionVersion` remains current and `scheduledFor` remains eligible; otherwise the fenced worker records a typed superseded skip.

Alternative considered: execute the full monitor snapshot in SQS. This can call a deleted target, ignore disable/maintenance, or use stale credentials and expectations.

### 6. One conditional transaction accepts the canonical result

Result commit validates that result identity equals work identity and performs one DynamoDB transaction that:

- condition-checks `in_progress` status and the current fencing token;
- conditionally creates the `CheckRun` for `runId` so only the first accepted result is canonical;
- marks work `completed` with terminal outcome and completion time;
- for an in-order recurring result, conditionally advances `MonitorStatus`, service rollup, threshold counters, and any incident transition using `(scheduledFor, runId)` as the ordering cursor;
- when a notification-relevant transition occurs, conditionally creates one deterministic transition/activity and one canonical notification outbox item in the same transaction; and
- conditionally removes the terminal work-recovery marker.

A duplicate result for terminal work returns an idempotent already-completed outcome and does not rewrite canonical data. A recurring result with an ordering key not newer than the status cursor is retained as its canonical `CheckRun` and completes its work, but does not change status, counters, service rollup, incidents/activity, or create an outbox item. Transaction cancellation is decoded into typed duplicate, lease-lost, stale-observation, conflict, or retryable-storage outcomes rather than string matching.

Alternative considered: order by `FinishedAt`. Slow older checks can finish after newer checks and incorrectly reverse monitor and incident state. Alternative separate writes expose partial status/incident state and duplicate transitions.

### 7. Manual and recurring observations share persistence but not projections

Manual runs use the same execution envelope, work terminal states, canonical `CheckRun` creation, identity validation, and typed result commit. Their `trigger=manual` remains visible in API responses and history. Manual results do not advance the recurring observation cursor, recurring threshold counters, `MonitorStatus`, service rollup, incident lifecycle/activity, or transition outbox. This prevents an operator's diagnostic request from silently opening, resolving, or delaying recurring health observations.

The synchronous endpoint remains `POST /api/v1/services/{serviceId}/monitors/{monitorId}/run`, requires `Idempotency-Key`, creates/claims work before the HTTP request, and commits through the shared result service. Same-key/same-fingerprint replay resumes or returns the same canonical response without another completed run; same-key/different-fingerprint returns conflict. The response envelope does not change.

Alternative considered: let manual results update only selected latest-status fields. A mixed snapshot becomes ambiguous and can make recurring counters and `LastCheckedAt` disagree about which observation owns state.

### 8. Result commit owns the transition outbox boundary, not notification dispatch

For a notification-relevant incident transition, the causal recurring `runId` deterministically derives one canonical identity value. That value is named `transitionId` in transition/outbox contracts, stored as `activityId` on the incident activity record, and stored as `eventId` where an event-shaped envelope requires that field; all three values MUST be equal. The result transaction conditionally creates exactly one activity and exactly one canonical outbox item carrying this identity, causal `runId`, `trigger=recurring`, `scheduleDefinitionVersion`, `scheduledFor`, incident identity, transition type, and timestamp.

This change stops at the committed pending outbox item. It does not send directly to notification SQS, mark dispatch acknowledged, claim notification/delivery work, initiate routes, or record per-channel outcomes. `assure-notification-and-escalation-delivery` is the sole owner of the dispatcher, dispatch acknowledgement, downstream duplicate suppression, delivery claims, and provider outcomes. There is no direct-send fallback or competing protocol.

The dispatcher reads the sparse, tenant/time-bucketed `DISPATCH_PENDING#<tenant>#<bucket>#<shard>` recovery index populated atomically with the outbox record. The publisher transaction creates the outbox record and the sparse dispatch-pending marker in one conditional transaction; the dispatcher conditionally removes the marker only after conditional SQS acknowledgement. This avoids DynamoDB Streams and the shard-hour cost it would impose on small single-tenant deployments, while keeping the at-least-once retry guarantee through bounded scheduled polling. The pending index uses the same hour bucket the scheduler already uses for execution-marker recovery; reconciliation runs from the scheduler invocation with the same per-tick bucket/shard bounds as publication recovery.

Rollout is dependency-ordered: provision and enable the notification-assurance dispatcher before enabling any retry-safe runtime that can commit outbox items, and deploy producer/dispatcher changes atomically where practical. Once that dispatcher exists, pending canonical outbox items may safely remain pending during temporary dispatcher unavailability and are later dispatched by the same protocol. Recurring execution and manual command exposure remain paused if the dispatcher prerequisite is not met, preventing a silent notification gap.

### 9. Typed failures drive retry and acknowledgement policy

Runtime operations return typed internal failures with stable code, retryability, operation, run identity, and safe details. Expected codes cover malformed envelope, immutable identity conflict, idempotency conflict, not-runnable/superseded skip, work already terminal, claim conflict, lease lost, stale observation, storage unavailable, execution publication failed, and result commit failed. These are internal execution classifications, not additions to the public HTTP error registry unless an API boundary explicitly maps them to an existing public code.

The execution event source enables `ReportBatchItemFailures`. Terminal skips and duplicates omit only their record from failures; retryable storage/result failures and malformed records report only their own SQS item identifiers. Lease conflict/loss stops that record without overwriting state according to the typed policy. `EXECUTION_EVENT_SOURCE_MAX_CONCURRENCY` is a named, finite deployment setting selected against DynamoDB/target capacity and covered by infrastructure tests; increasing batch size must not couple successful records to failed records.

Alternative considered: inspect AWS error strings. String matching is unstable and cannot reliably distinguish safe no-ops from retryable failures.

### 10. Tests model each failure boundary

Repository tests use conditional-operation fakes that preserve work versions and transaction cancellation reasons. Handler tests inject failures before and after work/marker create, execution send, publication acknowledgement, claim, HTTP response, result/outbox transaction, and marker removal. Tests also run duplicate scheduler invocations, duplicate/concurrent SQS deliveries, lease expiry/reclaim, stale fenced completion, out-of-order recurring results, schedule-definition changes, manual idempotency replay/conflict, manual/recurring interleaving, disabled/maintenance/config-change races, and multipage traversal with a middle-page failure.

The acceptance invariant is checked after each injected failure and retry: one stable work identity per `(scheduleDefinitionVersion, scheduledFor)`, no more than one canonical `CheckRun` per run, monotonic recurring projection cursor, at most one equal-valued `transitionId`/`activityId`/`eventId` and outbox item per causal transition, no manual projection effects, and eventual terminal or directly queryable recoverable durable state.

## Risks / Trade-offs

- [HTTP checks can execute more than once after an ambiguous crash] -> Document at-least-once attempts, fence result commits, and guarantee one canonical observation rather than claiming exactly-once side effects.
- [A lease shorter than a valid check can cause unnecessary concurrent retries] -> Derive lease duration from the maximum supported timeout plus a fixed persistence buffer and test boundary values.
- [Configuration changes near a schedule boundary can supersede accepted work] -> Revalidate against current state and record an explicit typed skip instead of silently running stale configuration.
- [Notification dispatcher is unavailable during rollout] -> Enable the notification-assurance dispatcher before producers; retain canonical pending outbox items during temporary unavailability and never fall back to direct send.
- [Conditional transactions and marker items increase write/read cost] -> Keep records in the existing table, use bounded marker-partition queries and point reads, and avoid unbounded tenant scans, new GSIs, queues, or polling services.
- [Old queue messages/work records lack the new identity fields] -> Use the deployment drain procedure rather than adding indefinite compatibility branches.

## Migration Plan

1. Disable recurring execution and manual-run exposure through deployment controls; wait for the execution queue and active workers to drain.
2. Confirm no `in_progress` legacy work remains; legacy audit-only `RUN_REQUEST` records can expire under existing TTL and are not replayed.
3. Provision and verify the `assure-notification-and-escalation-delivery` DynamoDB Stream dispatcher, partial-batch behavior, and existing notification-queue permissions before any new producer can emit outbox items.
4. Deploy shared record/schema changes, retry-safe result/outbox producer, manual result path, scheduler/worker logic, directly queryable markers, and validated timeout/concurrency settings as one coordinated rollout. Do not retain the old direct notification send path.
5. Smoke-test same-key manual replay/conflict, one controlled recurring `scheduledFor`, duplicate execution delivery, marker recovery, and outbox dispatch by the sole assurance dispatcher.
6. Re-enable manual and recurring execution. The first scheduler invocation creates identities only for current eligible schedule definitions/times; it does not backfill historical slots.

Rollback requires disabling recurring and manual execution and draining new-format messages before restoring old code. Pending outbox items are retained for the notification-assurance dispatcher and MUST NOT be deleted or bypassed. New execution/idempotency records are additive in the same table and expire under bounded retention; rollback does not require table restoration.

## Open Questions

None.
