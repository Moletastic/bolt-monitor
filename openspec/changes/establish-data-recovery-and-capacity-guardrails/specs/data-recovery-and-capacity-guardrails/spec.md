## ADDED Requirements

### Requirement: System classifies stored data by recovery criticality
System SHALL maintain an authority-specific item-family inventory for both DynamoDB recovery domains. `AppTable` SHALL remain authoritative for durable monitoring configuration, policies, incidents, incident activity, and monitoring audit evidence. `AuthTable` SHALL remain authoritative for immutable Cognito-subject links, application membership/status/role/session boundaries, administrator guards, durable lifecycle desired state, and identity lifecycle audit evidence. Cognito SHALL remain authoritative for provider identity, credentials, password/MFA state, challenges, and token issuance. The inventory SHALL classify projections, sessions, raw runs, and execution work by their actual reconstructable or transient behavior and SHALL NOT merge authority among these domains.

#### Scenario: New persisted item family is introduced
- **WHEN** a change adds an item family to the application table
- **THEN** the change classifies it as durable, reconstructable, or transient
- **AND** documents its authoritative source, retention behavior, and recovery validation

#### Scenario: Authentication data is classified
- **WHEN** the inventory describes authentication-related records
- **THEN** AuthTable immutable subject links, membership status and roles, session-valid-after boundaries, guards, lifecycle desired state, and lifecycle audit evidence are durable application authority
- **AND** Cognito credentials, password/MFA state, challenges, and token issuance remain provider-managed
- **AND** encrypted session/token bundles and auth transactions remain expiring AuthTable state rather than durable authority
- **AND** authentication records are not moved into AppTable during recovery

### Requirement: System defines explicit retention for every item family
System SHALL document retention for every persisted item family in both tables. Raw `CheckRun` records SHALL carry a 30-day TTL, transient `ExecutionWork` records SHALL carry a 7-day TTL, AuthTable sessions/auth transactions and completed lifecycle operation/private invite-intent records SHALL retain the expiry rules defined by the authentication/RBAC capabilities, reconstructable snapshots and projections SHALL exist only while their source exists or until rebuilt, and durable authority records SHALL have no TTL unless a later approved change defines one. DynamoDB TTL deletion SHALL be treated as asynchronous and not as an exact deletion deadline.

#### Scenario: Expiring operational record is written
- **WHEN** the system writes a raw check run or execution work record
- **THEN** it writes a numeric epoch-second `TTL` for the documented 30-day or 7-day window respectively

#### Scenario: Durable record is written
- **WHEN** the system writes durable AppTable monitoring data or durable AuthTable membership, guard, lifecycle, or audit authority
- **THEN** it does not assign the operational `TTL`
- **AND** deletion occurs only through an explicit domain operation or a later approved retention change

#### Scenario: Reconstructable projection becomes stale or missing
- **WHEN** a status, alert, search, scheduler, or rollup projection cannot be trusted after recovery
- **THEN** the system can rebuild or refresh it from durable configuration and retained operational records
- **AND** it does not treat the projection as the sole recovery source

### Requirement: Operators restore each table authority to a new table
System SHALL provide a versioned runbook for `AppTable`-only, `AuthTable`-only, and coordinated recovery. Every selected source SHALL restore to a new table, remain unchanged for rollback, receive settings from the prerequisite stage lifecycle policy, pass authority-specific validation, and switch only its own consumers after an explicit operator decision. The procedure SHALL validate cross-domain references without combining table authority or treating Cognito credentials as recoverable table data.

#### Scenario: Operator performs a recovery
- **WHEN** an operator selects one or both table domains, valid recovery points, and unique target names
- **THEN** the runbook restores each selected table to a new table rather than overwriting its source
- **AND** records each source table ARN, recovery point, target table ARN, region, stage, operator, commands, observed timings, and validation results

#### Scenario: Restored table fails validation
- **WHEN** any required integrity check fails
- **THEN** the runbook prohibits cutover
- **AND** leaves existing consumers on the source table

#### Scenario: Restored table passes validation
- **WHEN** all required integrity checks pass and cutover is approved
- **THEN** the runbook updates each selected table's consumers consistently, verifies authorization, health, and representative reads and writes, and preserves a documented path back to each unchanged source table

