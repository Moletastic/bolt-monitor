## ADDED Requirements

### Requirement: Monitor API responses expose generated request identity
The monitor API SHALL return `X-Request-Id` on success and error responses using the API Gateway request ID or a generated fallback for direct invocation.

#### Scenario: Client receives API error
- **WHEN** a monitor API request produces a response
- **THEN** the response includes `X-Request-Id`
- **AND** the ID is safe to quote when seeking operator support

### Requirement: Monitor API boundary logs correlate requests safely
The monitor API SHALL emit structured boundary logs with request ID, Lambda invocation ID when available, method, route, status, and safe tenant or subject context.

#### Scenario: Client supplies correlation ID
- **WHEN** a request includes `X-Correlation-ID`
- **THEN** logs may record it as secondary metadata
- **AND** it does not replace the generated request ID
