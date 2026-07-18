## MODIFIED Requirements

### Requirement: System stores raw check execution results
System SHALL conditionally store one canonical raw execution result for each completed stable `runId`.

#### Scenario: Execution completes
- **WHEN** a currently leased execution commits a normalized result
- **THEN** system conditionally persists one `CheckRun` containing run identity, monitor identity, trigger, `scheduleDefinitionVersion` and UTC `scheduledFor` when recurring, timing, and outcome data
- **AND** the record does not require probe location or region identity

#### Scenario: Same result is committed again
- **WHEN** result persistence is retried for a terminal `runId`
- **THEN** system returns the existing canonical result as an idempotent outcome
- **AND** it does not append another raw run

### Requirement: System stores latest monitor status snapshot
System SHALL store a latest recurring monitor-status snapshot whose projection cursor advances only for a newer accepted recurring `(scheduledFor, runId)` ordering key.

#### Scenario: New in-order recurring result is processed
- **WHEN** system commits a recurring result with an ordering key newer than the snapshot's recurring cursor
- **THEN** it conditionally updates current state, recurring counters, last recurring observation metadata, and service rollup

#### Scenario: Recurring result finishes out of order
- **WHEN** system commits a canonical recurring result whose ordering key is equal to or older than the current recurring cursor
- **THEN** it retains the raw `CheckRun` and completes work
- **AND** it does not change monitor status, counters, service rollup, incident lifecycle/activity, or create an outbox item

#### Scenario: Manual result is processed
- **WHEN** system commits a canonical result with `trigger=manual`
- **THEN** it stores the manual `CheckRun`
- **AND** does not advance or rewrite the recurring monitor-status snapshot, counters, service rollup, incident lifecycle/activity, or transition outbox

## ADDED Requirements

### Requirement: Result and terminal state commit atomically
System SHALL commit canonical `CheckRun`, completed work, work-recovery marker removal, applicable recurring projections, and any deterministic incident transition/activity/canonical notification outbox item in one conditional DynamoDB transaction.

#### Scenario: Transaction fails before commit
- **WHEN** any condition or write in result commit fails
- **THEN** no partial canonical run, work completion, marker removal, status, rollup, incident transition/activity, or notification outbox item from that transaction is visible

#### Scenario: Transaction commits
- **WHEN** all identity, lease, uniqueness, and ordering conditions pass
- **THEN** canonical run and terminal work become visible together with all applicable recurring projections

### Requirement: CheckRun preserves observation semantics
System SHALL make trigger and recurring schedule identity explicit on raw run history so consumers can distinguish manual diagnostics from recurring observations.

#### Scenario: Run history mixes trigger types
- **WHEN** clients read manual and recurring `CheckRun` records
- **THEN** each record exposes its stable `runId` and trigger
- **AND** recurring records expose `scheduleDefinitionVersion` and `scheduledFor` while manual records do not claim either
