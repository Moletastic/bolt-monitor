## Why

Bolt Monitor can report that a target is down while the monitoring installation itself has silently stopped scheduling checks, accumulated execution work, or failed to deliver notifications. Operators need bounded, conservative evidence that distinguishes target failure from delayed monitoring and failed notification delivery. Native service signals can provide useful early detection, but a persisted health summary must not claim completeness until every growing source has an exact deadline-ordered access path.

## What Changes

- Deliver one OpenSpec change in internal stages: first add native AWS alarms, scheduler heartbeat, structured correlation logs, retention, runbooks, and drills; then add indexed projection maintenance and coverage validation; only then enable the persisted summary, API, and dashboard.
- Treat `make-check-execution-retry-safe` and `assure-notification-and-escalation-delivery` as hard prerequisites for the persisted summary because they own canonical execution, publication, dispatch, retry, lease, and delivery states.
- Add a compact, reconstructable, four-shard tenant due projection ordered by `expectedBy` for enabled-monitor cadence, execution publication/retry/lease work, notification dispatch, and notification delivery when prerequisite artifacts do not already supply an equivalent exact key path.
- Require key-condition queries with explicit page, item, and time budgets. Runtime evaluation never scans a table or tenant, and incomplete, stale, unverified, or budget-exhausted traversal produces `UNKNOWN`/`INCOMPLETE`, never `HEALTHY`.
- Emit structured, secret-free runtime logs with stable `runId`, `incidentId`, notification `transitionId`/`deliveryId`, and queue `sqsMessageId` correlation fields across scheduling, execution, persistence, dispatch, and delivery boundaries.
- Provision a bounded repository-wide default CloudWatch pack coordinated with protected API and auth/RBAC operations. Use native metrics first, avoid per-monitor or other high-cardinality resources, and classify every signal as an optional-SNS alarm or dashboard-only evidence.
- Expose an admin-only installation pipeline health summary that separates target `DOWN`, monitoring `DELAYED`, notification `FAILED`, and evidence `UNKNOWN`/`INCOMPLETE` conditions, with safe evidence and remediation links.
- Keep scheduler heartbeat, execution and notification queue age, DLQs, structured correlation logs, stable runbooks, staging-first failure/recovery drills, and accurate optional external dead-man guidance.
- Add a reproducible FinOps worksheet that keeps the default low-use installation-health pack at or below USD 1 projected incremental monthly cost per persistent stage, excluding optional external delivery fees; document expected and stress-profile costs without presenting them as free-tier defaults.
- Exclude table/tenant scans, per-monitor CloudWatch resources, high-cardinality metrics, automatic DLQ redrive, and a general metrics/log/tracing platform.

## Capabilities

### New Capabilities

- `monitoring-pipeline-observability`: Repository-wide bounded signals, structured correlation logs, alarm/action inventory, retention and cost controls, runbooks, drills, and accurate external dead-man guidance.
- `monitoring-pipeline-health-api`: Exact bounded due projections, conservative completeness semantics, persisted installation health snapshots, and an admin-only health API.

### Modified Capabilities

- `dashboard-web-app`: Add installation pipeline health, completeness, and remediation without conflating target, execution, notification, or evidence states.

## Impact

- Infrastructure: adds a fixed alarm pack, retention, optional SNS actions, evaluator schedule, least-privilege queue/metric reads, and no resource whose count scales with monitors, services, incidents, deliveries, users, or tenants.
- Storage: adds compact sparse due projection and coverage records in the existing table; these are reconstructable and must later be included in recovery inventory/rebuild validation.
- Runtime: adds safe lifecycle logging and transactional due-projection maintenance after the retry-safe and notification-assurance state machines are authoritative.
- API and dashboard: add the protected `GET /api/v1/admin/pipeline-health` surface only after projection readiness is proven; unavailable or incomplete evidence fails closed.
- Operations: adds stable runbooks, alarm transition drills, upper-envelope multi-page tests, and measured stage-attributed cost acceptance.
