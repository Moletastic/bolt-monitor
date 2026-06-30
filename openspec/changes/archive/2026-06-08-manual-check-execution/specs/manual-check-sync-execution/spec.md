## ADDED Requirements

### Requirement: Manual run executes HTTP check synchronously
System SHALL execute the monitor's HTTP check synchronously within the API Lambda and return real execution results in the response.

#### Scenario: Operator triggers monitor run for valid target
- **WHEN** operator calls `POST /api/v1/services/{serviceId}/monitors/{monitorId}/run` for an enabled monitor with valid HTTP configuration
- **THEN** system executes the HTTP request against the configured target URL
- **AND** waits for the response (up to configured timeout)
- **AND** returns execution result with outcome, duration, status code, and timestamps

#### Scenario: Operator triggers monitor run for unreachable target
- **WHEN** operator calls manual run for a monitor targeting an unreachable URL
- **THEN** system returns execution result with outcome "error" or "timeout"
- **AND** includes error message describing the failure

### Requirement: Manual run response includes execution details
System SHALL return complete execution metadata in the manual run response.

#### Scenario: Successful execution response
- **WHEN** operator triggers a manual run that succeeds
- **THEN** response includes:
  - `runId`: stable identifier for the run
  - `outcome`: "success"
  - `durationMs`: execution time in milliseconds
  - `statusCode`: HTTP status code from target
  - `error`: null
  - `probeLocationId`: location used for execution (e.g., "iad")
  - `startedAt`: RFC3339 timestamp when execution started
  - `finishedAt`: RFC3339 timestamp when execution completed

#### Scenario: Failed execution response
- **WHEN** operator triggers a manual run that fails (non-2xx status, body mismatch, timeout, connection error)
- **THEN** response includes:
  - `runId`: stable identifier for the run
  - `outcome`: "failure", "timeout", or "error"
  - `durationMs`: execution time in milliseconds
  - `statusCode`: HTTP status code if available, null otherwise
  - `error`: description of failure reason
  - `probeLocationId`: location used for execution
  - `startedAt`: RFC3339 timestamp when execution started
  - `finishedAt`: RFC3339 timestamp when execution completed

### Requirement: Manual run result is recorded to DynamoDB
System SHALL persist the execution result to DynamoDB in the same format as automated executions.

#### Scenario: Execution result is persisted
- **WHEN** manual execution completes (success or failure)
- **THEN** system writes a CheckRun record with full execution details
- **AND** updates the MonitorStatus record with latest outcome
- **AND** handles incident creation/resolution per existing incident rules

### Requirement: Manual run uses manual trigger type
System SHALL mark the execution with trigger type "manual" distinct from "recurring".

#### Scenario: Execution metadata includes trigger type
- **WHEN** system records the manual execution result
- **THEN** the trigger field is set to "manual"