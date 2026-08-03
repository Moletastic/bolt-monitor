## ADDED Requirements

### Requirement: Operators can verify authenticated API readiness
The repository SHALL provide a low-frequency authenticated verification of a protected v1 read operation separate from public liveness.

#### Scenario: Authenticated API is available
- **WHEN** readiness verification runs with valid operator credentials
- **THEN** it verifies an authenticated v1 read response through API Gateway
- **AND** does not print credentials or response secrets

### Requirement: Readiness setup is target-scoped and idempotent
The repository SHALL provide `make setup-readiness` and `make readiness-api`. Infrastructure SHALL expose a dedicated Cognito `USER_PASSWORD_AUTH` app client for one target-scoped synthetic operator. Setup SHALL create or reuse that operator with minimum existing protected-read membership, store its generated password only in deterministic target-scoped SSM SecureString storage, and not print passwords or tokens. `ROTATE=yes` SHALL rotate the generated password and overwrite the same parameter.

#### Scenario: Operator reruns readiness setup
- **WHEN** an operator runs `make setup-readiness` for an already configured target without `ROTATE=yes`
- **THEN** the command reuses the synthetic identity and existing SSM parameter
- **AND** does not reset its password
