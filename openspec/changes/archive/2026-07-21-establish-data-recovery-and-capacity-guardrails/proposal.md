## Why

Bolt Monitor has two application recovery domains: `AppTable` stores monitoring/configuration history, while the separate `AuthTable` stores application identity links, sessions, and the membership/RBAC authority. Cognito remains authoritative for credentials and token issuance. The repository does not yet define detailed retention classes, coordinated restore validation, an operational support envelope, or capacity and cost evidence for growing access paths. Establishing those contracts now prevents data-loss assumptions and unbounded scheduler/API work from becoming production behavior as volume grows.

## What Changes

- Inventory both tables without merging authority: classify durable monitoring/configuration/history families in `AppTable`; classify durable identity links, memberships, roles, guards, lifecycle operations, and audit evidence in `AuthTable`; and classify projections, sessions, raw runs, and transient work by their actual retention and rebuild rules. Cognito credentials remain provider-managed and are never copied into either table as recovery data.
- Consume the basic stage-aware PITR, deletion-protection, and retain-on-delete baseline from prerequisite `standardize-stage-resource-lifecycle`; this change owns detailed item-family retention classes, restore drills, bounded reads, and measured capacity/cost evidence.
- Add an operator runbook for restore-to-new-table recovery of `AppTable` and `AuthTable`, validating each authority independently, checking their Cognito-subject and tenant references, switching each table's consumers safely, and preserving source tables for rollback.
- Require a measured non-production recovery drill and record observed timings and findings without presenting them as an SLA, RPO, or RTO commitment.
- Bound scheduler traversal and API collection reads with continuation, work/time budgets, and explicit failure or resume behavior; avoid table scans, N+1 reads, and full-history recomputation where key-based queries, batch reads, or incremental rollups are practical.
- Define a default low-cost owner profile, an expected validation profile, and a high-volume stress profile for a single-region installation. Treat measured dimensions as warning and operational support boundaries unless a separately documented safety limit requires hard rejection.
- Add raw response-size and consumed-capacity guardrails only where they are not owned by `harden-outbound-http-monitoring-boundaries`.
- Document reproducible low-cost, expected-validation, and high-volume-stress cost scenarios. Document a recommended optional stage-attributed AWS Budget setup without making an account-level budget or notification endpoint a clean-deploy prerequisite.
- Keep the existing single table and on-demand capacity model unless measurements justify a change; do not add multi-region operation, active-active recovery, a separate metrics database, or per-monitor infrastructure resources.

## Capabilities

### New Capabilities

- `data-recovery-and-capacity-guardrails`: Data criticality and retention policy, restore procedure and drill evidence, integrity verification, supported installation envelope, load/capacity acceptance criteria, and cost/budget guardrails.

### Modified Capabilities

- `dynamodb-single-table-storage`: Add detailed `AppTable` item-family retention, bounded-access, and measured capacity requirements while inheriting basic stage lifecycle protection from its prerequisite.
- `check-runtime-scheduler-mode`: Replace unbounded service-to-monitor enumeration with resumable, budgeted pagination that remains correct across scheduler invocations.

## Impact

- Affects shared schema/retention documentation, scheduler repository and orchestration, growing monitor API collection reads, operational runbooks, recovery/load test tooling, and cost documentation. Optional budget setup may add account-level infrastructure when an account and notification destination are explicitly configured.
- Reuses `AppTable`, `AuthTable`, existing GSIs where their access patterns fit, TTL attributes, queues, and Lambdas. It does not merge auth data into `AppTable` or monitoring data into `AuthTable`; any new index requires measured justification and coordination with pipeline-health access patterns.
- Depends on `standardize-stage-resource-lifecycle` for stage-aware PITR/deletion/retain defaults, coordinates with authentication/RBAC changes for `AuthTable` authority and its existing lifecycle due-work index, coordinates with `expose-monitoring-pipeline-health` for due-time evidence, and defers outbound HTTP body limits to `harden-outbound-http-monitoring-boundaries`.
