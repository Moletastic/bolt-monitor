## ADDED Requirements

### Requirement: System defines canonical probe-location catalog
System SHALL define a canonical catalog of probe locations that represent valid execution locations for service checks.

#### Scenario: Probe location is represented in system
- **WHEN** system stores or exposes a probe location
- **THEN** it includes stable identifier and enough metadata for human selection and execution routing

### Requirement: System controls available probe locations
System SHALL control the set of valid probe locations instead of accepting arbitrary user-defined location strings.

#### Scenario: User selects location for monitor
- **WHEN** user configures monitor execution locations
- **THEN** selected values must come from the system-defined probe-location catalog

### Requirement: System distinguishes enabled probe locations
System SHALL track whether a probe location is available for monitor selection and execution.

#### Scenario: Probe location is disabled
- **WHEN** probe location is not available for use
- **THEN** system can exclude it from monitor selection and future execution routing

### Requirement: System uses vendor-neutral probe-location semantics
System SHALL treat probe locations as vendor-neutral execution locations rather than cloud-provider-specific regions.

#### Scenario: Multiple protocols share execution catalog
- **WHEN** HTTP, TCP, or gRPC monitors reference execution locations
- **THEN** they use same probe-location catalog semantics without requiring cloud-specific naming
