## ADDED Requirements

### Requirement: Scheduler respects per-monitor intervalSeconds
System SHALL only enqueue a monitor for execution when sufficient time has elapsed since its last execution, as determined by the monitor's supported minute-based `intervalSeconds` configuration.

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
- **THEN** scheduler SHALL treat the monitor as always due for backward compatibility with existing persisted data

### Requirement: Scheduler records LastExecutionAt after enqueuing
System SHALL update `LastExecutionAt` to the current timestamp immediately after successfully sending a message to the SQS queue.

#### Scenario: LastExecutionAt recorded after SQS send
- **WHEN** scheduler successfully sends execution request to SQS queue
- **THEN** scheduler SHALL update the monitor's `LastExecutionAt` in DynamoDB to the current timestamp

#### Scenario: LastExecutionAt not recorded on SQS failure
- **WHEN** scheduler fails to send message to SQS queue
- **THEN** scheduler SHALL NOT update `LastExecutionAt`
- **AND** SHALL return error to trigger EventBridge retry

### Requirement: LastExecutionAt tracked per monitor
System SHALL track `LastExecutionAt` independently for each monitor, scoped by tenant, service, and monitor ID.

#### Scenario: LastExecutionAt is monitor-scoped
- **WHEN** scheduler records last execution for a monitor
- **THEN** the timestamp SHALL be associated with that specific monitor's identity
- **AND** SHALL NOT affect other monitors on the same service
