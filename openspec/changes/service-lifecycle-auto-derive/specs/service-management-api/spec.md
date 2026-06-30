## MODIFIED Requirements

### Requirement: System creates services through HTTP API

#### Scenario: Client creates service
- **WHEN** client submits valid create request for service with `serviceId`, `name`, and optional metadata
- **THEN** system persists service resource
- **AND** service is assigned lifecycle state `draft` initially
- **AND** service is allowed to exist with zero monitors

#### Scenario: Client submits duplicate service identity
- **WHEN** client submits create request with `serviceId` that already exists within current tenant
- **THEN** system rejects request without persisting service

#### Scenario: Client submits invalid service slug
- **WHEN** client submits create request with `serviceId` that is not valid slug
- **THEN** system rejects request without persisting service

### Requirement: System lists and reads services through HTTP API

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

#### Scenario: Client updates service display fields
- **WHEN** client submits valid update request for existing service
- **THEN** system persists updated mutable fields such as `name`, `description`, and optional metadata
- **AND** system preserves original `serviceId`

#### Scenario: Client attempts to change stable service identity
- **WHEN** client submits update request that changes `serviceId`
- **THEN** system rejects request without changing persisted service identity

#### Scenario: Client attempts to set lifecycle state via update
- **WHEN** client submits update request that includes `lifecycleState`
- **THEN** system rejects request with error indicating lifecycle is read-only
- **AND** lifecycle state is not modified

### Requirement: System validates service technology keys against supported catalog

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

#### Scenario: Service slug is immutable
- **WHEN** service exists with persisted `serviceId`
- **THEN** all future update operations preserve that `serviceId`

## REMOVED Requirements

### Requirement: System validates service lifecycle transitions

**Reason**: Lifecycle transitions are now automatic based on `enabledCount`. Clients no longer set lifecycle state directly. Explicit archive/reactivate actions replace lifecycle transitions.

**Migration**: Use `POST /api/v1/services/{serviceId}/archive` to archive and `POST /api/v1/services/{serviceId}/reactivate` to reactivate.
