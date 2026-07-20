## ADDED Requirements

### Requirement: Scheduler execution policy resolves queue resources before deployment
The infrastructure SHALL render the EventBridge Scheduler execution-role policy only after notification queue resource ARNs resolve. The policy SHALL grant `sqs:SendMessage` only to the notification queue and notification DLQ.

#### Scenario: Staging renders escalation scheduler policy
- **WHEN** SST synthesizes the staging stack
- **THEN** the scheduler execution-role policy contains concrete SQS ARNs for the notification queue and notification DLQ
- **AND** AWS accepts the policy document without a malformed resource error

#### Scenario: Policy resource output is unresolved during synthesis
- **WHEN** queue ARNs are represented as deferred infrastructure outputs
- **THEN** policy serialization composes their resolved values before creating IAM JSON
