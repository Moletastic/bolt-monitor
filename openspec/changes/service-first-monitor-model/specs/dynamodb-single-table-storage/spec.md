## MODIFIED Requirements

### Requirement: System uses one primary DynamoDB application table
System SHALL define one primary DynamoDB table for core monitoring application data instead of separate tables per entity.

#### Scenario: Core monitoring entities are persisted
- **WHEN** system persists monitoring application records
- **THEN** service, nested monitor, status, run, incident, and audit entities are modeled as typed items in one primary table

### Requirement: System defines canonical key patterns for core item families
System SHALL define canonical partition and sort key conventions for service-first item families used by monitoring workflows.

#### Scenario: Core entity type is mapped to storage
- **WHEN** system maps a service, nested monitor, status, run, incident, or audit record to DynamoDB
- **THEN** it uses documented PK and SK conventions for that item family, including service summary items, mutable service and monitor status items, and append-only `CheckRun` items

### Requirement: System supports tenant-aware query access patterns
System SHALL define item shapes and indexes that preserve tenant ownership and nested service-monitor ownership boundaries in storage queries.

#### Scenario: Tenant-scoped data is queried
- **WHEN** application reads monitoring data for one tenant
- **THEN** storage design supports tenant-scoped service listings, service detail reads, and nested monitor runtime reads without scanning unrelated tenant records

### Requirement: System defines initial operational GSIs
System SHALL define initial GSIs only for immediate operational reads needed by dashboard and incident workflows.

#### Scenario: Dashboard or incident view is planned
- **WHEN** system needs service rollup or open incident reads
- **THEN** storage design includes documented GSI access patterns only for those immediate operational reads

### Requirement: System defines retention strategy for high-volume run data
System SHALL define a retention strategy for raw check-run history items.

#### Scenario: Check-run history is persisted over time
- **WHEN** system stores high-volume run results under monitor-scoped runtime partitions
- **THEN** storage design includes explicit TTL or retention expectations for raw run items
