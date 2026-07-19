## 1. Execution Identity And Failure Model

- [x] 1.1 Extend shared execution request, work, and result models with required `runId`, accepted time, optional recurring `scheduleDefinitionVersion` and UTC `scheduledFor`, publication state, claim fencing token, lease expiry, attempt count, terminal reason, and causal transition identity.
- [x] 1.2 Implement and unit-test deterministic recurring `runId` derivation from normalized tenant, service, monitor, immutable `scheduleDefinitionVersion`, and UTC `scheduledFor`; assign that one ID before effects and keep manual work schedule-free.
- [x] 1.3 Add a shared internal typed runtime failure model with stable classification, retryability, operation, safe identity details, and helpers for duplicate, conflict, skip, lease-loss, stale-observation, storage, result-commit, and publication outcomes.
- [x] 1.4 Extend AWS facade aliases/helpers needed for conditional updates, transaction condition checks, expression values, consistent reads, pagination keys, and structured transaction-cancellation reason decoding without leaking SDK types into domain services.

## 2. Durable Work State Machine

- [x] 2.1 Update DynamoDB execution work records and canonical key mapping so one `runId` has one addressable work item, retaining TTL and all immutable identity, schedule, publication, lease, attempt, and terminal fields in the existing table.
- [x] 2.2 Implement conditional create that returns created, identical-existing, or typed immutable-identity conflict and never overwrites existing work.
- [x] 2.3 Add directly queryable tenant-scoped, bounded/sharded publication and work-recovery marker item patterns; query configured current/overlap buckets with limits/cursors rather than tenant scans, and conditionally remove or move markers on publication acknowledgement, lease changes, skip, and completion.
- [x] 2.4 Implement conditional claim/reclaim for pending or lease-expired work with a fresh fencing token and lease sized above the maximum supported check timeout plus persistence buffer.
- [x] 2.5 Implement fenced conditional complete and skip transitions so terminal work cannot reopen and an obsolete claimant cannot mutate reclaimed work.
- [x] 2.6 Add repository tests for create races, conflicting identity, duplicate publication acknowledgement, active lease conflict, stale reclaim, fencing-token loss, terminal immutability, bounded marker queries, stale-marker cleanup, conditional terminal removal, pagination, and TTL retention.

## 3. Retry-Safe Scheduler

- [x] 3.1 Replace elapsed `LastExecutionAt` uniqueness with captured-invocation-time `scheduleDefinitionVersion`/UTC `scheduledFor` eligibility and stable `runId` assignment before any scheduler side effect.
- [x] 3.2 Make tenant service and monitor discovery consume every DynamoDB page with a Lambda safety deadline and current enabled/maintenance filtering.
- [x] 3.3 Change per-monitor scheduling to conditionally persist work before SQS send, publish the stable envelope, and conditionally acknowledge publication.
- [x] 3.4 Add a bounded scheduler recovery pass that queries publication markers directly, including successful-send/failed-ack ambiguity, without creating a new run or scanning a tenant.
- [x] 3.5 Preserve earlier per-monitor progress on later persistence/publication failure and return typed retryable context that lets EventBridge retry the same captured event time.
- [ ] 3.6 Add scheduler tests for duplicate/overlapping invocation under one schedule definition/time, schedule-version changes, default interval, no missed later pages, deadline stop, middle-page failure, bounded marker recovery, persisted/send-failed recovery, and send-succeeded/ack-failed duplicate publication.

## 4. Leased SQS Worker

- [x] 4.1 Parse and validate immutable SQS execution envelopes, load canonical work by identity, and reject malformed or conflicting envelopes with typed outcomes.
- [x] 4.2 Claim work before external side effects and handle active claims, terminal duplicates, expired-lease reclaim, and stale fencing tokens according to acknowledgement policy.
- [x] 4.3 Strongly reload current monitor and status after claim; fenced-skip missing, disabled, maintenance, invalid, or recurring work no longer eligible under current interval configuration.
- [x] 4.4 Execute only the current persisted monitor configuration while preserving canonical work identity and trigger in the normalized result.
- [x] 4.5 Return SQS partial batch item failures for only retryable/malformed records and success for completed skips or duplicates; do not couple successful records to another record's failure.
- [ ] 4.6 Add worker tests for duplicate concurrent delivery, no HTTP under active claim, disabled/maintenance/deleted races, current target/config use, superseded schedule definition/time, crash recovery, and stale worker completion rejection.

## 5. Canonical Results And Ordered Projections

- [x] 5.1 Refactor recurring and manual result handling behind one conditional result-commit service that validates work/result identity and current fencing token.
- [x] 5.2 Extend `CheckRun` storage and API mapping with trigger and optional `scheduleDefinitionVersion`/`scheduledFor`, use a uniqueness condition for `runId`, and retain existing raw-run TTL behavior.
- [x] 5.3 Add recurring observation cursor/run identity to `MonitorStatus` and condition recurring projection updates on a strictly newer `(scheduledFor, runId)` ordering key.
- [x] 5.4 Commit work completion, work-marker removal, canonical `CheckRun`, in-order monitor status/counters, service rollup, deterministic incident transition/activity, and one canonical notification outbox item in one DynamoDB transaction.
- [x] 5.5 Resolve duplicate terminal commits idempotently and store out-of-order recurring `CheckRun` history without changing status, counters, rollup, incidents, audit/activity, or outbox items.
- [x] 5.6 Add result repository tests for atomic rollback, duplicate commit, conflicting result identity, lease loss, equal/older/newer ordering keys, first status creation, and service-rollup ordering.

