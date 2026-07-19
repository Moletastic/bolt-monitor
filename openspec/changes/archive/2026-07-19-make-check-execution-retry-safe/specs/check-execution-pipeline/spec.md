## MODIFIED Requirements

### Requirement: System builds execution work for due monitors
System SHALL conditionally create one canonical execution work record for each accepted manual run or eligible recurring monitor schedule definition/time before any HTTP or queue side effect.

#### Scenario: Enabled monitor is due
- **WHEN** an enabled non-maintenance monitor is eligible under an immutable schedule definition at UTC `scheduledFor`
- **THEN** system derives a stable `runId` from tenant, service, monitor, `scheduleDefinitionVersion`, and `scheduledFor` before side effects
- **AND** conditionally creates one work item for that identity
- **AND** the work item does not include probe-location routing state

#### Scenario: Same schedule definition and time are evaluated again
- **WHEN** scheduler retry evaluates the same monitor, `scheduleDefinitionVersion`, and `scheduledFor`
- **THEN** system resolves the same work identity
- **AND** does not create a second run for that identity

## MODIFIED Requirements

### Requirement: Execution follows an at-least-once durable state machine
System SHALL coordinate execution through conditional `pending`, `in_progress`, `completed`, and `skipped` work transitions with publication state, lease metadata, and terminal metadata.

#### Scenario: Work is created
- **WHEN** system accepts a run identity that does not exist
- **THEN** it conditionally creates pending work with immutable identity and retention metadata

#### Scenario: Work is claimed and completed
- **WHEN** a worker claims pending or lease-expired work and commits a matching result
- **THEN** only that current fencing token may transition work to completed
- **AND** terminal work cannot return to pending or in-progress

#### Scenario: Work becomes ineligible before execution
- **WHEN** current runnable-state validation fails after claim
- **THEN** only the current fencing token may transition work to skipped with a typed reason
- **AND** skipped work cannot produce a result later

### Requirement: Pipeline accepts one canonical observation per run
System SHALL accept no more than one canonical `CheckRun` and terminal result for a stable `runId`, including one stable `runId` per eligible recurring `(scheduleDefinitionVersion, scheduledFor)` identity.

#### Scenario: Worker crashes after HTTP response
- **WHEN** no result transaction committed before the lease expired
- **THEN** another attempt MAY repeat the HTTP request
- **AND** the first successful fenced transaction becomes the only canonical observation

#### Scenario: Result commit response is ambiguous
- **WHEN** a caller retries result commit for the same `runId`
- **THEN** system resolves committed terminal work as idempotent success
- **AND** does not create duplicate downstream effects

### Requirement: Execution identity propagates end to end
System SHALL carry stable run identity, trigger, and recurring `scheduleDefinitionVersion`/UTC `scheduledFor` when applicable through work, execution envelope, normalized result, `CheckRun`, status ordering metadata, incident transition/activity, and canonical notification outbox item.

#### Scenario: Recurring result is traced
- **WHEN** a recurring run causes an incident transition
- **THEN** every persisted execution artifact identifies the same causal `runId`, `scheduleDefinitionVersion`, and `scheduledFor`
- **AND** the transition outbox uses one equal-valued `transitionId`, activity `activityId`, and event `eventId`

#### Scenario: Manual result is traced
- **WHEN** a manual run completes
- **THEN** work, result, and `CheckRun` share its `runId` and `trigger=manual`
- **AND** no `scheduleDefinitionVersion` or `scheduledFor` is fabricated

### Requirement: Recoverable execution state is directly queryable
System SHALL represent pending publication and nonterminal work with directly queryable marker items in bounded tenant-scoped recovery partitions and SHALL NOT require an unbounded tenant or table scan.

#### Scenario: Recovery finds due work
- **WHEN** scheduler or worker recovery searches configured current/overlap time buckets and shards
- **THEN** it uses bounded paginated DynamoDB queries to resolve canonical work

#### Scenario: Work reaches publication-complete or terminal state
- **WHEN** publication is acknowledged or work commits completed/skipped
- **THEN** the corresponding marker is conditionally removed in the same durable transition
- **AND** stale recovery cannot remove or mutate newer fenced state

### Requirement: Pipeline exposes typed operational failures
System SHALL represent expected execution conflicts, skips, stale observations, lease loss, and retryable persistence/publication failures as typed values with stable classification and safe context.

#### Scenario: Caller receives an operational failure
- **WHEN** a scheduler, worker, manual run, or execution recovery operation fails
- **THEN** the failure identifies operation, retryability, and causal run when known
- **AND** secrets, monitor headers, and response bodies are excluded from error details

### Requirement: Retry invariants are fault-injection tested
System SHALL verify retry safety by injecting failures at every cross-system boundary.

#### Scenario: Failure and retry matrix executes
- **WHEN** tests inject failures before and after work/marker creation, execution queue send, publication acknowledgement, claim, HTTP response, result/outbox commit, and marker removal
- **THEN** retries leave one work identity per recurring schedule definition/time, at most one canonical `CheckRun`, monotonic recurring projections, and one equal-valued transition/activity/event identity with one outbox item

#### Scenario: Concurrency and ordering matrix executes
- **WHEN** tests exercise duplicate scheduler invocations, concurrent SQS delivery, lease reclaim, stale fenced completion, multipage partial failure, and out-of-order results
- **THEN** conditional invariants hold without FIFO assumptions or AWS error-string matching
