## ADDED Requirements

### Requirement: Expired ephemeral targets remain eligible for verified removal
The infrastructure orchestrator SHALL permit an expired target only when it is structurally valid, `ephemeral`, and `disposable=true`, and only for exact-stage removal plus residual verification. It SHALL reject that expired target before development, deployment, status, invitation, or credential/key mutation.

#### Scenario: Operator removes expired ephemeral stage
- **WHEN** an operator selects an otherwise valid expired disposable target for `make remove-infra`
- **THEN** the orchestrator runs the existing exact-stage SST removal and residual verification
- **AND** does not require expiry to be in the future

#### Scenario: Operator deploys expired ephemeral stage
- **WHEN** an operator selects an expired ephemeral target for deployment or development
- **THEN** validation fails before AWS mutation
- **AND** the error identifies expiry without exposing credentials
