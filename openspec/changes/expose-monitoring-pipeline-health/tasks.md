## 1. Hard Prerequisites And Staged Gates

- [ ] 1.1 Complete and verify `make-check-execution-retry-safe` canonical execution identity, publication state, scheduled slot, retry/not-before deadline, lease expiry, terminal timestamp, source version, and `runId`; block persisted-summary work until this contract is authoritative.
- [ ] 1.2 Complete and verify `assure-notification-and-escalation-delivery` canonical `transitionId`, per-channel `deliveryId`, dispatch state/deadline, delivery state, retry/claim expiry, terminal classification, source version, and incident correlation; block persisted-summary work until this contract is authoritative.
- [ ] 1.3 Define and test three explicit feature/deployment gates: signals-first, projection-readiness, and persisted-summary/API/dashboard; prove stage 1 cannot claim persisted health and stage 3 cannot enable without every stage-2 coverage record at the active generation.
- [ ] 1.4 Coordinate `GET /api/v1/admin/pipeline-health` with the protected-route manifest and current principal/RBAC contract so only an authenticated `ADMIN` in the resolved tenant can read it.

## 2. Early Native And Structured Signals

- [ ] 2.1 Add a small allowlisted JSON logging helper with bounded reason classifications and safe `runId`, `incidentId`, `transitionId`, `deliveryId`, and `sqsMessageId` fields.
- [ ] 2.2 Instrument scheduler materialization/publication/heartbeat and execution claim/attempt/commit/skip/retry/terminal transitions without monitor targets, request configuration, raw queue bodies, or unbounded errors.
- [ ] 2.3 Instrument notification outbox dispatch, queue consumption, delivery claim/attempt/retry/provider acceptance/terminal failure, suppression, and recovery without destinations, channel config, provider payloads, or provider responses.
- [ ] 2.4 Add log-shape and redaction tests seeded with recognizable targets, headers, expected content, channel destinations/config, queue bodies, provider errors/responses, credentials, tokens, and cookies.
- [ ] 2.5 Emit one fixed scheduler heartbeat signal and verify it is independent of dashboard traffic and cannot use a monitor, service, tenant, run, incident, delivery, channel, user, URL, or error-text metric dimension.

## 3. Exact Sparse Due Projection

- [ ] 3.1 Add compact reconstructable records with `PK=PIPELINE_DUE#<tenant>#<00..03>` and `SK=DUE#<fixed-width-UTC-expectedBy>#<kind>#<canonicalIdentity>`, plus point-addressed heartbeat, coverage-generation, and latest-snapshot records under `PIPELINE_HEALTH#<tenant>`.
- [ ] 3.2 Transactionally maintain `MONITOR_DUE` projection entries for create/enable, interval or maintenance changes, recurring pause/resume, qualifying progress, terminal result, disable, move, and delete, including a first-run `expectedBy`.
- [ ] 3.3 Transactionally maintain `EXECUTION_PUBLICATION`, `EXECUTION_RETRY`, and `EXECUTION_LEASE` entries when retry-safe work changes publication, retry/not-before, claim/lease, or terminal state.
- [ ] 3.4 Transactionally maintain `NOTIFICATION_DISPATCH` and `NOTIFICATION_DELIVERY` entries when outbox dispatch or assured delivery changes pending, retry, claim, terminal-failure, delivered, replay, suppression, or resolution state.
- [ ] 3.5 Implement four exact primary-index due queries using key conditions only, `Limit=100`, at most four pages per shard, at most 1,600 evaluated items, and at most 10 seconds; prohibit `Scan`, filter-based completeness, unbounded loops, and reliance on DynamoDB's implicit 1 MB boundary.
- [ ] 3.6 Implement a resumable, key-query-only migration/readiness procedure for each canonical source family; drain unsupported legacy active work rather than scanning, record `READY` coverage only after complete generation verification, and leave readiness unset when any source lacks an exhaustive key path.
- [ ] 3.7 Add projection reconciliation/invalidation hooks for restore and schema-version changes and document this projection as reconstructable evidence for `establish-data-recovery-and-capacity-guardrails` inventory and rebuild checks.
- [ ] 3.8 Add repository tests for transactional old-key removal/new-key insertion, disabled/maintenance removal, publication acknowledgement, lease movement, delivery replay/suppression, mixed generations, missing coverage, source-version drift, and no runtime or migration `Scan` call.

