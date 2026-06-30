## MODIFIED Requirements

### Requirement: System defines required monitor configuration fields
System SHALL require monitor configuration to include enough information to execute a check and manage its lifecycle.

#### Scenario: Required fields are validated
- **WHEN** system validates a monitor configuration
- **THEN** it requires fields for identity, ownership, human-readable name, monitor type, target, cadence, selected probe locations, and enabled state

## ADDED Requirements

### Requirement: System validates monitor probe-location selection against catalog
System SHALL validate monitor probe-location selections against the system-defined probe-location catalog.

#### Scenario: Monitor references probe locations
- **WHEN** user or API submits monitor configuration with `probeLocations`
- **THEN** every selected probe location must correspond to valid system-defined probe-location identifier
