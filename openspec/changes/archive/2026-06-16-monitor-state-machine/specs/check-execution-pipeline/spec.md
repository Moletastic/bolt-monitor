## MODIFIED Requirements

### Requirement: System executes enabled monitors through pipeline
System SHALL execute enabled monitors through a defined execution pipeline.

#### Scenario: Enabled monitor is selected for execution
- **WHEN** execution pipeline evaluates runnable monitors
- **THEN** it selects monitors whose lifecycle state is enabled AND whose monitor state is not MAINTENANCE

### Requirement: System emits normalized execution result
System SHALL emit a normalized execution result shape for downstream result and status processing.

#### Scenario: Check finishes
- **WHEN** a healthcheck execution completes
- **THEN** system produces normalized result data describing monitor identity, location, timing, outcome, and protocol-specific details needed downstream
- **AND** if the trigger was manual, the result indicates manual trigger type so state machine excludes it from counter accumulation

### Requirement: Disabled monitors must not execute
System SHALL prevent disabled monitors from periodic or manual scheduling paths that are meant for active monitoring.

#### Scenario: Monitor is disabled
- **WHEN** monitor lifecycle state is disabled
- **THEN** periodic execution pipeline does not execute that monitor

#### Scenario: Monitor is in MAINTENANCE
- **WHEN** monitor state is MAINTENANCE
- **THEN** periodic execution pipeline does not execute that monitor
- **AND** manual runs are still permitted for diagnostic purposes
