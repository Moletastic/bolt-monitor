## MODIFIED Requirements

### Requirement: System executes enabled monitors through pipeline
System SHALL execute enabled monitors through a defined execution pipeline that accepts work from both manual-run commands and recurring scheduler decisions.

#### Scenario: Enabled monitor is selected for recurring execution
- **WHEN** execution pipeline evaluates runnable monitors for recurring work
- **THEN** it selects monitors whose lifecycle state is enabled
- **AND** it materializes execution work for each selected monitor and selected probe location

#### Scenario: Accepted manual run feeds shared execution pipeline
- **WHEN** system accepts a manual run command for a runnable monitor
- **THEN** it materializes execution work for that monitor in the same pipeline used by recurring execution

### Requirement: System emits normalized execution result
System SHALL emit a normalized execution result shape for downstream result and status processing after evaluating every configured HTTP assertion supported by the monitor.

#### Scenario: Check finishes with matching status and body expectation
- **WHEN** a healthcheck execution completes with an allowed HTTP status code and configured body-content expectation satisfied
- **THEN** system produces normalized result data describing monitor identity, location, timing, outcome, and protocol-specific details needed downstream
- **AND** it marks the execution outcome as success

#### Scenario: Check finishes with missing expected body content
- **WHEN** a healthcheck execution completes with a configured `http.expectedBodyContains` value that is not present in the response body
- **THEN** system produces normalized result data for that execution
- **AND** it marks the execution outcome as failure

## ADDED Requirements

### Requirement: System enforces runnable-state checks before executing queued work
System SHALL verify that queued execution work is still runnable immediately before executing it.

#### Scenario: Queued work references disabled monitor
- **WHEN** worker claims queued execution work for a monitor that is now disabled
- **THEN** system does not execute the check
- **AND** it records the work as skipped or non-executed without creating a successful check completion

### Requirement: System uses scheduler control state to gate recurring work creation
System SHALL prevent recurring execution work from being created when recurring execution is disabled by admin scheduler configuration.

#### Scenario: Recurring scheduler is globally paused
- **WHEN** recurring scheduler trigger runs while scheduler configuration disables recurring execution
- **THEN** system does not materialize recurring execution work for monitors
