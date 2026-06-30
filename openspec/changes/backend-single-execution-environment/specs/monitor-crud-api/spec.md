## MODIFIED Requirements

### Requirement: System validates monitor CRUD payloads against canonical monitor contracts
System SHALL validate monitor CRUD payloads against the canonical monitor contract.

#### Scenario: Client submits valid monitor payload
- **WHEN** client submits a monitor create or update payload with valid name, type, cadence, enabled state, thresholds, and HTTP configuration
- **THEN** system accepts the payload without requiring probe-location selection

#### Scenario: Client submits obsolete probe-location selection
- **WHEN** client submits monitor payload fields for probe-location or region selection
- **THEN** system does not treat those fields as part of the accepted monitor contract
