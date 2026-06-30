## ADDED Requirements

### Requirement: Service ID is generated server-side as ULID
System SHALL generate service IDs server-side using ULID format with `SVC_` prefix.

#### Scenario: Service ID format
- **WHEN** service is created
- **THEN** system generates a service ID matching pattern `SVC_<ULID>`
- **AND** ULID portion is lexicographically sortable
- **AND** ID is unique across all services in the tenant

#### Scenario: Service ID returned in response
- **WHEN** service is created successfully
- **THEN** response body includes the generated `serviceId`
- **AND** `Location` header contains URL with generated service ID

#### Scenario: Service ID is stable
- **WHEN** service is created with generated ID
- **THEN** that ID SHALL NOT change on subsequent operations
- **AND** ID is used as the primary key for all service operations

### Requirement: Client does not provide service ID
System SHALL NOT accept client-provided service ID on create.

#### Scenario: Create request without serviceId
- **WHEN** client submits create request without `serviceId` field
- **THEN** system generates service ID server-side

#### Scenario: Create request with serviceId field
- **WHEN** client submits create request including `serviceId` field
- **THEN** system ignores the provided value
- **AND** generates server-side ID instead
