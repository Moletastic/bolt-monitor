## ADDED Requirements

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
