## MODIFIED Requirements

### Requirement: System exposes incident acknowledgement and resolution as commands
System SHALL allow operators to acknowledge or resolve incidents through dedicated action endpoints.

#### Scenario: Operator acknowledges incident
- **WHEN** operator calls `POST /api/v1/incidents/{id}/ack` for actionable incident
- **THEN** system records acknowledgement state for that incident

#### Scenario: Operator resolves incident
- **WHEN** operator calls `POST /api/v1/incidents/{id}/resolve` for actionable incident
- **THEN** system records resolution state for that incident

### Requirement: System manages incident lifecycle through escalation
System SHALL open and resolve incidents based on monitor state transitions, with notification delivery handled by the escalation system.

#### Scenario: Incident opens on DOWN transition
- **WHEN** monitor transitions from DEGRADED to DOWN state
- **THEN** system creates a new incident record with status OPEN
- **AND** system emits incident.down event to trigger escalation
- **AND** escalation system takes over notification delivery based on service policy

#### Scenario: Incident resolves on UP transition
- **WHEN** monitor transitions from RECOVERING to UP state
- **THEN** system resolves the open incident
- **AND** system emits incident.up event to suppress escalation

## ADDED Requirements

### Requirement: System exposes escalation.exhausted incident type
System SHALL distinguish escalation.exhausted incidents from regular incidents in the API.

#### Scenario: Operator lists incidents and escalation.exhausted is present
- **WHEN** operator calls `GET /api/v1/incidents`
- **THEN** escalation.exhausted incidents are included in the response with their type indicated

#### Scenario: Operator retrieves escalation.exhausted incident detail
- **WHEN** operator calls `GET /api/v1/incidents/{id}` for an escalation.exhausted incident
- **THEN** the incident resource includes the original incident ID reference
- **AND** the incident type field indicates escalation.exhausted

### Requirement: System creates escalation.exhausted incident when escalation path completes
System SHALL automatically create an escalation.exhausted incident when all escalation steps have fired and the original incident remains open.

#### Scenario: Escalation exhausts all steps
- **WHEN** all escalation steps fire for an open incident
- **AND** the original incident remains in OPEN status
- **THEN** system creates a new incident with type escalation.exhausted
- **AND** the escalation.exhausted incident references the original incident
- **AND** the original incident remains OPEN
