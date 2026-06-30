## Purpose

Define the service deletion API behavior for permanently removing eligible service resources while preserving required audit evidence.

## Requirements

### Requirement: System deletes services through HTTP API
System SHALL allow clients to permanently delete eligible service resources through HTTP API.

#### Scenario: Client deletes archived service
- **WHEN** client sends `DELETE /api/v1/services/{serviceId}` for an existing archived service in the current tenant
- **THEN** system removes the service from normal service list and read APIs
- **AND** system removes child monitor configuration from normal monitor list and read APIs
- **AND** system returns `204 No Content`

#### Scenario: Client deletes draft service
- **WHEN** client sends `DELETE /api/v1/services/{serviceId}` for an existing draft service in the current tenant
- **THEN** system removes the service from normal service list and read APIs
- **AND** system removes child monitor configuration from normal monitor list and read APIs
- **AND** system returns `204 No Content`

#### Scenario: Client deletes active service
- **WHEN** client sends `DELETE /api/v1/services/{serviceId}` for an existing active service in the current tenant
- **THEN** system rejects the request with conflict
- **AND** system preserves the service and all child monitor configuration

#### Scenario: Client deletes missing service
- **WHEN** client sends `DELETE /api/v1/services/{serviceId}` for a service that does not exist in the current tenant
- **THEN** system returns not found

### Requirement: System records service deletion audit events
System SHALL retain audit evidence for successful service deletion.

#### Scenario: Service deletion succeeds
- **WHEN** system successfully deletes a service
- **THEN** system records an audit event identifying the deleted service and deletion action
- **AND** the audit event is not removed as part of the same service configuration deletion

### Requirement: Deleted services are absent from active service reads
System SHALL treat deleted services as absent from active management APIs.

#### Scenario: Client reads deleted service
- **WHEN** client fetches a service after it has been deleted
- **THEN** system returns not found

#### Scenario: Client lists services after deletion
- **WHEN** client lists services after a service has been deleted
- **THEN** the deleted service is not included in the returned service collection
