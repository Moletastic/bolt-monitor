## MODIFIED Requirements

### Requirement: Validators return TypedError with field paths

`Service.Validate`, `Monitor.Validate`, `HTTPConfiguration.Validate`, and `Monitor.ValidateWithCatalog` SHALL each return `error` as today, but every error returned SHALL be a `*shared/errors.TypedError{Code: CodeValidationFailed}` with `Details["field"]` set to the offending field's dotted path and an optional `Details["reason"]` for human context. The exported function signatures SHALL be preserved.

#### Scenario: Service.Validate rejects empty tenantId
- **WHEN** `Service.Validate` runs against a service with empty `tenantId`
- **THEN** the returned error is `*TypedError{Code: CodeValidationFailed}` with `Details["field"] == "tenantId"`

#### Scenario: Monitor.ValidateWithCatalog rejects unknown probe location
- **WHEN** `Monitor.ValidateWithCatalog` runs with a probe location ID not in the catalog
- **THEN** the returned error is `*TypedError{Code: CodeValidationFailed}` with `Details["field"] == "probeLocations"`

#### Scenario: HTTPConfiguration.Validate rejects malformed target URL
- **WHEN** `HTTPConfiguration.Validate` runs against a target URL that fails `url.ParseRequestURI`
- **THEN** the returned error is `*TypedError{Code: CodeValidationFailed}` with `Details["field"] == "http.target"`

### Requirement: Exported validation signatures are preserved

The exported signatures of `Service.Validate()`, `Monitor.Validate()`, `HTTPConfiguration.Validate()`, and `Monitor.ValidateWithCatalog(catalog)` SHALL remain unchanged.

#### Scenario: Signature compatibility
- **WHEN** existing callers compile against the new validators
- **THEN** compilation succeeds without code changes
