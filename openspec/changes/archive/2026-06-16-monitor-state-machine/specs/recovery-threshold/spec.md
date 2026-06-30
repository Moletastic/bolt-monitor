## ADDED Requirements

### Requirement: System supports configurable recovery threshold
System SHALL allow configuration of a recovery threshold per monitor that determines how many consecutive successes are required before an open incident resolves.

#### Scenario: Recovery threshold is configurable
- **WHEN** operator creates or updates a monitor
- **THEN** monitor MAY specify a RecoveryThreshold value greater than or equal to 1

#### Scenario: Default recovery threshold is 1
- **WHEN** monitor does not specify RecoveryThreshold
- **THEN** system uses default value of 1, preserving binary success/resolve behavior

#### Scenario: Recovery threshold must be positive
- **WHEN** system validates monitor configuration with RecoveryThreshold
- **THEN** system rejects configuration where RecoveryThreshold is less than 1

#### Scenario: Recovery threshold used in RECOVERING to UP transition
- **WHEN** monitor is in RECOVERING state with ConsecutiveSuccesses >= RecoveryThreshold
- **THEN** monitor transitions to UP state and resolves the open incident
