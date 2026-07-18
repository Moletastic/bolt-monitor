## Why

Recent check samples and current traffic-light status cannot answer whether recurring monitoring met a user-defined availability target over time, especially when executions are missing or intentionally excluded. Operators need honest objective reporting that accounts for every expected scheduled opportunity, preserves incomplete data, and remains available after raw check-run TTL without presenting the result as an SLA.

## What Changes

- Add optional user-defined availability objectives for recurring-check success and success within a latency threshold, using integer basis-point targets and one bounded rolling-window family.
- Derive expected recurring opportunities independently from immutable effective-dated schedule definitions, enabled intervals, and maintenance intervals. Scheduler work satisfies an expectation but never creates it.
- Add an independently scheduled, cursor-driven finalizer/compactor that boundedly materializes and classifies matured expected slots even when no scheduler work exists, with replay-safe recovery and idempotent aggregates.
- Classify expected opportunities as good, bad, missing, or excluded, with immutable auditable maintenance and disabled intervals; manual checks are excluded by default.
- Treat missing scheduler, queue, execution-provider, or result data as missing and incomplete rather than silently excluding it, while correlating bounded pipeline evidence separately.
- Persist versioned hourly and daily aggregates that survive raw check-run TTL; define a 24-hour late-result correction horizon, bucket closure, and append-only authorized correction policy.
- Report raw counts, exact rational ratios, integer objective precision, allowed/consumed/remaining error budget, completeness, compliance state, and basic burn state using integer arithmetic at every boundary.
- Make maintenance exit return a monitor to `UNKNOWN`/awaiting observation until a new recurring result arrives rather than restoring `UP` from stale state.
- Define service aggregation semantics without conflating objective reporting with recent-sample service-card metrics.
- Add golden timeline and arithmetic tests, plus explicit storage/read/write cost constraints.
- Require the retry-safe execution, notification assurance, pipeline-health evidence, recovery/capacity, and stage-resource-lifecycle changes to land before implementation or rollout.
- Exclude SLA claims, PromQL, arbitrary metric ingestion or query engines, generalized alert-rule engines, status pages, multi-region analysis, and Grafana replacement.

## Capabilities

### New Capabilities

- `availability-observation-windows`: Defines optional availability objectives, recurring-slot accounting, exclusions, durable aggregates, monitor and service evaluation semantics, completeness, error budgets, and burn state.

### Modified Capabilities

- `monitor-configuration`: Allows an optional objective definition while keeping monitor execution valid without one.
- `check-result-status-model`: Associates recurring execution results with stable scheduled slots and selects one canonical result for accounting.
- `monitor-status-read-api`: Exposes awaiting-observation state after maintenance and keeps objective reporting separate from recent raw-run history.
- `incident-management-api`: Prevents maintenance exclusion and maintenance exit from fabricating recovery or resolving an incident from stale success.
- `dynamodb-single-table-storage`: Adds bounded observation aggregate and exclusion/audit item families that outlive raw-run TTL.

## Impact

- **Domain and runtime**: Objective configuration, immutable schedule and interval history, expected-slot derivation, scheduled-slot identity, canonicalization, independent finalization, classification, aggregate compaction, and evaluation rules in shared Go modules and runtime consumers.
- **API**: Monitor and service read surfaces for objective reports and explicit maintenance/disabled accounting metadata, using the existing response envelope and tenant/service ownership boundaries.
- **Storage**: New durable schedule/objective/interval/correction facts, finalizer cursors, expected-slot facts, and bounded versioned hourly/daily aggregate records in the protected primary DynamoDB table; raw `CheckRun` TTL remains unchanged.
- **Dashboard**: Objective configuration and reporting must use objective terminology and remain visually and semantically distinct from recent service-card metrics.
- **Testing and operations**: Golden timeline/math fixtures including a three-slot scheduler outage, retry/duplicate/missing-data cases, late-result and closed-bucket correction tests, finalizer lag/recovery evidence, and cost bounds for expected-slot writes and rolling-window reads.
