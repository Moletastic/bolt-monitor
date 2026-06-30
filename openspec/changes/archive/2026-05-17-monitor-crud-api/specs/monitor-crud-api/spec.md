## ADDED Requirements

### Requirement: System creates monitors through HTTP API
System SHALL allow clients to create monitor configurations through HTTP API.

#### Scenario: Client creates monitor
- **WHEN** client submits valid create request to monitor API
- **THEN** system validates monitor payload, persists monitor records, and returns created monitor resource

### Requirement: System lists and reads monitors through HTTP API
System SHALL allow clients to list monitors and fetch a single monitor through HTTP API.

#### Scenario: Client lists monitors
- **WHEN** client requests monitor collection
- **THEN** system returns monitor resources scoped to current tenant/workspace context

#### Scenario: Client fetches monitor by ID
- **WHEN** client requests existing monitor by ID
- **THEN** system returns persisted monitor resource for that monitor

### Requirement: System updates monitor configuration through HTTP API
System SHALL allow clients to update mutable monitor configuration fields through HTTP API.

#### Scenario: Client updates monitor
- **WHEN** client submits valid update request for existing monitor
- **THEN** system validates changed fields and persists updated monitor configuration

### Requirement: System enables and disables monitor lifecycle through HTTP API
System SHALL allow clients to explicitly enable and disable monitors through HTTP API.

#### Scenario: Client disables monitor
- **WHEN** client calls disable operation for existing monitor
- **THEN** system persists disabled lifecycle state for that monitor

#### Scenario: Client enables monitor
- **WHEN** client calls enable operation for existing monitor
- **THEN** system persists enabled lifecycle state for that monitor

### Requirement: System validates monitor CRUD payloads against shared contracts
System SHALL validate monitor CRUD payloads against canonical monitor and probe-location contracts.

#### Scenario: Client submits invalid probe-location selection
- **WHEN** client submits monitor payload with unknown or disabled probe location
- **THEN** system rejects request without persisting monitor changes

### Requirement: System writes audit records for monitor mutations
System SHALL write audit records for monitor create, update, enable, and disable operations.

#### Scenario: Monitor configuration changes
- **WHEN** client successfully changes monitor configuration or lifecycle state
- **THEN** system persists corresponding audit event records for that mutation
