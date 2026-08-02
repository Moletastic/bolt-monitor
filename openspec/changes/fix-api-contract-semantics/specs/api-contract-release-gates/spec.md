## ADDED Requirements

### Requirement: Contract gates validate API semantics
Deterministic API contract gates SHALL validate static OpenAPI semantic conventions in addition to route and authentication parity.

#### Scenario: Semantic contract regression is introduced
- **WHEN** an operation violates required error-envelope, idempotency-header, or create-location metadata conventions
- **THEN** the gate fails locally without cloud access and identifies the operation and missing convention
