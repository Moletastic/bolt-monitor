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
System SHALL emit a normalized execution result shape for downstream result and status processing. Failed outbound policy operations SHALL include a stable machine-identifiable failure code and a sanitized operator-safe error message.

#### Scenario: Check finishes
- **WHEN** a healthcheck execution completes
- **THEN** system produces normalized result data describing monitor identity, timing, outcome, and protocol-specific details needed downstream
- **AND** the result does not include probe-location or region identity

#### Scenario: Outbound policy rejects execution
- **WHEN** execution rejects a destination, redirect, timeout, oversized response, or transport operation
- **THEN** the normalized result has a non-success outcome, a stable outbound failure code, and a sanitized error message
- **AND** the result contains no monitor headers, URL credentials, sensitive query values, or response body

### Requirement: HTTP checks execute through the shared outbound policy
System SHALL execute recurring and manual HTTP checks through the same shared public outbound policy, including redirect validation, pinned resolution and dialing, timeout bounds, and bounded response reads.

#### Scenario: Safe public monitor succeeds
- **WHEN** a permitted public target responds within configured bounds and satisfies the monitor expectations
- **THEN** execution preserves the existing successful outcome and status/body assertion behavior

#### Scenario: Persisted target becomes private
- **WHEN** a previously stored hostname resolves to a blocked address at execution time
- **THEN** execution sends no HTTP request and emits a typed blocked-address failure result

#### Scenario: Response body assertion uses bounded input
- **WHEN** execution evaluates `expectedBodyContains`
- **THEN** it reads no more than 1 MiB and fails with the typed oversized-response code if the response exceeds that bound

#### Scenario: Monitor headers do not escape their origin
- **WHEN** a monitor request with configured headers receives a cross-origin or HTTPS-to-HTTP redirect
- **THEN** the redirect is rejected before any request is sent to the redirect target
- **AND** the redirect is not followed by merely stripping the configured headers

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
