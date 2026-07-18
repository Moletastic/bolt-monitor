## MODIFIED Requirements

### Requirement: System supports HTTP monitor configuration in v1
System SHALL support HTTP monitor definitions as the first monitor type in v1 and SHALL restrict monitor targets to absolute `http` or `https` URLs accepted by the shared default public outbound policy.

#### Scenario: HTTP monitor is configured
- **WHEN** user or API defines a monitor of type `http` with a safe public HTTP or HTTPS target
- **THEN** system accepts HTTP-specific configuration including target URL, request method, timeout, and expected response settings

#### Scenario: HTTP monitor uses an unsupported scheme
- **WHEN** user or API defines a monitor target with any scheme other than `http` or `https`
- **THEN** system rejects the configuration with `VALIDATION_FAILED` and `reason.details.field == "http.target"`

#### Scenario: HTTP monitor uses a blocked destination
- **WHEN** user or API defines a monitor target that is a blocked literal address or currently resolves to any address blocked by the default public policy
- **THEN** system rejects the configuration with `VALIDATION_FAILED` and `reason.details.field == "http.target"`

### Requirement: Validators return TypedError with field paths
`Service.Validate`, `Monitor.Validate`, and `HTTPConfiguration.Validate` SHALL each return `error`, but every error returned SHALL be a `*shared/errors.TypedError{Code: CodeValidationFailed}` with `Details["field"]` set to the offending field's dotted path and an optional sanitized `Details["reason"]` for human context. HTTP target validation SHALL use the shared URL policy for syntax, scheme, literal-address, and network-aware API validation; execution SHALL repeat network-aware validation for persisted configurations.

#### Scenario: Service.Validate rejects empty tenantId
- **WHEN** `Service.Validate` runs against a service with empty `tenantId`
- **THEN** the returned error is `*TypedError{Code: CodeValidationFailed}` with `Details["field"] == "tenantId"`

#### Scenario: HTTPConfiguration.Validate rejects malformed target URL
- **WHEN** `HTTPConfiguration.Validate` runs against a malformed or non-absolute target URL
- **THEN** the returned error is `*TypedError{Code: CodeValidationFailed}` with `Details["field"] == "http.target"`

#### Scenario: HTTPConfiguration.Validate rejects excessive timeout
- **WHEN** `HTTPConfiguration.Validate` runs with `timeoutMs` greater than 30000
- **THEN** the returned error is `*TypedError{Code: CodeValidationFailed}` with `Details["field"] == "http.timeoutMs"`

#### Scenario: Target validation does not expose URL secrets
- **WHEN** target validation rejects a URL containing user information or sensitive query values
- **THEN** the typed validation details do not contain those values
