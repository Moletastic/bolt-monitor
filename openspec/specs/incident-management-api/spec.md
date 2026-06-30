## Requirements

### Requirement: System exposes incident collection through HTTP API
System SHALL allow operators to read incident records through HTTP API.

#### Scenario: Operator requests incident collection
- **WHEN** operator calls `GET /api/v1/incidents`
- **THEN** system returns incident resources suitable for operational overview use

### Requirement: System exposes incident detail and monitor-scoped incident reads
System SHALL allow operators to inspect one incident directly and inspect incidents for one monitor.

#### Scenario: Operator requests incident by ID
- **WHEN** operator calls `GET /api/v1/incidents/{id}` for existing incident
- **THEN** system returns incident resource for that incident

#### Scenario: Operator requests incidents for one monitor
- **WHEN** operator calls `GET /api/v1/monitors/{id}/incidents` for existing monitor
- **THEN** system returns incidents associated with that monitor

### Requirement: System exposes incident acknowledgement and resolution as commands
System SHALL allow operators to acknowledge or resolve incidents through dedicated action endpoints.

#### Scenario: Operator acknowledges incident
- **WHEN** operator calls `POST /api/v1/incidents/{id}/ack` for actionable incident
- **THEN** system records acknowledgement state for that incident

#### Scenario: Operator resolves incident
- **WHEN** operator calls `POST /api/v1/incidents/{id}/resolve` for actionable incident
- **THEN** system records resolution state for that incident

### Requirement: System keeps incident lifecycle ownership in business rules
System SHALL keep incident creation and default closure under system business process rather than exposing generic incident create or delete CRUD.

#### Scenario: Client inspects API shape
- **WHEN** client needs incident visibility and operator actions
- **THEN** system provides read routes and action routes for incidents
- **AND** does not require or allow clients to create incidents directly through generic CRUD
