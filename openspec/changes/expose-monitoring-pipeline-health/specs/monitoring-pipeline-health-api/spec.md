## ADDED Requirements

### Requirement: Persisted pipeline health has hard prerequisite and readiness gates
The system SHALL enable persisted pipeline health evaluation only after `make-check-execution-retry-safe` and `assure-notification-and-escalation-delivery` are complete and every required due-projection shard and source family is verified `READY` at one active schema generation. Native alarms, structured logs, heartbeat, runbooks, and drills SHALL be deployable before this gate without representing a persisted health summary.

#### Scenario: Early observability stage is deployed
- **WHEN** native alarms, structured logs, scheduler heartbeat, retention, or runbooks are deployed before projection readiness
- **THEN** operators can use those individual signals
- **AND** the system does not persist or display a healthy pipeline summary from partial evidence

#### Scenario: Hard prerequisite is incomplete
- **WHEN** canonical execution deadlines or notification dispatch/delivery deadlines are not authoritative
- **THEN** persisted health evaluation remains disabled
- **AND** the system does not invent a parallel retry or delivery state model

#### Scenario: Projection becomes ready
- **WHEN** all required source families have been migrated or drained through exhaustive key-query paths and all four coverage records match the active generation
- **THEN** persisted health evaluation MAY be enabled

### Requirement: System maintains an exact sparse expected-by projection
The system SHALL maintain compact reconstructable projection items for enabled-monitor cadence, execution publication/retry/lease work, notification dispatch, and notification delivery under four fixed tenant shards. Each item SHALL use `PK=PIPELINE_DUE#<tenantId>#<00..03>` and `SK=DUE#<fixed-width-UTC-expectedBy>#<kind>#<canonicalIdentity>` and SHALL be updated with its canonical DynamoDB transition where both records are controlled by the system.

#### Scenario: Monitor receives a new due deadline
- **WHEN** an enabled non-maintenance monitor is created, enabled, resumed, changes interval, or records qualifying recurring progress
- **THEN** its prior `MONITOR_DUE` key is transactionally replaced with the new `expectedBy` key
- **AND** first-run monitors use their persisted create/enable scheduling time rather than absent history as healthy evidence

#### Scenario: Execution work changes active state
- **WHEN** retry-safe work changes publication-pending, retry/not-before, in-progress lease, or terminal state
- **THEN** the matching `EXECUTION_PUBLICATION`, `EXECUTION_RETRY`, or `EXECUTION_LEASE` projection is inserted, moved, or removed using the authoritative deadline

#### Scenario: Notification work changes active state
- **WHEN** an outbox or delivery changes dispatch, retry, claim, terminal failure, replay, delivered, suppression, or resolution state
- **THEN** the matching `NOTIFICATION_DISPATCH` or `NOTIFICATION_DELIVERY` projection is inserted, moved, retained for unresolved terminal failure, or removed using authoritative state

#### Scenario: Projection stores operational context
- **WHEN** a due projection item is written
- **THEN** it contains only schema/source version, tenant, shard, kind, expected deadline, point-read identities, and safe correlation identifiers
- **AND** it excludes targets, request configuration, notification destinations/configuration, payloads, provider responses, credentials, tokens, and cookies

### Requirement: Runtime health reads use concrete bounded access paths
The evaluator SHALL query each of the four known due-projection primary-index partitions with a sort-key upper bound at evaluation time, `Limit=100`, at most four pages per shard, at most 1,600 evaluated due items overall, and at most 10 seconds of DynamoDB traversal. It SHALL use point reads for scheduler heartbeat, coverage, and latest snapshot and bounded native AWS reads for queue age and DLQ depth. Runtime and migration paths SHALL NOT use table scans, tenant scans, filter-based completeness, or unbounded traversal.

#### Scenario: Due evidence spans multiple pages
- **WHEN** a shard returns `LastEvaluatedKey` and evaluator budgets remain
- **THEN** the evaluator follows the opaque continuation until end-of-results or a named budget is reached
- **AND** it does not treat DynamoDB's implicit 1 MB boundary as completion

#### Scenario: Due evidence fits within all budgets
- **WHEN** every required shard query reaches end-of-results and all required point/AWS reads succeed
- **THEN** the evaluator MAY produce exact counts and a complete stage evaluation

