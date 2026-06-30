## ADDED Requirements

### Requirement: `Code` is a typed string with named constants

`Code` SHALL be a Go string type defined in `shared/errors`. Each code currently emitted by `services/monitor-api/handler.go` SHALL be exported as a `CodeXxx` constant. The wire representation of a `Code` SHALL be exactly the value of its constant, with no transformation.

The full set of constants: `NOT_FOUND`, `INVALID_JSON`, `VALIDATION_FAILED`, `IMMUTABLE_FIELD`, `INLINE_CHANNEL_CONFIG`, `SERVICE_NOT_FOUND`, `SERVICE_ALREADY_EXISTS`, `SERVICE_ACTIVE`, `SERVICE_NOT_ARCHIVED`, `SERVICE_HAS_NO_POLICY`, `MONITOR_NOT_FOUND`, `MONITOR_ALREADY_EXISTS`, `MONITOR_DISABLED`, `MONITOR_STATUS_NOT_FOUND`, `LAST_MONITOR`, `INCIDENT_NOT_FOUND`, `INCIDENT_NOT_ACTIONABLE`, `POLICY_NOT_FOUND`, `POLICY_REFERENCED`, `CHANNEL_NOT_FOUND`, `INTERNAL`.

#### Scenario: New code constant for a not-found outcome
- **WHEN** a handler returns a `*TypedError{Code: CodeServiceNotFound}`
- **THEN** the wire body contains `reason.code == "SERVICE_NOT_FOUND"` exactly

#### Scenario: Adding a new code
- **WHEN** a developer adds a new `CodeXxx` constant
- **THEN** the test in `code_test.go` enumerating all constants fails if the new constant is not in the registry

### Requirement: `StatusOf` is the single source of `code → httpStatus`

For every `CodeXxx` constant there SHALL be exactly one entry in the registry. `StatusOf(code Code) int` SHALL return the registered HTTP status. `StatusOf` SHALL panic if `code` is not registered.

#### Scenario: StatusOf returns the registered status
- **WHEN** the registry maps `CodeServiceNotFound` to `http.StatusNotFound`
- **THEN** `StatusOf(CodeServiceNotFound) == 404`

#### Scenario: StatusOf panics on unknown code
- **WHEN** `StatusOf` is called with a code constant that is missing from the registry
- **THEN** it panics

### Requirement: `TypedError` carries code, details, and optional cause

`TypedError` SHALL expose `Code Code`, `Details map[string]any`, `Cause error`, and an optional `Message string`. `Error()` SHALL return the code, suffixed with the cause's message when `Cause` is non-nil. `Unwrap()` SHALL return `Cause`.

`Details` SHALL be the only field of `TypedError` that flows to `response.reason.details`. `Message` SHALL NOT be serialized.

#### Scenario: TypedError.Error() includes cause
- **WHEN** a `TypedError` is constructed with `Wrap(CodeInternal, fmt.Errorf("boom"), nil)`
- **THEN** `err.Error()` contains both `INTERNAL` and `boom`

#### Scenario: TypedError is reachable via errors.As
- **WHEN** an error chain wraps a `TypedError` via `fmt.Errorf("wrap: %w", te)`
- **THEN** `errors.As(err, &te)` returns the inner `TypedError`

### Requirement: Validation errors carry `details.field`

When `TypedError.Code == VALIDATION_FAILED`, `Details` SHALL include a `"field"` key whose value is a dotted path identifying the offending field. A `"reason"` key MAY accompany `"field"` for human context.

#### Scenario: Leaf field validation surfaces a path
- **WHEN** `Monitor.Validate` rejects an empty `tenantId`
- **THEN** the emitted `TypedError` has `Details["field"] == "tenantId"`

#### Scenario: Nested field validation surfaces a dotted path
- **WHEN** `HTTPConfiguration.Validate` rejects a malformed target URL
- **THEN** the emitted `TypedError` has `Details["field"] == "http.target"`

#### Scenario: Indexed step validation surfaces a bracket path
- **WHEN** `validateEscalationPolicy` rejects an empty channel on step 3 of the business-hours path
- **THEN** the emitted `TypedError` has `Details["field"] == "businessHoursPath.steps[2].channelId"`

### Requirement: `Respond` handles every error uniformly

