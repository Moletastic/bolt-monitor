## MODIFIED Requirements

### Requirement: System defines stable monitor identity and ownership
System SHALL define each monitor with stable child identity, tenant ownership metadata, and parent service ownership.

#### Scenario: New monitor is represented in system
- **WHEN** system creates or stores a monitor configuration
- **THEN** configuration includes `monitorId`, `serviceId`, and `tenantId` fields that identify the monitor, its parent service, and its ownership boundary
- **AND** `monitorId` uniqueness is enforced within its parent service rather than globally across the tenant

### Requirement: System supports HTTP monitor configuration in v1
System SHALL support HTTP monitor definitions as first monitor type in v1.

#### Scenario: HTTP monitor is configured
- **WHEN** user or API defines a monitor of type `http`
- **THEN** system accepts HTTP-specific configuration including target URL, request method, timeout, and expected response settings

### Requirement: System defines required monitor configuration fields
System SHALL require monitor configuration to include enough information to execute a check, relate the monitor to a service, and manage its lifecycle.

#### Scenario: Required fields are validated
- **WHEN** system validates a monitor configuration
- **THEN** it requires fields for child identity, parent service identity, tenant ownership, human-readable name, monitor type, target, cadence, selected probe locations, and enabled state

### Requirement: System validates monitor probe-location selection against catalog
System SHALL validate monitor probe-location selections against the system-defined probe-location catalog.

#### Scenario: Monitor references probe locations
- **WHEN** user or API submits monitor configuration with `probeLocations`
- **THEN** every selected probe location must correspond to valid system-defined probe-location identifier

### Requirement: System distinguishes monitor lifecycle enablement
System SHALL track whether a monitor is enabled or disabled for execution.

#### Scenario: Disabled monitor is stored
- **WHEN** monitor configuration has disabled lifecycle state
- **THEN** system preserves that state so downstream scheduling and execution systems skip running it

### Requirement: System preserves canonical monitor contract across subsystems
System SHALL treat nested monitor configuration as canonical contract reused by API, persistence, scheduling, and probing systems.

#### Scenario: Downstream subsystem consumes monitor definition
- **WHEN** persistence or execution subsystem reads monitor configuration
- **THEN** it uses same canonical field contract rather than subsystem-specific configuration shape
