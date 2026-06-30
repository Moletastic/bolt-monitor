## MODIFIED Requirements

### Requirement: EventBridge Schedule triggers scheduler Lambda periodically
System SHALL trigger the scheduler Lambda once per minute to initiate recurring execution.

#### Scenario: Scheduler triggers on schedule
- **WHEN** EventBridge Schedule fires at configured rate (`rate(1 minute)`)
- **THEN** scheduler Lambda is invoked with CloudWatch Event input

### Requirement: Scheduler reads scheduler configuration before enqueuing
System SHALL read SchedulerConfig from DynamoDB before enqueuing work to ensure recurring execution is enabled.

#### Scenario: Recurring disabled
- **WHEN** EventBridge fires and SchedulerConfig.RecurringEnabled is false
- **THEN** scheduler exits without enqueuing any work

#### Scenario: Recurring enabled
- **WHEN** EventBridge fires and SchedulerConfig.RecurringEnabled is true
- **THEN** scheduler proceeds to read monitors and enqueue execution requests

### Requirement: Scheduler reads all enabled monitors
System SHALL query DynamoDB for all monitors where enabled equals true.

#### Scenario: Scheduler queries enabled monitors
- **WHEN** scheduler processes an EventBridge trigger
- **THEN** it queries DynamoDB for all monitors with enabled flag set to true
- **AND** filters to monitors for the current tenant

### Requirement: Scheduler creates execution requests for each monitor-probe combination
System SHALL build an ExecutionRequest for each enabled monitor and each probe location assigned to that monitor.

#### Scenario: Execution requests built
- **WHEN** scheduler has list of enabled monitors
- **THEN** for each monitor, for each probeLocation in monitor.ProbeLocations, an ExecutionRequest is created
- **AND** trigger type is set to "recurring"

### Requirement: Scheduler uses a single EventBridge Schedule
System SHALL configure a single EventBridge Schedule for recurring execution.

#### Scenario: Single scheduler configured
- **WHEN** infrastructure is deployed
- **THEN** exactly one recurring scheduler SHALL invoke the scheduler Lambda with `RUNTIME_MODE=scheduler`

#### Scenario: No sub-minute scheduler configured
- **WHEN** infrastructure is deployed
- **THEN** no second 30-second-offset EventBridge schedule SHALL be configured
