## ADDED Requirements

### Requirement: System stores availability accounting item families
The DynamoDB single-table design SHALL define tenant-aware item families and canonical key patterns for immutable schedule definitions, objective versions, enabled/disabled and maintenance interval references, expected-slot facts, canonical slot results, authorized corrections, finalizer cursors, and versioned hourly and daily aggregates.

#### Scenario: Availability accounting record is persisted
- **WHEN** the system stores an availability accounting record
- **THEN** its keys preserve tenant, service, monitor, objective or schedule version, and applicable UTC time-bucket identity
- **AND** reads do not scan unrelated tenants or services

#### Scenario: Finalizer resumes from storage
- **WHEN** bounded finalization stops or fails
- **THEN** a versioned tenant/shard cursor preserves finalized watermark, schedule-definition position, and opaque continuation state
- **AND** deterministic slot keys and aggregate application markers permit idempotent replay

### Requirement: Aggregate storage is bounded and outlives raw runs
The system SHALL retain hourly and daily availability aggregates for 400 days using TTL metadata independent of the shorter raw-run retention window.

#### Scenario: Aggregate record is written
- **WHEN** an hourly or daily availability aggregate is persisted
- **THEN** it includes numeric epoch-second TTL metadata for the aggregate retention horizon
- **AND** expiration remains compatible with table TTL on the `TTL` attribute

#### Scenario: Raw run expires before aggregate
- **WHEN** DynamoDB removes an expired raw `CheckRun`
- **THEN** the corresponding retained aggregate counts remain available for rolling reports

### Requirement: Aggregate updates are idempotent and reconcilable
The system SHALL prevent duplicate slot processing from incrementing aggregate counts more than once and SHALL retain deterministic monotonic slot and bucket versions, source application markers, closure state, and revision metadata for bounded reconciliation.

#### Scenario: Duplicate delivery updates an aggregate
- **WHEN** the same scheduled-slot accounting event is delivered repeatedly
- **THEN** conditional storage semantics apply its classification to the aggregate exactly once

#### Scenario: Authorized correction is applied
- **WHEN** a late canonical result or approved correction changes a slot classification within the supported correction policy
- **THEN** the affected aggregate revision is updated deterministically
- **AND** correction identity and prior classification remain auditable

#### Scenario: Hourly bucket closes
- **WHEN** every slot in a UTC hour has passed its 24-hour post-maturity correction horizon
- **THEN** the bucket is marked `CLOSED` with its final ordinary revision
- **AND** daily compaction consumes only closed hourly versions

#### Scenario: Closed bucket is corrected
- **WHEN** an authorized correction affects a closed hourly or compacted daily bucket
- **THEN** storage appends new monotonic bucket versions and retains superseded versions
- **AND** ordinary late-result processing cannot overwrite either version

### Requirement: Availability data follows recovery and stage lifecycle contracts
Immutable schedule/objective/interval/correction facts SHALL be durable without operational TTL; finalizer cursors and aggregates SHALL be classified and recoverable according to the landed recovery inventory; and persistent-stage tables SHALL use the landed stage-resource protections.

#### Scenario: Availability item families are added to recovery inventory
- **WHEN** availability storage is introduced
- **THEN** the inventory names each family as durable, reconstructable, or transient, its authoritative source, retention, rebuild behavior, and restore validation
- **AND** ephemeral stages are not represented as recoverable objective-history installations