## 6. Incident Transition And Outbox Idempotency

- [x] 6.1 Derive one deterministic transition identity from the causal recurring run; use the same value as `transitionId`, persisted activity `activityId`, and outbox-envelope `eventId`.
- [x] 6.2 Add the canonical notification outbox model/key expected by `assure-notification-and-escalation-delivery`, carrying causal `runId`, `trigger=recurring`, `scheduleDefinitionVersion`, `scheduledFor`, incident identity, transition type, and safe routing context.
- [x] 6.3 Persist exactly one canonical pending outbox item and one sparse dispatch-pending marker atomically with deterministic incident transition/activity and result commit; add no direct notification SQS send, dispatch acknowledgement, delivery claim, route initiation, or per-channel outcome behavior.
- [x] 6.4 Remove the existing ignored direct escalation send from check execution so notification assurance remains the sole dispatcher/consumer protocol.
- [x] 6.5 Add tests for duplicate down/up transition commits, stale result suppression, transaction rollback, equal transition/activity/event identity, one outbox item, and absence of direct notification queue sends.

## 7. Manual Run Unification

- [x] 7.1 Change `POST /api/v1/services/{serviceId}/monitors/{monitorId}/run` to require and validate `Idempotency-Key`, deterministically map key scope to one idempotency-record address, canonicalize/fingerprint the service-scoped command, assign/store one `runId` before effects, conditionally create/claim work before HTTP, and use the shared canonical result commit.
- [x] 7.2 Ensure manual `CheckRun` remains distinguishable in response/history but does not modify recurring status cursor, counters, rollup, incidents, incident audit/activity, or transition outbox.
- [x] 7.3 Persist a bounded-TTL manual idempotency record containing fingerprint, `runId`, and replay/result reference; same-key/same-fingerprint resumes or returns the same response, while same-key/different-fingerprint maps to an existing conflict response code with safe details.
- [x] 7.4 Map other typed manual runtime failures to existing public response-envelope error codes with safe operation and `runId` details where available.
- [x] 7.5 Add manual tests for missing/invalid key, deterministic mapping, same-request in-progress/completed replay, different-payload conflict, retention expiry, duplicate commit, persistence failure, disabled/maintenance race, mixed manual/recurring ordering, and absence of status/counter/rollup/incident/activity/outbox effects.

## 8. Infrastructure And Rollout Safety

- [x] 8.1 Define, validate, document, and boundary-test named configuration satisfying `WORKER_LAMBDA_TIMEOUT > MAX_OUTBOUND_EXECUTION + RESULT_COMMIT_BUFFER`, `EXECUTION_QUEUE_VISIBILITY_TIMEOUT > WORKER_LAMBDA_TIMEOUT + VISIBILITY_MARGIN`, and `WORK_LEASE_DURATION > MAX_OUTBOUND_EXECUTION + RESULT_COMMIT_BUFFER`; keep standard non-FIFO queues and current DLQs.
- [ ] 8.2 Enable SQS `ReportBatchItemFailures`, configure finite `EXECUTION_EVENT_SOURCE_MAX_CONCURRENCY`, and test mixed-success batches plus infrastructure bounds.
- [ ] 8.3 Add bounded recovery bucket/shard/page/deadline configuration and safe structured metrics/logging for created, existing, published, recovered, claimed, reclaimed, skipped, duplicate, stale, completed, marker-cleaned, publication-failed, and dispatch-pending outcomes without recording secrets.
- [x] 8.4 Document the atomic/dependency-ordered deploy sequence: pause recurring and manual producers, drain execution workers, provision and verify the notification-assurance dispatcher first, deploy the sole outbox producer without direct send, smoke-test manual/recurring/outbox flows, then re-enable producers.
- [x] 8.5 Document rollback as pause-and-drain before old code restoration, retain pending canonical outbox records for notification assurance, and confirm legacy TTL records require no table migration, new queue, GSI, or backfill of missed schedules.

## 9. Fault Injection And Verification

- [ ] 9.1 Build stateful DynamoDB/execution-SQS/HTTP fakes that inject failures before and after each owned durable or external boundary and expose structured conditional cancellation reasons.
- [ ] 9.2 Add the full failure/retry matrix for work/marker create, execution send, publication acknowledgement/marker removal, claim/recovery-marker movement, HTTP response, result/outbox transaction, and terminal marker removal.
- [ ] 9.3 Add invariant assertions proving one work/run identity per eligible schedule definition/time, at most one canonical `CheckRun`, monotonic recurring cursor, fenced terminal transitions, no manual recurring effects, and one equal-valued transition/activity/event identity with one canonical outbox item.
- [ ] 9.4 Run `make test-go-all`, `make lint-go`, `make build-go`, `make check-infra`, `make check-bruno`, and strict OpenSpec validation; resolve every regression before marking the change complete.