## 4. Conservative Persisted Health Evaluation

- [ ] 4.1 Define shared health types for scheduler, execution, notification, target, administrative pause, freshness, completeness, reason codes, exact/lower-bound counts, and overall state; cap samples at 25 per stage.
- [ ] 4.2 Implement the once-per-minute evaluator only behind projection readiness, using exact due queries, scheduler heartbeat point read, and bounded native execution/notification queue age and DLQ attributes.
- [ ] 4.3 Derive monitor, publication, execution retry/lease, notification dispatch, and delivery outcomes from authoritative persisted `expectedBy` values without duplicating prerequisite retry constants.
- [ ] 4.4 Require end-of-results on every required shard/query plus current `READY` coverage before a stage can be `HEALTHY`; map continuation, budget exhaustion, stale/missing evidence, mixed generation, source-version drift, and partial AWS/DynamoDB failure to `INCOMPLETE` and externally `UNKNOWN`.
- [ ] 4.5 Persist one tenant latest snapshot atomically only after evaluation; ensure failure cannot replace prior evidence as healthy and snapshots older than three evaluator periods become `UNKNOWN` at read time.
- [ ] 4.6 Add tests for first-run overdue monitors, retryable work inside budget, expired publication/retry/lease deadlines, notification dispatch delay, terminal/retry-exhausted delivery, queue age, both DLQs, administrative pause, and target-state independence.
- [ ] 4.7 Add failure and scale tests with due records spanning at least two DynamoDB pages, all four shards, more than the 1,600-item evaluator budget, more than one page of monitors, and the supported 1,000-monitor envelope; assert complete traversal is exact and every truncated/stale/failed traversal is `UNKNOWN`/`INCOMPLETE`, never `HEALTHY`.

## 5. Protected Pipeline Health API

- [ ] 5.1 Add a monitor-api repository/service point read for the latest snapshot that recalculates freshness and projection readiness at request time and returns bounded secret-free evidence, exact or lower-bound counts, completeness reasons, and stable runbook links.
- [ ] 5.2 Add authenticated, `ADMIN`-authorized `GET /api/v1/admin/pipeline-health` using the standard envelope and principal tenant, with no mutation, replay, raw log, raw message, or raw CloudWatch proxy behavior.
- [ ] 5.3 Add API tests for healthy pipeline with target down, delayed execution without inferred target down, failed notification without target change, administrative pause, stale/absent/mixed/incomplete evidence as `UNKNOWN`, lower-bound/truncated counts, authorization denial, cross-tenant denial, and redaction.
- [ ] 5.4 Wire the route through the protected SST route helper, add exact Bruno tags/docs and authentication, update OpenAPI, and add route/auth/contract drift tests.

## 6. Repository-Wide Bounded CloudWatch Pack

- [ ] 6.1 Encode the source-controlled default inventory from the design, including scheduler errors/heartbeat, execution and notification queue age, both DLQs, fixed-role Lambda errors/throttles, protected API 5xx, and conditional fixed auth key/storage and sustained-refresh alarms.
- [ ] 6.2 Use native AWS metrics for every available signal and add only fixed scheduler-heartbeat and auth-domain custom metrics; reject all correlation/domain/resource/error-text dimensions and any resource generated from application rows.
- [ ] 6.3 Configure each inventory entry explicitly as `Optional SNS alarm` or `Dashboard/runbook evidence only`; support one optional SNS topic ARN without creating an external destination, and expose/document the no-action state when absent.
- [ ] 6.4 Configure documented thresholds, missing-data semantics, recovery periods, stage/owner/runbook descriptions, and a named 14-day low-use retention default for application Lambda log groups owned by the stack, with longer retention only through an explicit cost profile.
- [ ] 6.5 Add infrastructure tests proving the alarm/log-group/custom-metric count is fixed by the repository role manifest as monitor, service, incident, delivery, channel, user, and tenant fixtures grow.
- [ ] 6.6 Add auth/RBAC coordination tests proving fixed auth alarms appear only with auth resources, authorization denials remain dashboard/runbook evidence by default, and the health route never bypasses JWT plus current membership authorization.

