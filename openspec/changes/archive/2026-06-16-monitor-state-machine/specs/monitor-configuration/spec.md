## MODIFIED Requirements

### Requirement: System defines required monitor configuration fields
System SHALL require monitor configuration to include enough information to execute a check and manage its lifecycle.

#### Scenario: Required fields are validated
- **WHEN** system validates a monitor configuration
- **THEN** it requires fields for identity, ownership, human-readable name, monitor type, target, cadence, selected probe locations, enabled state, failure threshold, and recovery threshold

## ADDED Requirements

### Requirement: System supports threshold configuration
System SHALL support configuration of failure and recovery thresholds per monitor.

#### Scenario: Threshold fields are validated
- **WHEN** system validates a monitor configuration with FailureThreshold or RecoveryThreshold
- **THEN** it requires both values to be greater than or equal to 1
- **AND** it stores these values for use in state machine transitions