#### Scenario: Coordinated recovery spans both tables and Cognito
- **WHEN** an incident requires AppTable and AuthTable recovery
- **THEN** the runbook records explicit recovery points and a cross-domain consistency decision
- **AND** validates Cognito subjects and provider state separately without claiming cross-service atomicity or copying credentials

### Requirement: Recovery validation checks each authority and their boundaries
Recovery validation SHALL verify table/index readiness, key shape, required durable families, counts against authority-specific manifests or documented tolerances, tenant ownership, expiry rules, and representative reads. AppTable validation SHALL cover monitoring/configuration/history references. AuthTable validation SHALL cover immutable subject links, memberships/roles/statuses/session boundaries, active-admin guards, lifecycle operations/audits, and its due-work index. Cross-domain validation SHALL verify `DEFAULT` tenant authorization and Cognito subject references without deriving role from Cognito or moving auth records into AppTable. Offline validation MAY use bounded scans against isolated restore targets; runtime paths SHALL NOT use those scans.

#### Scenario: Recovery validator inspects a restored table
- **WHEN** validation runs before cutover
- **THEN** it emits pass/fail evidence for every required check without logging secrets or sensitive configuration values
- **AND** exits unsuccessfully when a required item family, reference, tenant boundary, key shape, or table/index setting is invalid

#### Scenario: Restored AuthTable is validated
- **WHEN** AuthTable validation runs before cutover
- **THEN** it verifies at least one valid active administrator, membership/guard consistency, session-valid-after boundaries, lifecycle desired state, due-work readiness, and referenced Cognito subjects
- **AND** it does not treat restored expired sessions or Cognito group claims as authorization authority

#### Scenario: Reconstructable records differ at the recovery point
- **WHEN** status, projection, raw-run, or transient-work counts differ from the current source table
- **THEN** validation reports the difference separately from durable-data integrity
- **AND** does not fail solely because data outside the selected recovery point or retention window is absent

### Requirement: System conducts measured non-production recovery drills
System SHALL provide an automated, repeatable non-production drill that restores to a new table, executes recovery validation, exercises a non-production cutover and rollback, and records observed measurements and remediation items. Drill results SHALL be labeled as observations and SHALL NOT be represented as an SLA, RPO, or RTO commitment.

#### Scenario: Recovery drill completes
- **WHEN** the drill is run in an approved non-production stage
- **THEN** evidence records dataset shape, restore duration, validation duration, cutover duration, rollback duration, AWS region, tool version, outcome, and discovered gaps
- **AND** the temporary restore table is removed through an explicit cleanup step after evidence is retained

#### Scenario: Recovery drill misses an internal exercise target
- **WHEN** an observed step exceeds a documented exercise target or fails
- **THEN** the result remains factual and non-contractual
- **AND** a remediation task is recorded before recovery readiness is declared

### Requirement: Growing work is paginated and bounded
Scheduler and API access paths whose result or work set grows with tenants, services, monitors, incidents, activities, audits, policies, channels, or history SHALL use bounded pages and opaque continuation state. Each request or invocation SHALL enforce a configured item, page, response-size, and execution-time budget and SHALL never rely on DynamoDB's implicit 1 MB response boundary as pagination behavior.

#### Scenario: API collection exceeds one page
- **WHEN** a collection contains more records than the endpoint's maximum page size
- **THEN** the endpoint returns only the bounded page and an opaque continuation cursor
- **AND** does not calculate totals by scanning or reading the full collection

#### Scenario: Internal workflow reaches its budget
- **WHEN** a scheduler or maintenance workflow reaches its item, page, or safe remaining-time budget
- **THEN** it persists or emits continuation state and resumes without silently omitting or duplicating logical work

#### Scenario: Runtime access path is reviewed
- **WHEN** an access path can use a key query, sparse projection, batch operation, or incrementally maintained rollup
- **THEN** it does not use a table scan, per-parent N+1 enumeration, or full-history recomputation
- **AND** any exception is measured and justified in the active design

### Requirement: System defines distinct measured operating profiles
The system SHALL document a default low-cost owner profile of one active tenant and up to 10 monitors at five-minute cadence; an expected validation profile of up to 100 services, 100 monitors at 60-second cadence, and 10 concurrent operator requests; and a high-volume stress profile of up to 100 services, 1,000 monitors at 60-second cadence, 30-day raw history, and 25 concurrent operator requests. The stress profile SHALL NOT be represented as the default owner posture, a Free Tier claim, or unlimited scaling.

