## Requirements

### Requirement: System executes enabled monitors through pipeline
System SHALL execute enabled monitors through a defined execution pipeline.

#### Scenario: Enabled monitor is selected for execution
- **WHEN** execution pipeline evaluates runnable monitors
- **THEN** it selects monitors whose lifecycle state is enabled

### Requirement: System builds execution work for due monitors
System SHALL create execution work for each enabled monitor that is due to run.

#### Scenario: Enabled monitor is due
- **WHEN** a monitor is enabled and due for execution
- **THEN** system creates one work item for that monitor and run
- **AND** the work item does not include probe-location routing state

### Requirement: System emits normalized execution result
System SHALL emit a normalized execution result shape for downstream result and status processing.

#### Scenario: Check finishes
- **WHEN** a healthcheck execution completes
- **THEN** system produces normalized result data describing monitor identity, timing, outcome, and protocol-specific details needed downstream
- **AND** the result does not include probe-location or region identity

### Requirement: Disabled monitors must not execute
System SHALL prevent disabled monitors from periodic or manual scheduling paths that are meant for active monitoring.

#### Scenario: Monitor is disabled
- **WHEN** monitor lifecycle state is disabled
- **THEN** periodic execution pipeline does not execute that monitor

### Requirement: Periodic monitoring requires stop control
System SHALL NOT enable recurring healthcheck execution unless the system provides a reliable way to stop checks at any time.

#### Scenario: Periodic execution is configured
- **WHEN** system enables recurring monitor execution
- **THEN** operators can stop ongoing future executions through monitor disablement or equivalent stop control without waiting for code changes

### Requirement: System expires persisted execution work records
System SHALL attach TTL metadata to persisted execution work records so transient scheduler and worker coordination state is automatically removed after its operational troubleshooting window.

#### Scenario: Execution work is persisted
- **WHEN** system creates or updates an execution work record
- **THEN** the record includes numeric Unix epoch-second TTL metadata
- **AND** the TTL is later than the work record's accepted timestamp by the configured execution-work retention window

#### Scenario: Execution work retention elapses
- **WHEN** an execution work record reaches its TTL timestamp
- **THEN** the record is eligible for automatic deletion by DynamoDB Time to Live
- **AND** execution result history remains represented by `CheckRun` records, not by retained execution work records
