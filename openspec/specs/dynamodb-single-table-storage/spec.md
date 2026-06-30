## Requirements

### Requirement: System uses one primary DynamoDB application table
System SHALL define one primary DynamoDB table for core monitoring application data instead of separate tables per entity.

#### Scenario: Core monitoring entities are persisted
- **WHEN** system persists monitoring application records
- **THEN** monitor, status, run, incident, and audit entities are modeled as typed items in one primary table

### Requirement: System defines canonical key patterns for core item families
System SHALL define canonical partition and sort key conventions for core item families used by monitoring workflows.

#### Scenario: Core entity type is mapped to storage
- **WHEN** system maps a monitor, status, run, incident, or audit record to DynamoDB
- **THEN** it uses documented PK/SK conventions for that item family, including append-only `CheckRun` items and mutable `MonitorStatus` items

### Requirement: System supports tenant-aware query access patterns
System SHALL define item shapes and indexes that preserve tenant/workspace ownership boundaries in storage queries.

#### Scenario: Tenant-scoped data is queried
- **WHEN** application reads monitoring data for one tenant
- **THEN** storage design supports tenant-scoped access without scanning unrelated tenant records

### Requirement: System defines initial operational GSIs
System SHALL define initial GSIs only for immediate operational reads needed by dashboard and incident workflows.

#### Scenario: Dashboard or incident view is planned
- **WHEN** system needs monitor status or open incident reads
- **THEN** storage design includes documented GSI access patterns for those reads

### Requirement: System defines retention strategy for high-volume run data
System SHALL define a retention strategy for raw check-run history items.

#### Scenario: Check-run history is persisted over time
- **WHEN** system stores high-volume run results
- **THEN** storage design includes explicit TTL or retention expectations for raw run items

### Requirement: System removes deleted configuration from active storage reads
System SHALL remove deleted service and monitor configuration records from storage access patterns used by active management APIs.

#### Scenario: Service configuration is deleted
- **WHEN** system deletes a service
- **THEN** storage no longer returns the service metadata, tenant service reference, service status, child monitor metadata, child monitor references, or current child monitor status through active service and monitor read paths

#### Scenario: Monitor configuration is deleted
- **WHEN** system deletes a monitor
- **THEN** storage no longer returns the monitor metadata, service monitor reference, current monitor status, or monitor notification links through active monitor read paths

### Requirement: System preserves deletion audit records
System SHALL preserve audit records that document successful service and monitor deletion.

#### Scenario: Deletion audit is written
- **WHEN** system writes an audit record for successful service or monitor deletion
- **THEN** storage retains that audit record independently of deleted configuration records

### Requirement: System does not require history purge for deletion
System SHALL NOT require deletion of historical operational records when service or monitor configuration is deleted.

#### Scenario: Configuration is deleted with existing history
- **WHEN** system deletes service or monitor configuration that has historical check runs, incidents, or prior audit records
- **THEN** system may leave historical records in storage under existing retention rules
- **AND** normal active service and monitor management APIs do not expose the deleted configuration as active resources
