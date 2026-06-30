## ADDED Requirements

### Requirement: System uses SQS queue for execution work distribution
System SHALL use an SQS queue to distribute execution work from scheduler to worker.

#### Scenario: Scheduler sends execution request to queue
- **WHEN** scheduler builds ExecutionRequests for enabled monitors
- **THEN** it sends each request as a JSON message to the execution-queue SQS queue

### Requirement: SQS queue has configurable visibility timeout
System SHALL configure SQS queue visibility timeout sufficient for HTTP check execution plus buffer.

#### Scenario: Message visibility timeout
- **WHEN** worker receives a message from the queue
- **THEN** the message is not visible to other consumers for 60 seconds (visibility timeout)
- **AND** if worker does not delete the message within 60 seconds, it becomes visible again

### Requirement: SQS queue has dead letter queue for failed messages
System SHALL configure a dead letter queue for messages that fail processing after maximum receive count.

#### Scenario: Message moved to DLQ after max receives
- **WHEN** a message is received and processed unsuccessfully 3 times
- **THEN** SQS moves the message to execution-queue-dlq

### Requirement: SQS message contains full execution context
System SHALL serialize the complete ExecutionRequest as the SQS message body.

#### Scenario: Message structure
- **WHEN** scheduler sends a message to the execution-queue
- **THEN** message body contains:
  - `monitor`: full monitor configuration object
  - `probeLocation`: probe location details
  - `runId`: unique identifier for this execution
  - `trigger`: "recurring" for scheduled executions

### Requirement: Worker receives and processes SQS messages
System SHALL configure SQS to trigger worker Lambda when messages are available.

#### Scenario: Worker triggered by SQS
- **WHEN** a message exists in the execution-queue
- **THEN** SQS invokes the worker Lambda with the message
- **AND** worker processes the execution request