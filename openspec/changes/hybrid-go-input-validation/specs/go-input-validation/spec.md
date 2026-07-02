## ADDED Requirements

### Requirement: API request DTOs use declarative input validation
Go API request DTOs SHALL use `go-playground/validator` tags for request-shape validation when a rule can be evaluated from the decoded request payload without repository access or domain context.

#### Scenario: Required input field is missing
- **WHEN** a monitor API request DTO declares a JSON field as required
- **AND** the decoded request omits or empties that field according to the DTO rule
- **THEN** request validation fails before persistence or domain mutation
- **AND** the response envelope contains `reason.code == "VALIDATION_FAILED"`
- **AND** `reason.details.field` is the JSON field path

#### Scenario: Input string exceeds maximum length
- **WHEN** a monitor API request DTO declares a maximum string length
- **AND** the decoded request supplies a value longer than that limit
- **THEN** request validation fails with `VALIDATION_FAILED`
- **AND** `reason.details.field` identifies the offending JSON field
- **AND** `reason.details.reason` describes the violated length limit

### Requirement: Validator errors map to typed API errors
Failures returned by `go-playground/validator` SHALL be translated into `shared/errors.CodeValidationFailed` typed errors before they reach API response construction.

#### Scenario: Validator rejects a request DTO
- **WHEN** `validator.Struct(dto)` reports a validation failure
- **THEN** the validation adapter returns a typed error with `Code == CodeValidationFailed`
- **AND** the error details include `field` and `reason`
- **AND** `shared/errors.Respond` can produce the standard error envelope from that error

#### Scenario: Nested request field fails validation
- **WHEN** validator rejects a nested request DTO field
- **THEN** the mapped `details.field` uses the nested JSON path rather than Go struct names

### Requirement: Domain validation remains explicit
Persisted/domain structs SHALL continue to express business invariants with explicit `Validate()` methods and `shared/rules`, not `go-playground/validator` tags.

#### Scenario: Monitor domain validation runs after DTO validation
- **WHEN** a create-monitor request passes DTO validation
- **THEN** the request is converted into a domain `Monitor`
- **AND** `Monitor.Validate()` still enforces monitor business invariants such as interval support and HTTP configuration validity

#### Scenario: Domain package dependencies are inspected
- **WHEN** dependencies for `shared/monitorconfig` are listed
- **THEN** `github.com/go-playground/validator/v10` is not included

### Requirement: Repository-backed validation stays explicit
Validation that requires repository reads or external catalog/domain state SHALL remain explicit Go code rather than validator tags.

#### Scenario: Escalation policy references an unknown channel
- **WHEN** an escalation policy request references a channel ID that is absent from the repository
- **THEN** explicit handler or service validation returns `VALIDATION_FAILED`
- **AND** `reason.details.field` identifies the offending path step channel ID

#### Scenario: Monitor references a disabled probe location
- **WHEN** a monitor references a probe location that is unknown or not selectable in the catalog
- **THEN** explicit domain/catalog validation returns `VALIDATION_FAILED`

### Requirement: Validation response contract is unchanged
Introducing `go-playground/validator` SHALL NOT change the API response envelope, `VALIDATION_FAILED` wire identity, or dashboard error parsing contract.

#### Scenario: Dashboard receives validator-backed validation error
- **WHEN** the dashboard receives a validation error from a DTO rule
- **THEN** the response still uses the standard envelope
- **AND** `reason.code == "VALIDATION_FAILED"`
- **AND** dashboard API error parsing continues to read the code from `response.reason.code`
