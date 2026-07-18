## ADDED Requirements

### Requirement: Detailed table controls consume the stage lifecycle baseline
The existing primary DynamoDB application table SHALL retain on-demand capacity and SHALL consume persistent/ephemeral PITR, deletion-protection, retain-on-delete, and ownership-tag behavior from prerequisite `standardize-stage-resource-lifecycle`. This capability SHALL define detailed item-family retention, bounded access, restore validation, and measured capacity evidence without redefining the prerequisite stage policy.

#### Scenario: Recovery guardrails are implemented
- **WHEN** this change applies detailed AppTable recovery and capacity controls
- **THEN** infrastructure uses the lifecycle behavior selected by the prerequisite
- **AND** this change does not introduce a second persistent/ephemeral classification or opt-out mechanism

### Requirement: Storage access patterns expose bounded continuation
Every growing primary-index or secondary-index query SHALL specify an explicit evaluated-item limit, propagate `LastEvaluatedKey` as opaque continuation state, and cap any multi-page internal traversal by item, page, and time budgets. Runtime application access patterns SHALL NOT use table scans.

#### Scenario: Query reaches an evaluated-item limit
- **WHEN** DynamoDB returns a `LastEvaluatedKey`
- **THEN** the caller returns or persists opaque continuation state instead of treating the partial result as complete

#### Scenario: Filtered query returns fewer logical records than requested
- **WHEN** a bounded query filters records after DynamoDB evaluation
- **THEN** the caller follows continuation only within its page and work budgets
- **AND** never issues an unbounded loop to fill a response

### Requirement: Scheduler uses a tenant-scoped monitor projection
The single-table design SHALL maintain a compact, reconstructable, tenant-scoped scheduling projection for enabled monitors so scheduler enumeration and bounded pipeline-health due-time evaluation reuse one key/index access pattern rather than one monitor-list query per service or duplicate due-time indexes. Projection maintenance SHALL be coupled to monitor create, update, enable, disable, move, and delete operations, and the projection SHALL be rebuildable from durable monitor configuration.

#### Scenario: Enabled monitor changes
- **WHEN** an enabled monitor is created or its scheduling fields change
- **THEN** the corresponding scheduler projection is written or updated atomically with canonical configuration where transaction limits permit

#### Scenario: Monitor no longer requires scheduling
- **WHEN** a monitor is disabled or deleted
- **THEN** its scheduler projection is removed from active scheduler reads

#### Scenario: Projection integrity is checked
- **WHEN** recovery validation or a repair tool compares scheduler projections with canonical monitors
- **THEN** missing, stale, and orphaned projection records are reported
- **AND** the tool can rebuild the projection without changing canonical monitor configuration

#### Scenario: Existing access pattern is evaluated
- **WHEN** the scheduler and pipeline-health evaluator require due-time queries
- **THEN** the design first evaluates primary keys and existing sparse indexes with measured bounded-read evidence
- **AND** adds at most one shared sparse due-time index only when the existing table access patterns cannot satisfy the measured criteria

## MODIFIED Requirements

### Requirement: System defines retention strategy for high-volume run data
System SHALL enforce a 30-day retention window for raw check-run history using numeric epoch-second values on the shared DynamoDB `TTL` attribute. Raw runs are reconstructable operational evidence, not durable configuration or an archive, and expiry SHALL NOT remove current status, incidents, or audit evidence.

#### Scenario: Check-run history is persisted over time
- **WHEN** system stores a raw check-run result
- **THEN** the item carries a `TTL` value 30 days after its creation time
- **AND** the primary DynamoDB table is configured so the expired item becomes eligible for asynchronous automatic deletion

#### Scenario: Raw run expires
- **WHEN** DynamoDB removes an expired raw run
- **THEN** durable incident and audit records associated with the monitor remain available
- **AND** current status is not derived by requiring the expired run to remain stored
