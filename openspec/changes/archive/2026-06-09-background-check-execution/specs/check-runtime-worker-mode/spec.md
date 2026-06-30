## ADDED Requirements

### Requirement: Worker Lambda operates in worker mode based on RUNTIME_MODE
System SHALL configure worker Lambda with RUNTIME_MODE environment variable set to "worker".

#### Scenario: Worker mode invocation
- **WHEN** SQS triggers the worker Lambda with a message
- **THEN** Lambda reads RUNTIME_MODE=worker from environment
- **AND** executes the worker workflow

### Requirement: Worker receives execution request from SQS
System SHALL parse the SQS message body as an ExecutionRequest.

#### Scenario: Message parsing
- **WHEN** worker receives SQS event with message
- **THEN** it unmarshals the message body as ExecutionRequest JSON
- **AND** extracts monitor config, probe location, runId, and trigger

### Requirement: Worker executes HTTP check against target
System SHALL execute the HTTP check as specified in the monitor configuration.

#### Scenario: HTTP execution
- **WHEN** worker has parsed ExecutionRequest
- **THEN** it calls checkexecution.ExecuteHTTP() with:
  - HTTP client configured with monitor's timeoutMs
  - Request built from monitor's HTTP configuration (target, method, headers, expectedStatusCodes, expectedBodyContains)

### Requirement: Worker records result to DynamoDB
System SHALL record execution result to DynamoDB after HTTP check completes.

#### Scenario: Result recording
- **WHEN** HTTP check completes (success or failure)
- **THEN** worker calls RecordExecutionResult with:
  - Monitor configuration
  - Run ID from request
  - Probe location ID
  - Trigger type (recurring)
  - ExecutionResult (outcome, durationMs, statusCode, error, timestamps)

### Requirement: Worker handles result recording failures
System SHALL NOT delete SQS message if DynamoDB write fails.

#### Scenario: DynamoDB write failure
- **WHEN** worker completes HTTP check but fails to record result
- **THEN** it does NOT delete the SQS message
- **AND** allows SQS to retry the message (visibility timeout)

### Requirement: Worker deletes SQS message on successful processing
System SHALL delete the SQS message after successfully recording the result.

#### Scenario: Successful processing
- **WHEN** worker completes HTTP check and records result to DynamoDB
- **THEN** it deletes the SQS message
- **AND** returns success