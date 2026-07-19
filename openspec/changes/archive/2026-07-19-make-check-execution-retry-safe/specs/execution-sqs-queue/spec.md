## MODIFIED Requirements

### Requirement: System uses SQS queue for execution work distribution
System SHALL use the existing standard SQS execution queue as an at-least-once wake-up transport for execution work whose canonical identity and state are persisted in DynamoDB.

#### Scenario: Scheduler sends execution request to queue
- **WHEN** scheduler accepts execution work for an enabled eligible monitor
- **THEN** it conditionally persists the work before sending an execution envelope to the execution queue
- **AND** the queue message and work record carry the same stable `runId`

#### Scenario: Queue send succeeds ambiguously
- **WHEN** the scheduler cannot determine whether a send succeeded or cannot persist publication acknowledgement after a successful send
- **THEN** retry MAY publish the same execution envelope again
- **AND** duplicate publication does not create a second work identity or canonical result

### Requirement: SQS message contains full execution context
System SHALL serialize the immutable execution identity and scheduling context needed to locate canonical work as the SQS message body; mutable monitor configuration in the message SHALL NOT be authoritative for execution.

#### Scenario: Recurring message structure
- **WHEN** scheduler sends a recurring message to the execution queue
- **THEN** the message contains tenant, service, and monitor identity, `runId`, `trigger=recurring`, `acceptedAt`, `scheduleDefinitionVersion`, and UTC `scheduledFor`
- **AND** those immutable fields match the persisted execution work
- **AND** worker reloads mutable monitor configuration from DynamoDB before the HTTP request

#### Scenario: Manual message structure
- **WHEN** manual work is distributed through the execution queue
- **THEN** the message contains the same stable identity and `trigger=manual`
- **AND** it does not claim `scheduleDefinitionVersion` or `scheduledFor`

### Requirement: Worker receives and processes SQS messages
System SHALL configure SQS to trigger worker Lambda when messages are available and SHALL treat each delivery as a request to advance the identified durable work state.

#### Scenario: Worker triggered by SQS
- **WHEN** a message exists in the execution queue
- **THEN** SQS invokes the worker Lambda with the message
- **AND** worker loads and conditionally claims the matching work before executing a check

#### Scenario: Terminal work is redelivered
- **WHEN** SQS delivers an envelope whose work is already completed or skipped
- **THEN** worker performs no additional HTTP check or result projection
- **AND** any canonical pending notification outbox item remains owned by the notification-assurance dispatcher rather than execution redelivery

## MODIFIED Requirements

### Requirement: Execution publication is recoverable
System SHALL retain publication state on nonterminal execution work so a persisted request cannot be lost solely because queue publication failed.

#### Scenario: Work persists and send fails
- **WHEN** work creation succeeds and the execution queue send fails
- **THEN** work remains publication-pending with the same `runId`
- **AND** a scheduler retry or bounded recovery pass republishes that work instead of creating a new run

#### Scenario: Publication acknowledgement persists
- **WHEN** an execution envelope is sent successfully
- **THEN** system conditionally marks that work publication as acknowledged
- **AND** a duplicate acknowledgement is an idempotent no-op

### Requirement: Standard queue delivery does not imply FIFO or exactly once
System SHALL remain correct under duplicate and out-of-order deliveries from the existing standard SQS queue without claiming FIFO or exactly-once execution.

#### Scenario: Two workers receive the same run
- **WHEN** duplicate messages for one `runId` are delivered concurrently
- **THEN** conditional claim and lease fencing allow at most one current claimant to commit
- **AND** the system accepts at most one canonical result for the run

### Requirement: Execution timing and concurrency configuration is ordered and bounded
System SHALL define named deployment configuration such that `WORKER_LAMBDA_TIMEOUT > MAX_OUTBOUND_EXECUTION + RESULT_COMMIT_BUFFER`, `EXECUTION_QUEUE_VISIBILITY_TIMEOUT > WORKER_LAMBDA_TIMEOUT + VISIBILITY_MARGIN`, and `WORK_LEASE_DURATION > MAX_OUTBOUND_EXECUTION + RESULT_COMMIT_BUFFER`. The execution event source SHALL have finite `EXECUTION_EVENT_SOURCE_MAX_CONCURRENCY` and partial batch responses enabled.

#### Scenario: Infrastructure configuration is deployed
- **WHEN** execution queue and worker settings are synthesized or validated
- **THEN** all strict inequalities hold within AWS limits
- **AND** named values are documented and covered by boundary tests

#### Scenario: One batch record fails
- **WHEN** a batch contains both completed and retryable execution records
- **THEN** `ReportBatchItemFailures` reports only failed record identifiers
- **AND** bounded event-source concurrency limits simultaneous worker invocations
