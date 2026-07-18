## 1. Durable Identity And Storage

- [ ] 1.1 Gate implementation on the corrected `make-check-execution-retry-safe` contract for one deterministic canonical transition/outbox record created atomically with result and incident state; verify direct post-commit notification publication is removed so this change is the sole dispatcher.
- [ ] 1.2 Add shared versioned canonical-dispatch, scheduled-step, replay-command, escalation-plan, exact delivery-state (`pending`, `in_flight`, `retryable_failed`, `ambiguous`, `delivered`, `terminal_failed`), outcome-classification, and safe provider-metadata models with deterministic delivery ID helpers and unit tests.
- [ ] 1.3 Add DynamoDB mappings that consume the canonical transition record and provide a sparse tenant/time-bucketed pending-dispatch access path, acknowledged-dispatch TTL, incident-scoped deliveries, replay commands, and bounded replay-idempotency records; add secret-free marshal/unmarshal and no-scan query tests.
- [ ] 1.4 Add repository operations to conditionally acknowledge canonical dispatch, list bounded pending buckets, reconcile one `eventId`, create/read immutable route plans, conditionally create/lease deliveries, complete exact outcomes, advance a step once, suppress escalation, list incident deliveries, and prepare idempotent replay; cover ownership, concurrency, and fencing.

## 2. Transactional Outbox Dispatch

- [ ] 2.1 Implement the sole transition dispatch path by consuming filtered inserts of retry-safe canonical outbox records, sending canonical versioned messages to the existing notification queue, and conditionally changing the same record from pending to acknowledged; test duplicate and ambiguous enqueue acknowledgement.
- [ ] 2.2 Implement Stream partial-batch responses using sequence-number item identifiers so only queue-send/repository failures retry and mixed-success batches do not replay successful records.
- [ ] 2.3 Preserve exhausted Stream records as durable pending work, emit source-kind-safe DLQ identity, alarm, and implement bounded recent-bucket plus point-by-`eventId` reconciliation/manual repair without a table scan; test Stream exhaustion, dispatcher/reconciler races, and repair.
- [ ] 2.4 Configure the primary table stream, insert-only canonical-record filter, escalation-runtime subscription, queue-send permission, `ReportBatchItemFailures`, bounded retries, sparse pending access path, bounded reconciler, and existing-DLQ failure destination in SST, with infrastructure assertions for filters, bounds, and resource scoping.

## 3. Provider Outcome Classification

- [ ] 3.1 Replace opaque notification sender errors with typed sanitized outcomes that classify accepted, timeout, transport, `429`, `5xx`, terminal `4xx`, invalid configuration, and unsupported channel results.
- [ ] 3.2 Update Telegram, email, SMS, webhook, and PagerDuty senders to return safe status/request-ID/retry-after metadata, omit raw bodies/headers/targets/secrets, and pass the stable delivery ID as a provider idempotency or deduplication key where supported.
- [ ] 3.3 Add sender tests for every retry class, provider acceptance boundary, malformed configuration, retry-after parsing/bounds, idempotency headers or fields, and redaction of credential-bearing responses and URLs.
- [ ] 3.4 Adapt notification channel test-send behavior to the typed sender contract while preserving its stateless execution and existing sanitized API/audit semantics.

## 4. Idempotent Delivery Orchestration

- [ ] 4.1 Refactor incident-down handling to persist escalation state and an immutable selected route plan before provider I/O, then create deterministic per-step/per-channel pending deliveries.
- [ ] 4.2 Implement fenced lease claims and exact transitions among `pending`, `in_flight`, `retryable_failed`, `ambiguous`, `delivered`, and `terminal_failed`; recover expired in-flight attempts as ambiguous and reject claims for active leases or terminal states.
- [ ] 4.3 Implement per-channel processing so confirmed transient outcomes become `retryable_failed`, uncertain acceptance becomes `ambiguous`, provider acceptance becomes `delivered`, and non-retryable outcomes become `terminal_failed`; retries send only unfinished channels and never resend delivered channels.
- [ ] 4.4 Use named automatic-attempt and SQS receive limits to terminalize known unfinished deliveries as `terminal_failed/retry_exhausted` before still failing the source record for notification-DLQ redrive; test timeout/transport/`429`/`5xx`, ambiguity, exhaustion, and terminal config/`4xx` acknowledgement.
- [ ] 4.5 Return SQS partial batch responses by message ID for malformed, unsupported-version, unsupported-kind, repository-failed, retryable, ambiguous, and mixed-success records so poison messages reach the existing DLQ without replaying successful records.
- [ ] 4.6 Re-read incident and escalation state before every scheduled step and attempt claim, suppressing future work after recovery without inventing a delivery state; test resolved, acknowledged/open, duplicate, historical delivered, and unfinished delivery cases.
- [ ] 4.7 Advance each step and create escalation-exhausted state/incident exactly once after all channel deliveries are terminal (`delivered` or `terminal_failed`), with tests for duplicate workers, transient/ambiguous active outcomes, partial failure, and recovery races.
- [ ] 4.8 Define named provider timeout, completion buffer, Lambda timeout, termination buffer, claim-start budget, attempt lease, redelivery buffer, SQS visibility/max receives, retry backoff/attempt limit, and Scheduler retry age/attempt constants; add infra/unit assertions for every required inequality and terminalization boundary.

