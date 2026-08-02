## MODIFIED Requirements

### Requirement: Failure carries `reason` with a machine-readable code
On failure, the envelope SHALL include `reason.code` as a stable, machine-readable string. `reason.details` SHALL be an object (Go: `map[string]any`; TypeScript: `Record<string, unknown>`) carrying structured error context. `data` SHALL be absent on failure. `message` SHALL be absent on failure; human-readable detail lives in `reason.details` or is logged server-side. Every HTTP response with a failure status SHALL use this error envelope rather than a success envelope.

#### Scenario: Handler returns typed conflict
- **WHEN** a handler returns a typed conflict error
- **THEN** its HTTP status is `409`
- **AND** its envelope has `status: "error"` and the stable `reason.code`

#### Scenario: Handler error sites route through shared/errors.Respond
- **WHEN** a handler produces an error response
- **THEN** the response body conforms to the envelope shape with `status: "error"` and `reason.code` sourced from a typed `shared/errors.Code` constant
