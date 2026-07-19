## Purpose

Define the at-least-once retry-safe execution pipeline that turns accepted monitor runs into canonical observations, ordered projections, and transition outbox items.
## Requirements
### Requirement: System executes enabled monitors through pipeline
System SHALL execute enabled monitors through a defined execution pipeline.

#### Scenario: Enabled monitor is selected for execution
- **WHEN** execution pipeline evaluates runnable monitors
- **THEN** it selects monitors whose lifecycle state is enabled

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

