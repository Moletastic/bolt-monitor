## ADDED Requirements

### Requirement: System stores raw check execution results
System SHALL store raw execution results for completed healthchecks.

#### Scenario: Execution completes
- **WHEN** monitor execution finishes
- **THEN** system persists a `CheckRun` record containing monitor identity, probe location, timing, and outcome data

### Requirement: System stores latest monitor status snapshot
System SHALL store a latest-status snapshot for each monitor.

#### Scenario: New result is processed
- **WHEN** system processes a completed execution result
- **THEN** it updates the monitor's current status snapshot with the latest derived state

### Requirement: System keeps raw runs and latest status as different storage concerns
System SHALL distinguish append-only run history from mutable current status state.

#### Scenario: Recent results are queried
- **WHEN** application needs historical run data and current state
- **THEN** it can read raw run history separately from latest status snapshot

### Requirement: System defines retention for raw run history
System SHALL define raw run retention expectations for high-volume `CheckRun` records.

#### Scenario: Raw runs accumulate over time
- **WHEN** system persists ongoing execution results
- **THEN** raw run items include documented TTL or retention handling