## 5. Self-Cleaning One-Time Scheduling

- [ ] 5.1 Replace EventBridge rule/Lambda permission scheduling clients with an AWS Scheduler facade that builds deterministic bounded names and idempotently creates `at` schedules with flexible windows off and action-after-completion deletion.
- [ ] 5.2 Configure schedules to target the existing notification queue with canonical identity and `sourceKind=scheduler_target`, bounded retry age/attempts, the existing notification DLQ, and conflict detection tests for duplicate names with mismatched inputs.
- [ ] 5.3 Provision a managed schedule group and dedicated execution role limited to the notification queue/DLQ, and scope runtime `scheduler:*` plus `iam:PassRole` permissions to that group and role; add infra tests that reject wildcard or Lambda invoke permissions.
- [ ] 5.4 Retain a temporary adapter for already-created legacy direct scheduled payloads that re-enqueues canonical work, and verify the runtime no longer calls `PutRule`, `PutTargets`, `AddPermission`, or STS for new schedules.
- [ ] 5.5 Add scheduler and handler tests proving duplicate invocation safety, completed-schedule deletion configuration, target retry/DLQ settings, annual replay elimination, and recovery suppression.

## 6. Delivery API And Dashboard

- [ ] 6.1 Add typed shared/API errors and monitor-api repository/response mapping for the exact delivery-state enum, incident delivery list, replay eligibility, required `Idempotency-Key`, and idempotency conflict without exposing raw provider or channel configuration data.
- [ ] 6.2 Add `GET /api/v1/incidents/{incidentId}/deliveries` with tenant/incident validation, stable chronological ordering, empty-state behavior, standard envelope, and handler/repository tests.
- [ ] 6.3 Add `POST /api/v1/incidents/{incidentId}/deliveries/{deliveryId}/replay` using a required `Idempotency-Key` and one conditional transaction for eligible `terminal_failed` delivery state, audit metadata, bounded idempotency record, and canonical `delivery_replay` dispatch record; test same-key/same-request replay, payload mismatch conflict, expiry, concurrent requests, enqueue durability, delivered/active records, cross-tenant access, and recovered incidents.
- [ ] 6.4 Wire both routes in SST and add Bruno requests with required tags/docs and success, not-found, and replay-conflict expectations.
- [ ] 6.5 Add dashboard Result-based API adapters/types and a server action for replay that follows the response-envelope, typed-error, and router conventions.
- [ ] 6.6 Extend the incident escalation tab with all six exact per-transition/step/channel delivery states, separate recovery-suppression eligibility, provider-acceptance wording, sanitized outcomes, attempts/timestamps, loading/empty/error states, and replay controls only for eligible terminal failures.
- [ ] 6.7 Add dashboard tests for partial multi-channel success, all six delivery states, no-human-receipt wording, replay pending/result feedback, suppression outside the delivery enum, accessibility, and secret redaction.

## 7. Contracts, Observability, And Operations

- [ ] 7.1 Update `openapi/openapi.yaml` with delivery list/replay paths, exact six-state enum, required `Idempotency-Key`, same-request replay response, mismatch conflict, bounded-retention semantics, safe metadata, provider-acceptance semantics, and typed errors; run the OpenAPI validation workflow.
- [ ] 7.2 Add structured secret-free logs and bounded metrics for outbox dispatch, normalized delivery attempts/outcomes, retries/exhaustion, suppression, schedule failures/conflicts, and replay, with tests that sensitive values never appear.
- [ ] 7.3 Add `docs/runbooks/notification-delivery.md` covering delivery/activity correlation, logs/metrics, source-kind-aware notification-DLQ inspection, point/recent-bucket outbox reconciliation, manual repair, configuration correction, ambiguous outcome handling, idempotent API replay, redrive restrictions/quarantine, and rollback.
- [ ] 7.4 Add a dry-run-by-default scoped cleanup utility and runbook procedure that inventories then removes only legacy `esc-*-step-*` targets/rules and matching `allow-events-*` Lambda statements, with fixture tests for unrelated-resource preservation.
- [ ] 7.5 Document the bounded cost model for Stream invocations, pending-index reconciliation, delivery/replay/idempotency writes, retries, short-lived schedules, TTL, and incident-partition reads, plus explicit exclusions for scans, new integrations, on-call behavior, and human-receipt guarantees.

## 8. Verification

- [ ] 8.1 Run `make test-go-all`, `make lint-go`, and `make build-go`; resolve all delivery, runtime, repository, and sender failures.
- [ ] 8.2 Run `make test-dashboard`, `make lint-dashboard`, `make check-dashboard`, and `make build-dashboard`; resolve all incident delivery UI and API adapter failures.
- [ ] 8.3 Run `make check-infra`, `make lint-infra`, and `make check-bruno`; verify route coverage, Scheduler least privilege, source-aware Stream/SQS/Scheduler redrive, timeout/retry ordering assertions, reconciliation bounds, and formatting.
- [ ] 8.4 Run OpenSpec strict validation and review the final diff for changes limited to the approved capabilities, secret-free examples/logging, unchanged integration count, reused queues/table, and complete runbook/test evidence.
