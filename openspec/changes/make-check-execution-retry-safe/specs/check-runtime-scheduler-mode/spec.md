## MODIFIED Requirements

### Requirement: Scheduler sends messages to SQS queue
System SHALL durably create execution work before publishing the corresponding stable execution envelope to the existing SQS execution queue.

#### Scenario: Scheduler sends to SQS
- **WHEN** scheduler accepts execution work for an enabled eligible monitor
- **THEN** it conditionally writes one `RUN_REQUEST#` work item and directly queryable recovery markers with a stable `runId`
- **AND** sends an envelope carrying that same identity to the SQS execution queue

#### Scenario: Existing work is encountered on retry
- **WHEN** scheduler retries the same monitor, `scheduleDefinitionVersion`, and `scheduledFor`
- **THEN** conditional create resolves to the existing identical work
- **AND** scheduler publishes only as required by the work's durable publication state

### Requirement: Scheduler maintains audit trail in DynamoDB
System SHALL use the persisted `RUN_REQUEST#` item as the durable coordination and troubleshooting record for every accepted execution.

#### Scenario: DynamoDB work record is created
- **WHEN** scheduler accepts a recurring execution before SQS publication
- **THEN** it writes a `RUN_REQUEST#` item containing tenant, service, monitor, `runId`, trigger, `acceptedAt`, `scheduleDefinitionVersion`, UTC `scheduledFor`, status, publication state, and TTL
- **AND** immutable identity fields cannot be replaced by a conflicting retry

### Requirement: Scheduler handles SQS send failures gracefully
System SHALL preserve accepted work and return a typed retryable publication failure when SQS send fails.

#### Scenario: SQS send failure
- **WHEN** scheduler persists work but fails to send its envelope to SQS
- **THEN** it leaves that work publication-pending with a directly queryable publication marker
- **AND** returns a typed retryable error with safe run and monitor context
- **AND** EventBridge retry or recovery republishes the same run rather than creating another run

#### Scenario: SQS send succeeds but acknowledgement fails
- **WHEN** SQS accepts the message but publication acknowledgement cannot be persisted
- **THEN** work remains recoverable with the same identity
- **AND** a later duplicate send remains safe

### Requirement: Scheduler respects per-monitor intervalSeconds
System SHALL determine recurring eligibility by the monitor's immutable current schedule-definition version and a UTC `scheduledFor` boundary rather than using queue-send time as the idempotency authority.

#### Scenario: Monitor is eligible at current scheduled time
- **WHEN** an enabled non-maintenance monitor has no accepted recurring work for its current `scheduleDefinitionVersion` and `scheduledFor`
- **THEN** scheduler accepts one recurring work identity for that tuple

#### Scenario: Monitor already has work for current schedule identity
- **WHEN** scheduler evaluates a monitor for a schedule definition/time whose stable work identity already exists
- **THEN** it does not create a second run
- **AND** it may recover publication for the existing nonterminal run

#### Scenario: Monitor is not due for a new scheduled time
- **WHEN** scheduler evaluates a monitor before its next effective interval boundary
- **THEN** scheduler does not create work for a future `scheduledFor`

#### Scenario: Monitor has zero intervalSeconds
- **WHEN** scheduler encounters a legacy monitor where `intervalSeconds` is zero or absent
- **THEN** it uses the existing default 60-second effective interval in the immutable schedule definition
- **AND** retries for that definition and `scheduledFor` still derive one identity

### Requirement: Scheduler records LastExecutionAt after enqueuing
System SHALL NOT use mutable `LastExecutionAt` as the uniqueness or retry-safety authority for recurring execution.

#### Scenario: Work is accepted
- **WHEN** scheduler conditionally creates or finds recurring work for a schedule definition/time
- **THEN** `scheduleDefinitionVersion` and `scheduledFor` determine whether the run already exists
- **AND** failure to update optional display metadata does not create another run

#### Scenario: Queue publication fails
- **WHEN** accepted work cannot be published
- **THEN** the durable work remains eligible for publication recovery regardless of `LastExecutionAt`

## ADDED Requirements

### Requirement: Scheduler assigns stable recurring identity before side effects
System SHALL capture one invocation time and derive immutable `scheduleDefinitionVersion`, UTC `scheduledFor`, and deterministic `runId` before the first write or queue send for each recurring execution.

#### Scenario: EventBridge retries one invocation
- **WHEN** the same scheduler event is retried for the same monitor and schedule definition/time
- **THEN** scheduler derives the same schedule-definition/time identity and `runId`
- **AND** all durable and queued representations use those values

#### Scenario: Scheduler invocation spans monitor pages
- **WHEN** scheduler processes multiple pages in one invocation
- **THEN** it uses the same captured invocation time for `scheduledFor` calculation across all pages

### Requirement: Scheduler paginates monitor discovery
System SHALL consume DynamoDB pagination for tenant services and monitors and SHALL bound processing before the Lambda execution deadline.

#### Scenario: Enabled monitors span multiple pages
- **WHEN** a tenant's services or monitors produce a `LastEvaluatedKey`
- **THEN** scheduler continues from that key until all available pages for the invocation are processed or the safety deadline is reached
- **AND** it does not silently omit enabled monitors from later pages

#### Scenario: Safety deadline is reached
- **WHEN** insufficient Lambda time remains to safely materialize and publish another monitor
- **THEN** scheduler stops at a durable boundary and returns a typed retryable partial-progress result

### Requirement: Scheduler preserves partial progress
System SHALL process monitor work independently so a later failure does not undo or duplicate earlier accepted work.

#### Scenario: Middle-page monitor fails publication
- **WHEN** earlier monitors have durable work and a later monitor fails during persistence or publication
- **THEN** earlier work remains valid
- **AND** scheduler reports typed page and monitor context for retry
- **AND** retry converges on existing identities before continuing remaining work

### Requirement: Scheduler recovery uses bounded marker queries
System SHALL recover publication-pending execution work through directly queryable, tenant-scoped marker partitions with configured shard, bucket-overlap, page, and invocation-deadline bounds.

#### Scenario: Publication recovery runs
- **WHEN** scheduler searches for persisted work whose queue acceptance is unacknowledged
- **THEN** it queries configured marker partitions and pages rather than scanning tenant work
- **AND** conditionally removes an obsolete marker only after canonical work proves publication-complete or terminal

### Requirement: Scheduler revalidates recurring eligibility
System SHALL create recurring work only for a monitor that is currently enabled and not in maintenance at evaluation time.

#### Scenario: Monitor is disabled or in maintenance
- **WHEN** scheduler evaluates current monitor and status state
- **THEN** it creates no recurring work for that monitor at the current `scheduledFor`
