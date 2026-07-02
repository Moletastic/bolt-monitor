## MODIFIED Requirements

### Requirement: Rules are composable

Business rules and domain invariants SHALL be expressed as `Rule[T] func(T) error` and combined with combinators `All`, `Any`, `Not`, `When`, and a `Builder[T]` that aggregates failures with `errors.Join`. Inline `if` chains in domain code SHALL be replaced with rule chains when the chain is more than one check. Request DTO shape validation MAY use a dedicated input validation adapter instead of `shared/rules` when the rule does not require domain state.

#### Scenario: Domain validation remains rule-based
- **WHEN** a domain struct has more than one business invariant to enforce
- **THEN** the validation is expressed through `shared/rules` composition rather than request DTO validator tags

#### Scenario: Request DTO validation uses the input adapter
- **WHEN** a request DTO has a simple required-field rule that does not require domain state
- **THEN** the request path may validate that rule through the input validation adapter instead of adding a `shared/rules` rule

### Requirement: Field-scoped rules report the field name

Rules that validate a domain struct field SHALL use a `Field[T]` helper that wraps the failure with `details.field` set, so the response envelope's `reason.details` can pinpoint the offending field. Request DTO validation adapters SHALL provide the same `details.field` contract using JSON field paths.

#### Scenario: Domain rule reports a field
- **WHEN** a domain rule rejects a leaf field such as `tenantId`
- **THEN** the emitted typed error includes `details.field == "tenantId"`

#### Scenario: Request DTO adapter reports a JSON field path
- **WHEN** the request DTO validator rejects a nested JSON field
- **THEN** the emitted typed error includes `details.field` using the JSON field path rather than the Go struct field name

### Requirement: Monitor-config validation uses the rules package

`shared/monitorconfig/model.go` `Validate` SHALL be a rule chain built via the `Builder`. The exported signature SHALL be preserved. Failures SHALL surface as `*shared/errors.TypedError` with `code = VALIDATION_FAILED` and `details.field` per failing rule. `shared/monitorconfig` SHALL NOT depend on `go-playground/validator`; request DTO validation SHALL happen before conversion into monitor-config domain structs.

#### Scenario: Monitor config dependency graph excludes DTO validator
- **WHEN** dependencies for `shared/monitorconfig` are listed
- **THEN** `github.com/go-playground/validator/v10` is absent

#### Scenario: Monitor domain validation still runs
- **WHEN** a request DTO passes input validation and is converted to a `Monitor`
- **THEN** `Monitor.Validate()` still enforces monitor configuration invariants through `shared/rules`
