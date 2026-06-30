## ADDED Requirements

### Requirement: System creates services through HTTP API
System SHALL allow clients to create service resources through HTTP API.

#### Scenario: Client creates draft service without monitors
- **WHEN** client submits valid create request for service with `serviceId`, `name`, and optional metadata
- **THEN** system persists service resource
- **AND** service is allowed to exist with zero monitors
- **AND** service rollup status is `draft`

#### Scenario: Client submits duplicate service identity
- **WHEN** client submits create request with `serviceId` that already exists within current tenant
- **THEN** system rejects request without persisting service

#### Scenario: Client submits invalid service slug
- **WHEN** client submits create request with `serviceId` that is not valid slug
- **THEN** system rejects request without persisting service

### Requirement: System lists and reads services through HTTP API
System SHALL allow clients to list service summaries and fetch one service with child monitor summaries through HTTP API.

#### Scenario: Client lists services
- **WHEN** client requests service collection
- **THEN** system returns service summaries scoped to current tenant
- **AND** each summary includes service identity, lifecycle state, derived rollup status, and monitor counts

#### Scenario: Client fetches service with no monitors
- **WHEN** client requests existing service that has zero monitors
- **THEN** system returns persisted service resource
- **AND** response includes empty `monitors` collection

#### Scenario: Client fetches missing service
- **WHEN** client requests service that does not exist within current tenant
- **THEN** system returns not found

### Requirement: System updates mutable service fields through HTTP API
System SHALL allow clients to update mutable service fields without changing stable service identity.

#### Scenario: Client updates service display fields
- **WHEN** client submits valid update request for existing service
- **THEN** system persists updated mutable fields such as `name`, `description`, and optional metadata
- **AND** system preserves original `serviceId`

#### Scenario: Client attempts to change stable service identity
- **WHEN** client submits update request that changes `serviceId`
- **THEN** system rejects request without changing persisted service identity

### Requirement: System validates service lifecycle transitions
System SHALL enforce lifecycle rules for service activation and archival.

#### Scenario: Client activates service with at least one monitor
- **WHEN** client updates existing draft or archived service to lifecycle state `active`
- **AND** service contains at least one monitor
- **THEN** system persists lifecycle state `active`

#### Scenario: Client activates service with zero monitors
- **WHEN** client updates existing service to lifecycle state `active`
- **AND** service contains zero monitors
- **THEN** system rejects request without changing persisted lifecycle state

#### Scenario: Client archives active service
- **WHEN** client updates existing active service to lifecycle state `archived`
- **THEN** system persists archived lifecycle state

### Requirement: System validates service technology keys against supported catalog
System SHALL accept only supported semantic `technologyKey` values for service metadata.

#### Scenario: Client submits supported technology key
- **WHEN** client submits valid service create or update request with supported `technologyKey`
- **THEN** system persists that technology key on service resource

#### Scenario: Client omits technology key
- **WHEN** client submits valid service create or update request without `technologyKey`
- **THEN** system persists service without primary technology metadata

#### Scenario: Client submits unsupported technology key
- **WHEN** client submits service create or update request with unsupported `technologyKey`
- **THEN** system rejects request without persisting service changes

### Requirement: System treats service slug as stable identifier
System SHALL require `serviceId` to remain stable after creation.

#### Scenario: Service slug is immutable
- **WHEN** service exists with persisted `serviceId`
- **THEN** all future update operations preserve that `serviceId`
