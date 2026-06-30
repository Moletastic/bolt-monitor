## MODIFIED Requirements

### Requirement: System stores raw check execution results
System SHALL store raw execution results for completed healthchecks from both manual and recurring execution paths.

#### Scenario: Execution completes
- **WHEN** monitor execution finishes
- **THEN** system persists a `CheckRun` record containing monitor identity, probe location, timing, trigger, and outcome data

### Requirement: System stores latest monitor status snapshot
System SHALL store a latest-status snapshot for each monitor whenever a completed execution result is processed.

#### Scenario: New result is processed
- **WHEN** system processes a completed execution result
- **THEN** it updates the monitor's current status snapshot with the latest derived state from that result

## ADDED Requirements

### Requirement: System persists run history and latest status together for one completed execution
System SHALL persist append-only run history and mutable latest status as one completion unit for each finished execution.

#### Scenario: Execution completion is recorded
- **WHEN** system records a finished execution result
- **THEN** it writes `CheckRun` history and latest `MonitorStatus` snapshot for that same result without leaving only one of those records as the final stored outcome

### Requirement: System preserves manual-run correlation in stored results
System SHALL make accepted manual run identifiers observable through downstream stored run history.

#### Scenario: Operator reviews history after manual run
- **WHEN** an accepted manual run completes and operator requests recent monitor runs
- **THEN** system returns a persisted run record whose identity can be correlated back to the `runId` returned when the manual run was accepted
