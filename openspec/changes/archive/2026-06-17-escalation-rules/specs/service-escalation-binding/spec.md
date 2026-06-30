## ADDED Requirements

### Requirement: Service references an escalation policy
System SHALL allow each service to reference an escalation policy by ID.

#### Scenario: Service has escalation policy assigned
- **WHEN** a service has an escalationPolicyId set
- **THEN** the service's monitors use that policy for escalation when incidents open

#### Scenario: Service has no escalation policy assigned
- **WHEN** a service has escalationPolicyId set to null
- **THEN** no escalation occurs for incidents on that service's monitors

### Requirement: Service defines business hours configuration
System SHALL allow each service to define business hours for escalation path selection.

#### Scenario: Service defines business hours
- **WHEN** a service has businessHours configuration with timezone, startHour, endHour, and daysOfWeek
- **THEN** system uses this configuration to determine whether current time is within business hours

#### Scenario: Current time is within business hours
- **WHEN** escalation is triggered for an incident on a service
- **AND** current time falls within the service's businessHours configuration
- **THEN** system uses the escalation policy's business-hours path

#### Scenario: Current time is outside business hours
- **WHEN** escalation is triggered for an incident on a service
- **AND** current time falls outside the service's businessHours configuration
- **THEN** system uses the escalation policy's off-hours path

### Requirement: Escalation policy deletion is blocked when in use
System SHALL reject deletion of an escalation policy that is referenced by any service.

#### Scenario: Operator attempts to delete policy in use
- **WHEN** operator calls `DELETE /api/v1/escalation-policies/{id}`
- **AND** one or more services reference this policy
- **THEN** system returns conflict (409) and does not delete the policy

#### Scenario: Operator deletes unreferenced policy
- **WHEN** operator calls `DELETE /api/v1/escalation-policies/{id}`
- **AND** no services reference this policy
- **THEN** system deletes the policy and returns 204
