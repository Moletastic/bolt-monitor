## ADDED Requirements

### Requirement: System creates escalation.exhausted incident type
System SHALL create a distinct incident type `escalation.exhausted` when an escalation path completes all steps while the original incident remains open.

#### Scenario: Escalation path exhausts with open incident
- **WHEN** all escalation steps have fired
- **AND** the original incident status is OPEN
- **THEN** system creates a new incident with type escalation.exhausted
- **AND** the new incident's summary indicates escalation exhaustion

### Requirement: Escalation.exhausted incident links to original incident
System SHALL maintain a reference from the escalation.exhausted incident back to the original incident.

#### Scenario: Operator views escalation.exhausted incident
- **WHEN** operator retrieves an escalation.exhausted incident
- **THEN** the incident record contains the original incident ID
- **AND** operator can retrieve the original incident via that reference

### Requirement: Escalation.exhausted incident does not trigger further escalation
System SHALL NOT initiate a new escalation path for an escalation.exhausted incident.

#### Scenario: Escalation.exhausted incident is created
- **WHEN** system creates an escalation.exhausted incident
- **THEN** no escalation state is created for this incident
- **AND** no notification escalation path is initiated

### Requirement: Escalation.exhausted incident requires manual resolution
System SHALL require escalation.exhausted incidents to be resolved manually through operator action.

#### Scenario: Operator resolves escalation.exhausted incident
- **WHEN** operator calls `POST /api/v1/incidents/{id}/resolve` for an escalation.exhausted incident
- **THEN** system records resolution state for that incident
- **AND** the original incident remains in its existing state
