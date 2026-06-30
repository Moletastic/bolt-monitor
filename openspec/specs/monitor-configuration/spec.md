## Requirements

### Requirement: System defines stable monitor identity and ownership
System SHALL define each monitor with stable unique identity and tenant ownership metadata.

#### Scenario: New monitor is represented in system
- **WHEN** system creates or stores a monitor configuration
- **THEN** configuration includes `monitorId` and `tenantId` fields that identify monitor and its ownership boundary

### Requirement: System supports HTTP monitor configuration in v1
System SHALL support HTTP monitor definitions as first monitor type in v1.

#### Scenario: HTTP monitor is configured
- **WHEN** user or API defines a monitor of type `http`
- **THEN** system accepts HTTP-specific configuration including target URL, request method, timeout, and expected response settings

### Requirement: System defines required monitor configuration fields
System SHALL require monitor configuration to include enough information to execute a check and manage its lifecycle.

#### Scenario: Required fields are validated
- **WHEN** system validates a monitor configuration
- **THEN** it requires fields for identity, ownership, human-readable name, monitor type, target, cadence, and enabled state
- **AND** it does not require an operator-selected probe location or region

### Requirement: System distinguishes monitor lifecycle enablement
System SHALL track whether a monitor is enabled or disabled for execution.

#### Scenario: Disabled monitor is stored
- **WHEN** monitor configuration has disabled lifecycle state
- **THEN** system preserves that state so downstream scheduling and execution systems skip running it

### Requirement: System preserves canonical monitor contract across subsystems
System SHALL treat monitor configuration as canonical contract reused by API, persistence, scheduling, and probing systems.

#### Scenario: Downstream subsystem consumes monitor definition
- **WHEN** persistence or execution subsystem reads monitor configuration
- **THEN** it uses same canonical field contract rather than subsystem-specific configuration shape

### Requirement: Validators return TypedError with field paths

`Service.Validate`, `Monitor.Validate`, and `HTTPConfiguration.Validate` SHALL each return `error` as today, but every error returned SHALL be a `*shared/errors.TypedError{Code: CodeValidationFailed}` with `Details["field"]` set to the offending field's dotted path and an optional `Details["reason"]` for human context.

#### Scenario: Service.Validate rejects empty tenantId
- **WHEN** `Service.Validate` runs against a service with empty `tenantId`
- **THEN** the returned error is `*TypedError{Code: CodeValidationFailed}` with `Details["field"] == "tenantId"`

#### Scenario: HTTPConfiguration.Validate rejects malformed target URL
- **WHEN** `HTTPConfiguration.Validate` runs against a target URL that fails `url.ParseRequestURI`
- **THEN** the returned error is `*TypedError{Code: CodeValidationFailed}` with `Details["field"] == "http.target"`
