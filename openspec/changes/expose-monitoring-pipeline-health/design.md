## Context

The deployed stack has a scheduler, execution and notification SQS queues and DLQs, check and escalation runtimes, one DynamoDB table, a monitor API, and a dashboard. CloudWatch receives native metrics and Lambda logs, but there is no installation health model, fixed alarm inventory, or trusted global access path for overdue work.

`make-check-execution-retry-safe` makes execution identity, publication state, retry/lease deadlines, and terminal state authoritative. `assure-notification-and-escalation-delivery` makes transition dispatch and per-channel delivery state authoritative. They are hard prerequisites for a persisted health summary. Neither prerequisite provides all global deadline-ordered reads needed by an evaluator. `establish-data-recovery-and-capacity-guardrails` later adds a scheduler monitor projection and recovery inventory, but this change cannot depend on that later roadmap item; its own due projection is compact, reconstructable, and must be adopted by that recovery inventory.

Authentication and RBAC protect all `/api/v1/**` routes and resolve current tenant/role. Pipeline health follows that boundary: the API and dashboard are administrator-only, while CloudWatch alarm operations remain deployment-operator AWS permissions. No health evidence contains credentials, targets, destinations, or raw payloads.

## Goals / Non-Goals

**Goals:**

- Detect scheduler absence, overdue monitoring, stuck execution, failed notification dispatch/delivery, queue age, and DLQ work.
- Keep target `DOWN`, monitoring `DELAYED`, notification `FAILED`, administrative pause, and evidence `UNKNOWN`/`INCOMPLETE` distinct.
- Deliver useful native/structured signals before the persisted summary is safe to enable.
- Use exact key access, finite traversal budgets, fixed-cardinality metrics, stable runbooks, and repeatable failure/recovery drills.
- Bound recurring cost and prove behavior beyond one DynamoDB page and through the supported 1,000-monitor envelope.

**Non-Goals:**

- Per-monitor, per-service, per-incident, per-message, per-delivery, per-channel, per-user, or per-tenant CloudWatch resources or dimensions.
- Table scans, tenant scans, full-history recomputation, or treating DynamoDB's 1 MB boundary as completion.
- Changing retry, notification, target status, authentication, or authorization semantics owned by prerequisite changes.
- Automatic DLQ redrive, production mutation from a health read, an APM/tracing platform, or claims of independent assurance from the same AWS account.

## Decisions

### 1. Keep one change with three implementation gates

The change is implemented in this order:

1. **Signals first:** structured allowlisted lifecycle logs, scheduler heartbeat, native queue/Lambda/API alarms, finite retention, runbooks, and drills. These signals do not create or imply a persisted health summary.
2. **Exact access paths:** after both hard prerequisites are complete, deploy due-projection writers, perform a key-query/checkpoint migration, and write coverage records only after every source family and shard is verified. No evaluator snapshot is persisted while coverage is absent, stale, mixed-generation, or failed.
3. **Persisted summary:** enable the evaluator, API, and dashboard only when all projection shards report the active generation `READY`. If readiness is later lost, the evaluator records no healthy replacement and the read surface reports `UNKNOWN`/`INCOMPLETE`.

This sequencing prevents a polished dashboard from turning partial traversal into false confidence. A single simultaneous rollout was rejected because prerequisite records can exist before they are globally queryable.

### 2. Add one compact sparse due-time projection with exact keys

The existing table receives reconstructable projection records, not another source of truth. Four fixed shards are selected by a stable hash of tenant plus canonical identity:

```text
PK = PIPELINE_DUE#<tenantId>#<00..03>
SK = DUE#<fixed-width-UTC-expectedBy>#<kind>#<canonicalIdentity>
```