`Respond(err error) (int, response.Envelope[any])` SHALL: typed errors (found via `errors.As`) return `(StatusOf(code), response.Err[any](string(code), details))`. Non-typed errors return `(StatusOf(CodeInternal), response.Err[any]("INTERNAL", nil))`.

#### Scenario: Typed error passes through with details intact
- **WHEN** `Respond` receives a `*TypedError{Code: CodeServiceNotFound, Details: {"field": "tenantId"}}`
- **THEN** it returns status 404 and an envelope whose `reason.code` is `SERVICE_NOT_FOUND` and `reason.details.field` is `tenantId`

#### Scenario: Non-typed error becomes INTERNAL with no details
- **WHEN** `Respond` receives a plain `errors.New("boom")`
- **THEN** it returns status 500 and an envelope whose `reason.code` is `INTERNAL` and `reason.details` is empty

#### Scenario: INTERNAL strips cause details
- **WHEN** `Respond` receives a non-typed error whose `Error()` returns a long internal message
- **THEN** the resulting envelope's `reason.details` contains no key carrying that message

### Requirement: `shared/errors` has no domain dependencies

`shared/errors` SHALL import only `net/http`, the Go standard library, and `shared/api/response`. It SHALL NOT import `shared/monitorconfig`, `shared/escalation`, `shared/checkexecution`, `shared/probelocationcatalog`, or any service package.

#### Scenario: shared/errors does not import service packages
- **WHEN** `go list -deps` is run against `shared/errors`
- **THEN** no `services/...` module appears in the dependency list

### Requirement: Wire identity is preserved for every code

Every `CodeXxx` constant SHALL have the same string value as the literal currently emitted by `services/monitor-api/handler.go`. Renaming a constant's string value is a breaking API change.

#### Scenario: Constant value matches wire identity
- **WHEN** the test enumerates every `CodeXxx` constant and compares its `string(...)` value against the literal currently in `services/monitor-api/handler.go`
- **THEN** every comparison succeeds

### Requirement: TypeScript consumers see unchanged code values

The wire identity of every code SHALL be unchanged. The dashboard parser at `apps/dashboard/lib/api.ts:60` continues to populate `ApiError.code` from `response.reason.code` with no modification.

#### Scenario: Dashboard receives unchanged codes
- **WHEN** the dashboard receives a 404 with `reason.code == "SERVICE_NOT_FOUND"`
- **THEN** `ApiError.code == "SERVICE_NOT_FOUND"` exactly as before this change

## MODIFIED Requirements

### Requirement: `Respond` replaces `errResponse` and `serverError`

The `errResponse` and `serverError` helpers in `services/monitor-api/handler.go` SHALL be removed. Every handler error path SHALL route through `shared/errors.Respond(err)`. Handlers SHALL NOT type-assert `*TypedError` directly; they SHALL call `Respond` for every error path.

#### Scenario: Handler routes a typed error through Respond
- **WHEN** the repository returns `*TypedError{Code: CodeServiceAlreadyExists}`
- **THEN** the handler calls `errors.Respond(err)` exactly once and the response body contains `reason.code == "SERVICE_ALREADY_EXISTS"` with status 409

#### Scenario: Handler routes a non-typed error through Respond
- **WHEN** the repository returns a plain error from a DynamoDB call
- **THEN** the handler calls `errors.Respond(err)` and the response body contains `reason.code == "INTERNAL"` with status 500 and no details

### Requirement: Repository sentinels are typed

`services/monitor-api/repository.go` sentinel variables (`errServiceAlreadyExists`, `errMonitorAlreadyExists`, `errCannotDeleteActiveService`, `errCannotDeleteLastMonitorFromActiveService`, `errIncidentNotActionable`, `errMissingTableName`) SHALL be `*TypedError` values with the matching code. The three sites that return `errors.New("service not found")` (in `ArchiveService` and `ReactivateService`) SHALL return `*TypedError{Code: CodeServiceNotFound}`.

#### Scenario: Repository returns a typed sentinel
- **WHEN** `CreateService` rejects a duplicate service
- **THEN** the returned error is `*TypedError{Code: CodeServiceAlreadyExists}`

#### Scenario: Handler no longer matches sentinels via errors.Is
- **WHEN** a grep of `services/monitor-api/handler.go` runs for `errors.Is\(err, err`
- **THEN** no matches exist
