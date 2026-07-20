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
## ADDED Requirements

### Requirement: System consumes the retry-safe canonical transition outbox
The DynamoDB single-table design SHALL use the single canonical tenant-scoped transition event/outbox item created atomically with recurring result and incident transition state by `make-check-execution-retry-safe`. This change SHALL add no competing transition item or producer protocol. The item SHALL be keyed by stable `eventId`, carry immutable causal identity and pending/acknowledged dispatch metadata, and remain the authority for Stream dispatch acknowledgement and repair.

#### Scenario: Incident changes notification-relevant state
- **WHEN** retry-safe execution commits an incident down or recovery transition
- **THEN** its transaction stores exactly one canonical outbox item containing safe immutable routing context and the stable incident activity transition identity
- **AND** notification assurance consumes rather than recreates it

#### Scenario: Queue acceptance is acknowledged
- **WHEN** the dispatcher confirms SQS acceptance for a canonical identity
- **THEN** it conditionally changes that same item from pending to acknowledged
- **AND** an ambiguous acknowledgement leaves the item pending

#### Scenario: Duplicate transition transaction is retried
- **WHEN** incident transition persistence is retried with the same transition identity
- **THEN** the table contains at most one corresponding outbox item

### Requirement: Pending dispatch has a durable bounded access path
Canonical pending transition and replay dispatch records SHALL remain queryable through a sparse tenant/time-bucketed access path until acknowledged. Automatic reconciliation SHALL bound buckets, pages, and items per invocation; manual repair SHALL use a point lookup by canonical identity. Pending records SHALL NOT expire merely because Stream retries exhaust, and no repair path SHALL scan unrelated table records.

#### Scenario: Stream retries exhaust
- **WHEN** a canonical dispatch insert exhausts DynamoDB Stream retries
- **THEN** the item remains pending in its sparse bucket
- **AND** its canonical identity supports direct manual repair

#### Scenario: Automatic reconciliation runs
- **WHEN** the reconciler searches for unacknowledged dispatch work
- **THEN** it reads only configured recent tenant/time buckets and bounded pages
- **AND** does not scan the primary table

#### Scenario: Dispatch is acknowledged
- **WHEN** an item transitions to acknowledged
- **THEN** it leaves the pending access path and receives the configured bounded acknowledged-record TTL

### Requirement: System stores incident-scoped notification delivery records
The DynamoDB single-table design SHALL define a canonical item pattern that groups delivery records under their incident for bounded API queries and enforces deterministic identity for transition, policy step, and channel. Delivery records SHALL persist safe operational fields only.

#### Scenario: Delivery record is written
- **WHEN** escalation work resolves a policy step and channel
- **THEN** the system stores the delivery under the incident partition with a sort key containing stable delivery ordering and identity
- **AND** a conditional write prevents duplicate identity creation

#### Scenario: Incident deliveries are queried
- **WHEN** the API lists delivery outcomes for an incident
- **THEN** it performs a bounded query of that incident partition
- **AND** does not scan unrelated incidents or tenants

#### Scenario: Delivery data is retained
- **WHEN** an incident and its delivery records become historical
- **THEN** delivery outcomes remain available with incident history under the repository's operational retention policy
- **AND** secret-bearing provider data is never stored in those records

### Requirement: System stores replay commands and idempotency records boundedly
Delivery replay SHALL use the canonical dispatch-record schema with `sourceKind=delivery_replay` and the same Stream dispatcher. The replay transaction SHALL store a tenant/incident/delivery/operation/key-scoped idempotency item containing a canonical request fingerprint and result identity. Idempotency records SHALL use a named bounded TTL longer than the maximum replay dispatch/retry window and SHALL be conditionally unique during retention.

#### Scenario: First replay request commits
- **WHEN** an eligible replay with a new `Idempotency-Key` is accepted
- **THEN** one transaction updates the delivery, creates one replay dispatch record, and creates one idempotency record

#### Scenario: Same replay request is retried
- **WHEN** the same key and request fingerprint are presented during retention
- **THEN** the stored result identity is returned without another delivery update or dispatch record

#### Scenario: Replay payload conflicts
- **WHEN** the same key is presented with a different request fingerprint during retention
- **THEN** the conditional operation rejects it as an idempotency conflict
