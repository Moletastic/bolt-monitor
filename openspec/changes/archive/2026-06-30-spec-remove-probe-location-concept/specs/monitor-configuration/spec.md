## MODIFIED Requirements

### Requirement: System defines required monitor configuration fields
System SHALL require monitor configuration to include enough information to execute a check and manage its lifecycle.

#### Scenario: Required fields are validated
- **WHEN** system validates a monitor configuration
- **THEN** it requires fields for identity, ownership, human-readable name, monitor type, target, cadence, and enabled state
- **AND** it does not require an operator-selected probe location or region

## REMOVED Requirements

### Requirement: System validates monitor probe-location selection against catalog
System SHALL validate monitor probe-location selections against the system-defined probe-location catalog.

#### Scenario: Monitor references probe locations
- **WHEN** user or API submits monitor configuration with `probeLocations`
- **THEN** every selected probe location must correspond to valid system-defined probe-location identifier

### Requirement: Exported validation signatures are preserved
The exported signatures of `Service.Validate()`, `Monitor.Validate()`, `HTTPConfiguration.Validate()`, and `Monitor.ValidateWithCatalog(catalog)` SHALL remain unchanged.

#### Scenario: Signature compatibility
- **WHEN** existing callers compile against the new validators
- **THEN** compilation succeeds without code changes
