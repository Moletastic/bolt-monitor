## ADDED Requirements

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