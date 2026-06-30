## ADDED Requirements

### Requirement: System supports maintenance mode for monitors
System SHALL allow operators to place a monitor in MAINTENANCE state, silencing notifications for open incidents without closing them.

#### Scenario: Operator enables maintenance mode
- **WHEN** operator calls maintenance enable endpoint for a monitor
- **THEN** monitor transitions to MAINTENANCE state
- **AND** ConsecutiveFailures and ConsecutiveSuccesses are reset to 0
- **AND** open incidents remain open but notification events are suppressed

#### Scenario: Operator disables maintenance mode
- **WHEN** operator calls maintenance disable endpoint for a monitor in MAINTENANCE state
- **THEN** monitor transitions to UP state
- **AND** system re-evaluates current check result to determine next state

#### Scenario: MAINTENANCE state suppresses notification events
- **WHEN** monitor is in MAINTENANCE state
- **THEN** system SHALL NOT emit incident.opened or incident.resolved notification events
- **AND** open incidents persist in DynamoDB

#### Scenario: Monitor in MAINTENANCE does not execute checks
- **WHEN** monitor is in MAINTENANCE state
- **THEN** periodic execution pipeline SHALL NOT schedule check execution for that monitor
- **AND** manual runs are still allowed but do not affect counters or state transitions

#### Scenario: MAINTENANCE disable re-evaluates from current state
- **WHEN** monitor exits MAINTENANCE state
- **THEN** system evaluates the most recent check result to determine next state transition
- **AND** if that result is a failure, monitor transitions to DEGRADED (not directly to DOWN)