Each item contains only `schemaVersion`, `tenantId`, `shard`, `kind`, `expectedBy`, canonical IDs needed for a point read, safe correlation IDs, and a source version. It excludes targets, monitor request configuration, channel configuration/destination, payloads, and provider responses. Moving a deadline transactionally deletes the old projection key and writes the new one with the canonical state transition wherever DynamoDB owns both writes. A terminal or no-longer-relevant transition removes the prior key; a terminal notification failure keeps a due entry until replay, suppression, or acknowledged resolution changes its canonical state.

Kinds and exact maintenance rules are:

| Kind | Canonical source | `expectedBy` | Removed or replaced when |
| --- | --- | --- | --- |
| `MONITOR_DUE` | enabled non-maintenance monitor plus latest qualifying recurring progress | first enable/create time or latest qualifying slot plus interval and authoritative grace | disabled, maintenance, recurring pause, or next qualifying progress changes the deadline |
| `EXECUTION_PUBLICATION` | retry-safe work with publication pending | authoritative next publication recovery deadline | publication is acknowledged or work becomes terminal |
| `EXECUTION_RETRY` | pending/retryable execution work | authoritative retry/not-before deadline | claimed, rescheduled, or terminal |
| `EXECUTION_LEASE` | in-progress execution work | authoritative `leaseUntil` | reclaimed, completed, skipped, or lease changes |
| `NOTIFICATION_DISPATCH` | notification outbox not confirmed dispatched | authoritative next dispatch deadline | dispatch is confirmed or terminally quarantined |
| `NOTIFICATION_DELIVERY` | pending/attempted or terminal-failed delivery | authoritative retry/claim expiry, or failure timestamp for terminal failure | delivered, replayed to a new deadline, suppressed, or otherwise resolved |

Scheduler heartbeat is a constant point read at `PK=PIPELINE_HEALTH#<tenantId>`, `SK=HEARTBEAT#SCHEDULER`; the latest snapshot is `SK=SNAPSHOT#LATEST`. Projection readiness uses `SK=COVERAGE#v1#<00..03>` in the same partition. Queue age and DLQ depth use bounded native AWS metric/attribute reads and never enumerate messages.

This change does not consume a fourth GSI. Every due lookup is an exact primary-index `Query` on one known shard with `SK <= DUE#<evaluation-time>#~`; no filter expression or scan is allowed. Four shards are fixed for this supported envelope and are not one shard per tenant or resource.

### 3. Bound traversal and make completeness fail closed

Each evaluator run queries all four known shards with `Limit=100`, at most four pages per shard, at most 1,600 evaluated due items total, and at most 10 seconds of DynamoDB traversal. It samples at most 25 evidence items per stage. It follows `LastEvaluatedKey` only inside those budgets.

A stage is eligible for `HEALTHY` only when all required point reads and every due query reach end-of-results, all four coverage records are `READY` for the active generation, source versions match, and the evidence/evaluation timestamps are within tolerance. If any query returns a continuation after a budget is reached, a coverage record is missing/stale, an AWS read fails, a source version is newer than projection evidence, or migration/reconciliation is unfinished, the affected stage is `INCOMPLETE` and its externally conservative state is `UNKNOWN`. Counts are exact only for completed traversal; incomplete responses expose a lower-bound count and truncation reason, never a false exact total or healthy state.

Coverage activation uses only named key-query paths with explicit cursors and checkpoints. Existing active work that lacks a safe global key path is drained or migrated through its known domain partitions before readiness. If a source family cannot be exhaustively verified without a table or tenant scan, its coverage cannot become `READY`, and stage 3 remains disabled. Runtime evaluation and migration do not use DynamoDB `Scan`.

### 4. Derive health from authoritative deadlines

An enabled, non-maintenance monitor becomes overdue after its persisted `expectedBy`, including first-run enable/create time, cadence, scheduler jitter, and the remaining authoritative execution retry budget. A publication-pending work item is delayed after its publication recovery deadline. Retryable or leased execution is stuck only after its retry deadline or lease expiry. Notification dispatch is delayed after its dispatch deadline; delivery is failed after terminal failure/exhaustion or after a retry/claim deadline passes. Queue age and DLQ evidence can independently degrade the matching stage.

