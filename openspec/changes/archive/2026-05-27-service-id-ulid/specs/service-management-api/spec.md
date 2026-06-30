## MODIFIED Requirements

### Requirement: System creates services through HTTP API

#### Scenario: Client creates service with server-generated ID
- **WHEN** client submits valid create request for service with `name` and optional metadata (no `serviceId`)
- **THEN** system generates a unique `serviceId` with `SVC_` prefix and ULID suffix
- **AND** system persists service resource with generated `serviceId`
- **AND** system returns `201 Created` with `serviceId` in response body

#### Scenario: Client submits duplicate service name
- **WHEN** client submits create request with same `name` as existing service
- **THEN** system persists new service (no uniqueness constraint on name)

### Requirement: System lists and reads services through HTTP API

#### Scenario: Client fetches service by generated ID
- **WHEN** client requests existing service using its generated `serviceId`
- **THEN** system returns persisted service resource

### Requirement: System treats service ID as stable identifier

#### Scenario: Service ID is immutable
- **WHEN** service exists with persisted `serviceId`
- **THEN** all future update operations preserve that `serviceId`
- **AND** `serviceId` cannot be changed via update request

## REMOVED Requirements

### Requirement: System accepts client-provided serviceId

**Reason**: Service ID is now server-generated ULID. Clients no longer provide `serviceId` on create.

**Migration**: Remove `serviceId` field from create request. Use returned `serviceId` from response.
