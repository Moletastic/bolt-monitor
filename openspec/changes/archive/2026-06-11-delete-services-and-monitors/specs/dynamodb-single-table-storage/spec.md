## ADDED Requirements

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