The evaluator does not copy or reinterpret retry constants. It consumes deadlines persisted by the hard-prerequisite state machines. Target outcomes remain independent: stale monitoring never proves a target up or down, and failed delivery never changes target state.

### 5. Use a repository-wide bounded default signal pack

Native AWS metrics are used before custom metrics. The pack has a constant repository-owned inventory; no deployment loop creates alarms from application data. Every alarm description includes stage, owner, and a stable runbook URL. `Optional SNS` means the alarm receives the configured topic action when an ARN is supplied and remains a visible CloudWatch alarm with no action otherwise; this change never creates or requires an external destination.

| Default signal | Source | Default behavior |
| --- | --- | --- |
| Scheduler invocation errors | native Lambda `Errors`, 2 of 5 one-minute periods | Optional SNS alarm |
| Scheduler heartbeat missing | one fixed custom heartbeat metric, 3 consecutive one-minute periods | Optional SNS alarm |
| Execution queue oldest age | native SQS age, >5 minutes for 3 of 5 periods | Optional SNS alarm |
| Execution DLQ visible depth | native SQS depth, >0 for 1 period | Optional SNS alarm |
| Notification queue oldest age | native SQS age, >5 minutes for 3 of 5 periods | Optional SNS alarm |
| Notification DLQ visible depth | native SQS depth, >0 for 1 period | Optional SNS alarm |
| Actionable runtime errors | native Lambda `Errors` across the fixed scheduler, execution, notification, evaluator, and monitor-API roles, 2 of 5 periods | Optional SNS alarm(s), fixed by role manifest |
| Runtime throttles | native Lambda `Throttles` across the same fixed role manifest, 1 of 5 periods | Optional SNS alarm(s), fixed by role manifest |
| Protected API server failures | native API Gateway `5xx`, 2 of 5 periods | Optional SNS alarm |
| Auth key/storage failures | one fixed low-cardinality auth metric because no native domain metric exists, 1 of 3 periods | Optional SNS alarm when auth/RBAC is deployed |
| Sustained auth refresh failures | one fixed low-cardinality auth metric, documented threshold over 5 minutes | Optional SNS alarm when auth/RBAC is deployed |
| Authorization denials, sign-in/recovery events | structured security metrics/log evidence | Dashboard/runbook evidence only; no default alarm |
| Overdue/stuck/terminal-delivery aggregates | evaluator snapshot | Dashboard/runbook evidence only; no default alarm |
| Target `DOWN` | canonical monitor/service status | Dashboard evidence only; not a pipeline alarm |

Custom metric dimensions are limited to fixed `service`, `stage`, `component`, `operation`, and `outcome` values from a source-controlled allowlist. Correlation and domain IDs remain log fields. Missing data is breaching for heartbeat and evaluator readiness, not breaching for sparse event counters unless the paired liveness signal is also absent. Recovery periods are explicit and drills verify both `ALARM` and `OK`.

### 6. Standardize structured correlation logs

JSON lifecycle records contain event name, component, stage, outcome, timestamp, bounded reason code, attempt when relevant, and available `runId`, `incidentId`, `transitionId`, `deliveryId`, and `sqsMessageId`. They cover scheduler materialization/publication/heartbeat, execution claim/attempt/commit/retry/terminal state, notification outbox dispatch, queue consumption, provider acceptance/failure, suppression, and recovery.

An allowlist prohibits target URLs, headers, bodies, expected content, channel destination/configuration, queue bodies, provider bodies, credentials, token/cookie values, and unbounded exception text. Correlation IDs are never metric dimensions.

### 7. Persist and expose one conservative admin summary

After readiness, a once-per-minute evaluator atomically replaces one tenant-scoped latest snapshot. The API re-evaluates time freshness on every read; snapshots older than three evaluator periods are `UNKNOWN`. A failed or incomplete run does not overwrite prior evidence as healthy. It may persist an explicit incomplete result, or allow the prior snapshot to age, but either path must expose the failure reason.

