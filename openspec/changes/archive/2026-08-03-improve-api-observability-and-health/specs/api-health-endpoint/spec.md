## MODIFIED Requirements

### Requirement: Health endpoint returns deterministic JSON body
System SHALL return `200 OK` with the standard API success envelope containing a stable machine-readable healthy result. This public endpoint SHALL be liveness-only and SHALL not depend on DynamoDB, authorization, or protected monitor API availability.

#### Scenario: Health check succeeds
- **WHEN** Go Lambda handles a valid request for `/api/health`
- **THEN** response status is `200`
- **AND** response body has `status: "success"`
- **AND** response `data` contains the stable healthy result
- **AND** error-only envelope fields are omitted

#### Scenario: Protected API dependency is unavailable
- **WHEN** a protected API dependency is unavailable
- **THEN** public health remains a liveness response
- **AND** authenticated readiness verification reports the protected-path failure separately
