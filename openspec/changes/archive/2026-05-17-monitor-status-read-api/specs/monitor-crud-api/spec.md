## MODIFIED Requirements

### Requirement: System lists and reads monitors through HTTP API
System SHALL allow clients to list monitors and fetch a single monitor through HTTP API.

#### Scenario: Client lists monitors
- **WHEN** client requests monitor collection
- **THEN** system returns monitor resources scoped to current tenant/workspace context and may include current status summary

#### Scenario: Client fetches monitor by ID
- **WHEN** client requests existing monitor by ID
- **THEN** system returns persisted monitor resource for that monitor with operational status data available through the read surface