`GET /api/v1/admin/pipeline-health` uses the standard envelope, authenticated principal tenant, and `ADMIN` authorization. It returns independent scheduler, execution, notification, target, pause, and completeness dimensions, active alarm references when available, exact or lower-bound counts, bounded evidence, and stable runbook links. It does not proxy raw logs/CloudWatch or offer replay/redrive.

### 8. Keep runbooks, drills, retention, and external assurance explicit

Application Lambda log groups owned by the stack use a named 14-day default retention for the low-use profile; installations may explicitly select a longer documented profile. Runbooks cover scheduler heartbeat/error, execution age/DLQ, notification dispatch/age/DLQ, projection incompleteness, auth alarm coordination, safe correlation, mitigation, replay/quarantine constraints, escalation, and recovery checks.

Staging-first drills use synthetic non-customer work and verify signal creation, API/dashboard state where stage 3 is enabled, alarm `ALARM`, remediation, alarm `OK`, and recovered health. Internal CloudWatch heartbeat is not independent dead-man assurance. A separately operated external heartbeat is optional and has its own owner, timeout, secret-safe payload, test, and itemized cost.

### 9. Enforce FinOps acceptance

The repository records pricing date, region, item/log sizes, cadence, retention, evaluator queries, custom metrics, alarms, and low/expected/stress request volumes. Resources enabled by default for the low-use owner profile must project at or below USD 1 incremental cost per persistent stage per month; optional SNS deliveries and external dead-man provider fees are excluded from that cap but itemized separately. Expected and 1,000-monitor stress-profile costs are documented separately and are not described as free-tier defaults or blocked by the low-use cap. Staging measurements are normalized to a month and compared with the applicable profile before production enablement.

If the estimate or normalized measurement exceeds the cap, stage 3 is not enabled until log volume, retention, query frequency, or alarm/custom-metric design is reduced through this change or the cap is explicitly revised in OpenSpec. No budget response automatically disables monitoring.

## Risks / Trade-offs

- [Projection drift can create false health] -> Transactional maintenance, generation coverage, source-version checks, fail-closed traversal, and recovery rebuild validation.
- [A widespread outage exceeds evaluator budgets] -> Return lower-bound counts and `UNKNOWN`/`INCOMPLETE`; never scan or claim exact healthy state.
- [The evaluator shares an AWS failure domain] -> Alarm errors and heartbeat absence and document optional external dead-man assurance.
- [Structured logs increase cost or leak data] -> Transition-only events, allowlisted fields, redaction tests, finite retention, and upper-envelope cost acceptance.
- [Auth resources land before or after this change] -> Keep a source-controlled role/signal manifest; enable fixed auth alarms only with auth/RBAC resources, while the API always inherits protected-route and `ADMIN` authorization requirements.
- [Recovery tooling omits the new projection] -> Classify it as reconstructable and require inventory, rebuild, and readiness invalidation after restore.

## Migration Plan

1. Land both hard prerequisites and verify their canonical states/deadlines.
2. Deploy structured logs, native/default alarms, scheduler heartbeat, retention, runbooks, cost worksheet, and signal drills without persisted health claims.
3. Deploy projection writers in shadow mode, migrate or drain pre-projection active work through bounded key-query paths, and verify all four shards/source families for one generation.
4. Mark coverage `READY` only after exact verification, then enable the evaluator and test multi-page and budget-exhausted traversal.
5. Enable the admin API and dashboard; absent, stale, mixed, or incomplete evidence renders `UNKNOWN`/`INCOMPLETE`.
6. Run staging failure/recovery and cost acceptance, then optionally attach an SNS topic and external dead-man service.

Rollback removes dashboard/API exposure, alarm actions, evaluator schedule, and alarms in that order. Projection writers can remain inert and reconstructable. Removing snapshots or projection records never changes target, execution, incident, authorization, or notification state.

## Open Questions

None.
