## ADDED Requirements

### Requirement: Scheduler configuration controls recurring work materialization
System SHALL make recurring execution behavior follow the latest persisted scheduler configuration.

#### Scenario: Recurring execution is enabled
- **WHEN** administrator has persisted scheduler configuration with recurring execution enabled and valid stop control
- **THEN** recurring scheduler trigger may materialize execution work for enabled monitors

#### Scenario: Recurring execution is disabled
- **WHEN** administrator has persisted scheduler configuration with recurring execution disabled
- **THEN** recurring scheduler trigger does not materialize recurring execution work for monitors
