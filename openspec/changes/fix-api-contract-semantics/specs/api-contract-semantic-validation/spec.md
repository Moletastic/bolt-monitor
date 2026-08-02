## ADDED Requirements

### Requirement: Contract validation detects semantic envelope drift
Repository API contract validation SHALL reject OpenAPI error responses that reference a success envelope and SHALL validate declared required idempotency headers and create-response location metadata.

#### Scenario: Error response has success schema
- **WHEN** an OpenAPI non-2xx response references a success envelope
- **THEN** contract validation fails with operation and response status

#### Scenario: Idempotent command lacks header contract
- **WHEN** a command requiring idempotency lacks a required `Idempotency-Key` OpenAPI parameter
- **THEN** contract validation fails with the operation identity
