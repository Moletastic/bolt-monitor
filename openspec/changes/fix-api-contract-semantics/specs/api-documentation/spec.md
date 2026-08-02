## ADDED Requirements

### Requirement: OpenAPI describes HTTP operation semantics
The source-controlled OpenAPI document SHALL distinguish success and error envelopes and document request bodies, path and query parameters, cursor pagination, required idempotency headers, and `Location` headers where these apply.

#### Scenario: Developer inspects a create operation
- **WHEN** a developer reads an OpenAPI create operation
- **THEN** it documents request body, `201` success envelope, `Location` header, and applicable typed failures

#### Scenario: Developer inspects an error operation
- **WHEN** a developer reads a documented non-2xx response
- **THEN** it references the standard error envelope rather than a success schema