#### Scenario: Traversal budget is exhausted
- **WHEN** any required query retains continuation after its page, item, or time budget
- **THEN** the affected stage is `INCOMPLETE` with an explicitly labeled lower-bound count
- **AND** its externally conservative health state is `UNKNOWN`, never `HEALTHY`

#### Scenario: A source lacks exhaustive indexed migration access
- **WHEN** readiness cannot verify a source family through exact key queries without a table or tenant scan
- **THEN** coverage for that source remains not ready
- **AND** persisted health evaluation remains disabled

### Requirement: System evaluates installation-level pipeline health conservatively
The system SHALL evaluate scheduler, execution, and notification health independently of monitored target outcomes and SHALL require complete, current evidence before reporting any stage healthy.

#### Scenario: Health evaluator completes
- **WHEN** the periodic evaluator has current matching coverage and completes all required reads
- **THEN** it atomically records a timestamped installation health snapshot with freshness, completeness, exact counts, and bounded evidence

#### Scenario: Evaluator read or write fails
- **WHEN** a required DynamoDB or AWS operation fails or returns inconsistent source/projection versions
- **THEN** the evaluator does not overwrite prior evidence as healthy
- **AND** the affected state is exposed as `UNKNOWN`/`INCOMPLETE` with a safe reason

#### Scenario: Snapshot becomes stale
- **WHEN** the latest snapshot is older than three evaluator periods
- **THEN** the API reports stale `UNKNOWN` state rather than healthy

#### Scenario: Projection evidence is incomplete or stale
- **WHEN** coverage is missing, stale, mixed-generation, invalidated after restore, or older than its source version
- **THEN** the affected stage is `UNKNOWN`/`INCOMPLETE`
- **AND** no prior healthy snapshot overrides that result

### Requirement: System derives overdue and stuck states from authoritative deadlines
The system SHALL classify overdue monitoring, publication delay, stuck execution, notification dispatch delay, and notification delivery failure from persisted `expectedBy` values supplied by the authoritative prerequisite state machines rather than duplicating retry constants.

#### Scenario: Enabled monitor is within tolerance
- **WHEN** an enabled non-maintenance monitor has qualifying progress before its cadence, scheduler jitter, and authoritative retry grace expire
- **THEN** it is not counted overdue

#### Scenario: Enabled monitor exceeds tolerance
- **WHEN** recurring execution is enabled and `MONITOR_DUE.expectedBy` passes without qualifying progress or terminal result
- **THEN** execution health reports overdue monitoring with safe monitor/service identity, deadline, last progress, and reason evidence

#### Scenario: Monitor is intentionally not recurring
- **WHEN** the monitor is disabled, in maintenance, or recurring execution is administratively paused
- **THEN** it is excluded from overdue failure counts
- **AND** administrative pause is reported separately

#### Scenario: Execution publication misses recovery deadline
- **WHEN** retry-safe work remains publication-pending after `EXECUTION_PUBLICATION.expectedBy`
- **THEN** execution health reports publication delay with safe `runId` and deadline evidence

#### Scenario: Execution remains retryable
- **WHEN** execution work is before its retry/not-before or active lease deadline
- **THEN** it is not classified as stuck
- **AND** native queue age can still degrade execution health independently

#### Scenario: Execution exceeds retry or lease deadline
- **WHEN** `EXECUTION_RETRY.expectedBy` or `EXECUTION_LEASE.expectedBy` passes while work remains non-terminal
- **THEN** execution health reports stuck work with safe `runId`, stage, age, deadline, and reason evidence

#### Scenario: Notification dispatch misses deadline
- **WHEN** notification outbox work remains unconfirmed after `NOTIFICATION_DISPATCH.expectedBy`
- **THEN** notification health reports dispatch delay with safe `transitionId`, incident, age, deadline, and reason evidence

#### Scenario: Notification delivery fails
- **WHEN** delivery reaches terminal failed/exhausted state, enters the notification DLQ, or remains active after `NOTIFICATION_DELIVERY.expectedBy`
- **THEN** notification health reports failure with safe `transitionId` or `deliveryId`, incident, stage, age, deadline, and reason evidence
- **AND** target status remains unchanged

