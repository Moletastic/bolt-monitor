## ADDED Requirements

### Requirement: Audit history reads are resource-scoped and bounded
System SHALL read monitor and service audit history through a resource-scoped storage access pattern, newest first, limited to 20 matching events per request.

#### Scenario: Operator requests monitor audit history in tenant with unrelated events
- **WHEN** operator requests audit history for one monitor and tenant contains audit events for other services or monitors
- **THEN** system reads only audit events indexed for requested monitor resource
- **AND** system does not retrieve unrelated tenant audit events for application-side filtering

#### Scenario: Operator requests service audit history
- **WHEN** operator requests audit history for one service
- **THEN** system reads only service-level audit events indexed for requested service resource
- **AND** system returns at most 20 events and cursor continuation metadata when more exist

### Requirement: Audit history preserves existing events during index migration
System SHALL preserve read visibility of audit events created before resource-index writes are enabled.

#### Scenario: Historical audit event is backfilled
- **WHEN** an existing audit event lacks resource index keys during migration
- **THEN** migration writes its resource index keys idempotently
- **AND** monitor or service audit readers return that event after index reader cutover