#### Scenario: Installation is tested with the high-volume stress profile
- **WHEN** the standard load suite runs with 100 services, 1,000 enabled monitors at 60-second cadence, representative 30-day history, and 25 concurrent operator requests
- **THEN** scheduler traversal completes or checkpoints within its configured budgets without starvation
- **AND** API pages remain bounded, DynamoDB reports no throttled requests, queues do not grow without recovery, and no Lambda times out

#### Scenario: Operator exceeds a supported dimension
- **WHEN** configuration or measured workload exceeds a measured profile dimension
- **THEN** the system emits an actionable operational support-boundary warning naming the exceeded dimension
- **AND** bounded processing continues unless a separately documented safety limit requires rejection
- **AND** documentation requires new evidence before expanded support is claimed

#### Scenario: Configuration reaches a hard safety limit
- **WHEN** a request reaches a named transaction, payload, page-size, minimum-cadence, or resource-invariant limit required for correctness or safety
- **THEN** the system rejects it with the documented reason
- **AND** does not present an evidence-only support boundary as the reason for rejection

### Requirement: Load evidence governs capacity and rollup decisions
The repository SHALL include repeatable default low-cost owner, expected validation, and high-volume stress scenarios and SHALL capture request volume, item sizes, consumed read/write capacity for both tables, throttles, queue depth/age, Lambda duration/errors/timeouts, page counts, and response bytes. New indexes, batch reads, or incremental rollups SHALL be added only when measurements show they remove a measured blocker or materially reduce bounded recurring work. Scheduler and pipeline health SHALL reuse one AppTable due-time access pattern; AuthTable lifecycle due-work remains a distinct security workflow index.

#### Scenario: Capacity change is proposed
- **WHEN** measurements show an access path exceeds a guardrail
- **THEN** the implementation records before/after evidence and selects the smallest key, batching, projection, or rollup change that satisfies the envelope
- **AND** keeps the existing table unless evidence demonstrates it cannot satisfy the requirement

#### Scenario: Due-time index is proposed
- **WHEN** scheduler or pipeline-health measurements show existing AppTable keys and indexes cannot satisfy bounded due-time reads
- **THEN** the design records shared before/after evidence for both consumers
- **AND** adds at most one shared sparse due-time index rather than duplicate indexes

#### Scenario: Raw response guardrail overlaps outbound monitoring
- **WHEN** a response or capacity limit concerns bytes downloaded from a monitored outbound HTTP target
- **THEN** this change defers that limit to `harden-outbound-http-monitoring-boundaries`
- **AND** limits here apply only to DynamoDB work and Bolt Monitor API serialization not owned by that change

### Requirement: System documents cost scenarios and optional budget setup
The change SHALL document monthly AWS cost estimates for the default low-cost owner, expected validation, and high-volume stress profiles using stated region, pricing date, cadence, item-size, retention, request, both-table PITR, Cognito/SSM, and restore-drill assumptions. It SHALL disclaim Free Tier eligibility and identify the dominant cost driver. Documentation SHALL recommend optional stage-attributed AWS Budget setup with forecast notification at 80 percent and actual-cost notification at 100 percent. Account-level budget permission, budget amount, and a notification endpoint SHALL NOT be required for a clean default deployment; when enabled, the budget SHALL alert and SHALL NOT automatically disable monitoring.

#### Scenario: Cost model is reviewed
- **WHEN** the implementation is proposed for deployment
- **THEN** reviewers can reproduce the DynamoDB storage/read/write/PITR, Lambda, SQS, API, log, and drill estimates from documented assumptions
- **AND** the model identifies the dominant cost driver and the effect of the high-volume stress profile

#### Scenario: Clean deployment has no budget endpoint
- **WHEN** a deployment does not configure optional AWS Budget permission, amount, or notification destination
- **THEN** deployment succeeds without creating a budget
- **AND** operations documentation explains how to add and verify one later

#### Scenario: Stage approaches its budget
- **WHEN** an enabled optional budget forecasts stage-attributed monthly cost at 80 percent of its configured amount
- **THEN** the configured owner receives an alert identifying the stage and budget

#### Scenario: Stage reaches its budget
- **WHEN** an enabled optional budget reports actual stage-attributed monthly cost at 100 percent of its configured amount
- **THEN** the configured owner receives an alert
- **AND** monitoring continues until an operator makes an explicit change
