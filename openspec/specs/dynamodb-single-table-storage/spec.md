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
System SHALL define and enforce a retention strategy for raw check-run history items using native DynamoDB TTL.

#### Scenario: Check-run history is persisted over time
- **WHEN** system stores high-volume run results
- **THEN** storage design includes explicit TTL retention expectations for raw run items
- **AND** the primary DynamoDB table is configured so expired raw run items are eligible for automatic deletion

### Requirement: System enables native TTL for eligible DynamoDB items
The primary DynamoDB application table SHALL enable DynamoDB Time to Live on the `TTL` attribute so item families that write numeric epoch-second TTL values can expire without custom cleanup jobs.

#### Scenario: Table is provisioned
- **WHEN** infrastructure provisions the primary application DynamoDB table
- **THEN** the table has DynamoDB Time to Live enabled on the `TTL` attribute
- **AND** item families that do not write `TTL` remain persistent

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

### Requirement: System stores tenant-scoped search index records

The DynamoDB single-table design SHALL store sparse search index records that support low-I/O global resource search for a tenant.

#### Scenario: Search index record is written
- **WHEN** a searchable service, monitor, escalation policy, or notification channel is created or updated
- **THEN** system stores compact search index records under the tenant partition
- **AND** each search index sort key begins with `SEARCH#` followed by a normalized searchable prefix
- **AND** each search index record includes the resource type, resource stable identifiers, safe display label, safe display description, navigation href, icon discriminator, and match metadata

#### Scenario: Search query reads index records
- **WHEN** system searches for a normalized query
- **THEN** system queries DynamoDB with `PK = TENANT#<tenant>` and a `begins_with(SK, SEARCH#<normalized-query>)` key condition
- **AND** system does not scan the table or list all services, monitors, policies, or channels to satisfy the query
- **AND** system uses a bounded read limit before result de-duplication and ranking

#### Scenario: Search index fields are selected
- **WHEN** system builds search index entries for services
- **THEN** searchable service fields include name, normalized service slug, description, service category, lifecycle state, and rollup status
- **WHEN** system builds search index entries for monitors
- **THEN** searchable monitor fields include name, normalized monitor slug, parent service name, HTTP target safe display, monitor type, and enabled state
- **WHEN** system builds search index entries for escalation policies
- **THEN** searchable policy fields include name, normalized policy slug, description, route structure summary, and referenced channel IDs as safe references
- **WHEN** system builds search index entries for notification channels
- **THEN** searchable channel fields include name, normalized channel slug, channel type, and safe target display

#### Scenario: Sensitive fields are excluded from search storage
- **WHEN** system builds search index entries
- **THEN** system excludes monitor headers, expected body text, notification channel config JSON, inline channel config, tenant identifiers, and secret-bearing URL query strings from search index text and search API display text

#### Scenario: Search index records are bounded
- **WHEN** system derives searchable prefixes from a resource
- **THEN** system limits indexed tokens, prefix lengths, and total records per resource to protect write cost and transaction size

#### Scenario: Search index entries are removed
- **WHEN** a searchable resource is deleted
- **THEN** system deletes the search index records associated with that resource

### Requirement: System indexes audit events by resource and occurrence time
System SHALL maintain a sparse secondary index for audit events with a partition key identifying tenant, service, and monitor resource scope and a sort key ordering audit occurrence time with stable audit identity.

#### Scenario: System records monitor audit event
- **WHEN** system writes an audit event associated with a monitor
- **THEN** event includes resource index keys for that tenant, service, and monitor
- **AND** index sort key orders event by timestamp and audit identity

#### Scenario: System records service-level audit event
- **WHEN** system writes an audit event associated with a service and no monitor
- **THEN** event includes resource index keys for that tenant and service with empty monitor scope
- **AND** service-level audit lookup does not collide with any monitor-scoped audit lookup

#### Scenario: Resource audit history is queried
- **WHEN** application reads audit history for a monitor or service resource
- **THEN** it queries the audit resource secondary index with an exact resource partition key
- **AND** it applies a bounded limit before returning records
