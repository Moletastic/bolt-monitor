## ADDED Requirements

### Requirement: System provides escalation policy CRUD API
System SHALL allow operators to create, read, update, and delete escalation policies through HTTP API.

#### Scenario: Operator creates escalation policy
- **WHEN** operator calls `POST /api/v1/escalation-policies` with valid policy body
- **THEN** system creates the escalation policy and returns the created resource with generated ID

#### Scenario: Operator lists escalation policies
- **WHEN** operator calls `GET /api/v1/escalation-policies`
- **THEN** system returns all escalation policies in the current tenant

#### Scenario: Operator reads one escalation policy
- **WHEN** operator calls `GET /api/v1/escalation-policies/{id}` for existing policy
- **THEN** system returns the full escalation policy resource

#### Scenario: Operator updates escalation policy
- **WHEN** operator calls `PUT /api/v1/escalation-policies/{id}` with valid policy body
- **THEN** system updates the escalation policy and returns the updated resource

#### Scenario: Operator deletes escalation policy
- **WHEN** operator calls `DELETE /api/v1/escalation-policies/{id}` for existing policy
- **THEN** if no service references this policy, system deletes it and returns 204
- **AND** if any service references this policy, system returns conflict and does not delete

### Requirement: System requires escalation policy to have at least one step
System SHALL validate that escalation policies contain at least one step in both business-hours and off-hours paths.

#### Scenario: Operator creates policy with empty business-hours path
- **WHEN** operator creates an escalation policy with an empty business-hours path steps array
- **THEN** system rejects the request with validation error

#### Scenario: Operator creates policy with valid steps
- **WHEN** operator creates an escalation policy with at least one step in business-hours path
- **THEN** system accepts the policy

### Requirement: System requires channel configuration for each step
System SHALL validate that each escalation step specifies at least one notification channel.

#### Scenario: Operator creates step with no channels
- **WHEN** operator creates an escalation step with empty channels array
- **THEN** system rejects the request with validation error

### Requirement: System supports multiple channel types per step
System SHALL allow each escalation step to specify multiple notification channels of different types.

#### Scenario: Step specifies telegram and email channels
- **WHEN** escalation step specifies channels of type telegram and email
- **THEN** system fires both channels when the step executes
