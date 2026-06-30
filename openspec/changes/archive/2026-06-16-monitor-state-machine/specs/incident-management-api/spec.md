## MODIFIED Requirements

### Requirement: System exposes incident acknowledgement and resolution as commands
System SHALL allow operators to acknowledge or resolve incidents through dedicated action endpoints.

#### Scenario: Operator acknowledges incident
- **WHEN** operator calls `POST /api/v1/incidents/{id}/ack` for actionable incident
- **THEN** system records acknowledgement state for that incident

#### Scenario: Operator resolves incident
- **WHEN** operator calls `POST /api/v1/incidents/{id}/resolve` for actionable incident
- **THEN** system records resolution state for that incident

## ADDED Requirements

### Requirement: System manages incident lifecycle through state machine
System SHALL open incidents only when failure threshold is met (DOWN state) and resolve incidents only when recovery threshold is met (UP state transition from RECOVERING).

#### Scenario: Incident opens on threshold breach
- **WHEN** monitor transitions from DEGRADED to DOWN state
- **THEN** system creates a new incident record with status OPEN
- **AND** emits incident.opened notification event

#### Scenario: Incident resolves on recovery threshold met
- **WHEN** monitor transitions from RECOVERING to UP state
- **THEN** system resolves the open incident
- **AND** emits incident.resolved notification event

#### Scenario: Incident remains open through RECOVERING state
- **WHEN** monitor is in RECOVERING state
- **THEN** open incident remains in OPEN status
- **AND** no notification event is emitted during accumulation
