## ADDED Requirements

### Requirement: System supports configurable failure threshold
System SHALL allow configuration of a failure threshold per monitor that determines how many consecutive failures are required before an incident opens.

#### Scenario: Failure threshold is configurable
- **WHEN** operator creates or updates a monitor
- **THEN** monitor MAY specify a FailureThreshold value greater than or equal to 1

#### Scenario: Default failure threshold is 1
- **WHEN** monitor does not specify FailureThreshold
- **THEN** system uses default value of 1, preserving binary failure/incident behavior

#### Scenario: Failure threshold must be positive
- **WHEN** system validates monitor configuration with FailureThreshold
- **THEN** system rejects configuration where FailureThreshold is less than 1

#### Scenario: Failure threshold used in DEGRADED to DOWN transition
- **WHEN** monitor is in DEGRADED state with ConsecutiveFailures >= FailureThreshold
- **THEN** monitor transitions to DOWN state and opens an incident
