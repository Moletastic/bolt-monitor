# check-runtime-scheduler-mode Specification

## Purpose
TBD - created by archiving change background-check-execution. Update Purpose after archive.
## Requirements
### Requirement: Scheduler Lambda operates in scheduler mode based on RUNTIME_MODE
System SHALL configure scheduler Lambda with RUNTIME_MODE environment variable set to "scheduler".

#### Scenario: Scheduler mode invocation
- **WHEN** EventBridge triggers the scheduler Lambda
- **THEN** Lambda reads RUNTIME_MODE=scheduler from environment
- **AND** executes the scheduler workflow

### Requirement: Scheduler sends messages to SQS queue
System SHALL send execution requests to SQS queue in addition to DynamoDB audit trail.

#### Scenario: Scheduler sends to SQS
- **WHEN** scheduler builds ExecutionRequests for enabled monitors
- **THEN** it sends each request to the SQS execution-queue
- **AND** continues to write RUN_REQUEST# items to DynamoDB for audit

### Requirement: Scheduler maintains audit trail in DynamoDB
System SHALL write RUN_REQUEST# items to DynamoDB even when sending to SQS.

#### Scenario: DynamoDB audit trail
- **WHEN** scheduler sends execution request to SQS
- **THEN** it also writes a RUN_REQUEST# item to DynamoDB
- **AND** the item includes runId, monitorId, serviceId, probeLocationId, trigger, acceptedAt, status=pending

### Requirement: Scheduler handles SQS send failures gracefully
System SHALL return error if SQS send fails, preventing DynamoDB write.

#### Scenario: SQS send failure
- **WHEN** scheduler fails to send message to SQS
- **THEN** it does NOT write RUN_REQUEST# item to DynamoDB
- **AND** returns error to trigger EventBridge retry

### Requirement: Scheduler respects per-monitor intervalSeconds
System SHALL only enqueue a monitor for execution when sufficient time has elapsed since its last execution, as determined by the monitor's `intervalSeconds` configuration.

#### Scenario: Monitor due (first execution)
- **WHEN** scheduler evaluates a monitor where `LastExecutionAt` is null
- **THEN** scheduler SHALL enqueue the monitor for execution

#### Scenario: Monitor due (interval elapsed)
- **WHEN** scheduler evaluates a monitor where `time.Since(LastExecutionAt) >= intervalSeconds`
- **THEN** scheduler SHALL enqueue the monitor for execution

#### Scenario: Monitor not yet due (interval not elapsed)
- **WHEN** scheduler evaluates a monitor where `time.Since(LastExecutionAt) < intervalSeconds`
- **THEN** scheduler SHALL skip that monitor

#### Scenario: Monitor with zero intervalSeconds
- **WHEN** scheduler evaluates a monitor where `intervalSeconds` is 0 or not set
- **THEN** scheduler SHALL treat the monitor as always due (backward compatible behavior)

### Requirement: Scheduler records LastExecutionAt after enqueuing
System SHALL update `LastExecutionAt` to the current timestamp immediately after successfully sending a message to the SQS queue.

#### Scenario: LastExecutionAt recorded after SQS send
- **WHEN** scheduler successfully sends execution request to SQS queue
- **THEN** scheduler SHALL update the monitor's `LastExecutionAt` in DynamoDB to the current timestamp

#### Scenario: LastExecutionAt not recorded on SQS failure
- **WHEN** scheduler fails to send message to SQS queue
- **THEN** scheduler SHALL NOT update `LastExecutionAt`
- **AND** SHALL return error to trigger EventBridge retry