## 7. Dashboard Pipeline Summary

- [ ] 7.1 Add typed server-side pipeline-health integration preserving `Result`, API boundary, clock, and date-fns conventions.
- [ ] 7.2 Show freshness and separate scheduler, execution `DELAYED`, notification `FAILED`, target `DOWN`, administrative pause, and evidence `UNKNOWN`/`INCOMPLETE` states without changing service/monitor traffic lights.
- [ ] 7.3 Show exact or labeled lower-bound aggregate counts, at most the bounded safe sample, coverage/truncation reasons, and supplied remediation links; never render unavailable or incomplete evidence as healthy.
- [ ] 7.4 Add dashboard tests for independent failure domains, target down with healthy pipeline, delayed execution, failed notification, stale/missing/mixed/incomplete evidence, no-services state, multi-page lower-bound evidence, responsive rendering, accessible text/links, and admin authorization behavior.

## 8. Runbooks, Drills, And FinOps

- [ ] 8.1 Add stable scheduler, execution queue/DLQ, notification dispatch/queue/DLQ, projection-incomplete, and auth-coordination runbooks covering symptoms, safe correlation queries, mitigation, replay/redrive/quarantine constraints, escalation, and recovery.
- [ ] 8.2 Add staging-first synthetic failure/recovery drills for scheduler error/missing heartbeat, execution queue age/DLQ, notification dispatch/queue age/DLQ, projection incompleteness, and applicable auth alarms; verify `ALARM` then `OK`, safe logs, and API/dashboard recovery where enabled.
- [ ] 8.3 Document optional external dead-man ownership, timeout, secret-safe payload, test, and cost, explicitly stating no independent assurance exists when unconfigured and internal CloudWatch is not external assurance.
- [ ] 8.4 Produce a pricing-date/region worksheet for low, expected, and 1,000-monitor upper-envelope usage covering logs, retention, alarms, custom metrics, evaluator Lambda, DynamoDB projection reads/writes/storage, SQS/API reads, optional SNS delivery, and external provider cost.
- [ ] 8.5 Assert the default low-use health pack remains at or below USD 1 projected incremental monthly cost per persistent stage, excluding but separately itemizing optional external delivery fees; document expected and stress-profile costs separately, require explicit opt-in above the default cap, and never auto-disable monitoring as a cost response.

## 9. End-To-End Verification

- [ ] 9.1 Run `make test-go-all`, `make lint-go`, and `make build-go` and resolve all runtime, projection, evaluator, API, redaction, authorization, and multi-page failures.
- [ ] 9.2 Run `make lint-dashboard`, `make check-dashboard`, `make test-dashboard`, and `make build-dashboard` and resolve pipeline summary regressions.
- [ ] 9.3 Run `make check-infra`, `make format-infra`, `make check-bruno`, and OpenAPI checks and resolve route, alarm-inventory, action, retention, auth, and contract drift.
- [ ] 9.4 Deploy each gate to staging in order, capture projection readiness plus alarm failure/recovery evidence, run upper-envelope cost/load acceptance, and verify no real customer notification or secret-bearing telemetry is produced.
- [ ] 9.5 Run `openspec validate expose-monitoring-pipeline-health --strict` and confirm implementation matches every scenario without scans, false healthy states, per-resource CloudWatch infrastructure, or a general telemetry platform.
