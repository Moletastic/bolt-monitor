## MODIFIED Requirements

### Requirement: System stores latest monitor status snapshot
System SHALL store a latest-status snapshot for each monitor including state machine counters.

#### Scenario: New result is processed
- **WHEN** system processes a completed execution result
- **THEN** it updates the monitor's current status snapshot with latest derived state, consecutive failure/success counters, and current state (UP/DEGRADED/DOWN/RECOVERING/MAINTENANCE)

#### Scenario: State machine counters are persisted
- **WHEN** system updates monitor status snapshot
- **THEN** it persists ConsecutiveFailures, ConsecutiveSuccesses, and CurrentStatus fields
- **AND** these values reflect the accumulated outcome history since last counter reset