### Requirement: System exposes protected pipeline health through an admin API
The monitor API SHALL expose read-only `GET /api/v1/admin/pipeline-health` using the standard response envelope, authenticated principal tenant, and current `ADMIN` authorization. It SHALL expose no mutation, replay, raw log, raw message, or raw CloudWatch proxy operation.

#### Scenario: Authorized administrator requests current health
- **WHEN** an authenticated current tenant administrator calls the endpoint
- **THEN** the response includes `evaluatedAt`, freshness, completeness, scheduler, execution, notification, administrative-pause and target summaries, active alarm references when available, overall state, and exact or lower-bound counts
- **AND** each non-healthy or incomplete stage includes a stable reason, oldest relevant time/age, and remediation link

#### Scenario: Caller lacks current administrator authorization
- **WHEN** an unauthenticated, disabled, cross-tenant, non-member, or non-admin caller requests pipeline health
- **THEN** the protected route and current membership authorization fail closed according to the authentication/RBAC contract
- **AND** no pipeline evidence is returned

#### Scenario: Pipeline is healthy but a target is down
- **WHEN** pipeline evidence is complete/current while one or more monitored targets are `DOWN`
- **THEN** target `DOWN` is reported separately
- **AND** execution is not labeled `DELAYED` and notification is not labeled `FAILED` solely from target state

#### Scenario: Monitoring execution is delayed
- **WHEN** heartbeat, due monitor, publication, queue-age, retry, lease, or DLQ evidence exceeds its threshold
- **THEN** the affected monitoring stage is `DELAYED`
- **AND** monitored targets are not inferred `DOWN`

#### Scenario: Notification delivery has failed
- **WHEN** dispatch, queue-age, terminal delivery, stuck delivery, or notification DLQ evidence fails
- **THEN** notification state is `FAILED`
- **AND** target and execution states remain independent

### Requirement: Pipeline health evidence is bounded and secret-free
Persisted snapshots and API responses SHALL expose at most 25 evidence samples per stage and SHALL identify whether counts are exact or lower bounds. They SHALL expose only allowlisted safe identifiers and timestamps.

#### Scenario: Complete unhealthy evidence exceeds sample limit
- **WHEN** complete traversal finds more than 25 unhealthy items for a stage
- **THEN** the response returns the exact aggregate count and at most 25 samples
- **AND** it indicates that additional matching evidence exists

#### Scenario: Evaluator cannot complete unhealthy traversal
- **WHEN** traversal reaches a named budget before end-of-results
- **THEN** the response labels the observed count as a lower bound and the stage `UNKNOWN`/`INCOMPLETE`
- **AND** it does not present that lower bound as an exact total

#### Scenario: Health evidence references runtime work
- **WHEN** the system persists or returns evidence
- **THEN** it may include stable resource IDs, `runId`, `incidentId`, `transitionId`, `deliveryId`, `sqsMessageId`, timestamps, stages, attempts, and reason codes
- **AND** it excludes target URLs, request configuration, notification destinations/configuration, raw payloads, provider responses, credentials, tokens, and cookies

### Requirement: Pipeline health is verified beyond one page
The change SHALL include deterministic tests for evidence spanning multiple pages and shards, budget exhaustion, stale/inconsistent coverage, and the supported upper installation envelope.

#### Scenario: More than one page of monitors is evaluated
- **WHEN** enabled-monitor due evidence spans at least two DynamoDB pages and remains within evaluator budgets
- **THEN** no overdue monitor is omitted or duplicated
- **AND** the resulting count is exact

#### Scenario: Evidence exceeds evaluator capacity
- **WHEN** due evidence exceeds 1,600 items or another named traversal budget
- **THEN** evaluation terminates within its bound
- **AND** reports `UNKNOWN`/`INCOMPLETE` with lower-bound evidence rather than healthy

#### Scenario: Upper-envelope fixture is evaluated
- **WHEN** the 1,000-monitor supported fixture and concurrent publication, lease, dispatch, and delivery failures are exercised
- **THEN** access remains scan-free and within documented page/item/time budgets
- **AND** completeness semantics remain conservative under partial AWS and DynamoDB failures
