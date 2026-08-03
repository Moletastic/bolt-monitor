## ADDED Requirements

### Requirement: Documentation separates liveness and readiness
Repository documentation SHALL identify `/api/health` as public liveness and document authenticated v1 readiness verification as a separate operator check.

#### Scenario: Operator follows health guidance
- **WHEN** an operator follows repository health guidance
- **THEN** they can distinguish a running health Lambda from authenticated monitor API availability

### Requirement: Installation and readiness teardown are documented
Repository documentation SHALL provide a from-zero installation guide and a safe teardown order for synthetic readiness credentials and resources.

#### Scenario: Operator removes readiness verification
- **WHEN** an operator no longer needs authenticated readiness verification
- **THEN** the guide disables probes before revoking the synthetic user, removing its membership and SSM parameter, deleting the user, and removing readiness-only IAM access
- **AND** the guide states that no password or token is stored locally
