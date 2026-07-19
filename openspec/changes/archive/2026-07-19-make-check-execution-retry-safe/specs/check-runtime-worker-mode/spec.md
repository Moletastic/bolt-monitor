## MODIFIED Requirements

### Requirement: Worker receives execution request from SQS
System SHALL parse the SQS body as an immutable execution envelope and resolve canonical work from DynamoDB before processing.

#### Scenario: Message parsing
- **WHEN** worker receives a valid SQS execution message
- **THEN** it extracts tenant, service, monitor, `runId`, trigger, accepted time, and recurring `scheduleDefinitionVersion`/UTC `scheduledFor` when applicable
- **AND** it rejects a mismatch between the envelope and persisted immutable work identity as a typed conflict

#### Scenario: Message is malformed
- **WHEN** worker cannot parse or validate required identity fields
- **THEN** it returns a typed malformed-envelope failure
- **AND** the existing SQS retry and DLQ policy applies

### Requirement: Worker executes HTTP check against target
System SHALL execute an HTTP check only after obtaining the current work lease and reloading runnable monitor state from DynamoDB.

#### Scenario: HTTP execution
- **WHEN** worker holds the current lease and the monitor remains eligible
- **THEN** it calls `checkexecution.ExecuteHTTP()` using the current persisted HTTP target, method, headers, expectations, and timeout
- **AND** identity and trigger are taken from canonical work

#### Scenario: Queued monitor snapshot differs
- **WHEN** mutable monitor configuration in a message differs from current persisted configuration
- **THEN** worker does not execute the queued snapshot
- **AND** it either executes current eligible configuration or records a typed superseded-work skip according to recurring eligibility

### Requirement: Worker records result to DynamoDB
System SHALL commit execution result through the shared conditional result transaction after an HTTP attempt completes.

#### Scenario: Result recording
- **WHEN** HTTP check completes with any normalized outcome
- **THEN** worker submits monitor configuration, stable work identity, fencing token, trigger, `scheduleDefinitionVersion`/`scheduledFor` when recurring, and normalized result
- **AND** result identity, including recurring schedule identity, MUST match canonical work identity
- **AND** at most one canonical `CheckRun` is accepted

#### Scenario: Duplicate result arrives
- **WHEN** work is already terminal with a canonical result for the same `runId`
- **THEN** worker treats result recording as an idempotent duplicate
- **AND** it does not rewrite run, status, incident, activity, or outbox state

### Requirement: Worker handles result recording failures
System SHALL acknowledge only terminal/no-op outcomes and SHALL expose retryable result persistence failures to SQS.

#### Scenario: DynamoDB write failure
- **WHEN** worker completes an HTTP check but the result transaction fails with a retryable storage error
- **THEN** it returns a typed retryable result-commit failure
- **AND** the SQS message is not acknowledged so work can be reclaimed after lease expiry

#### Scenario: Worker loses lease
- **WHEN** result commit finds a different current fencing token
- **THEN** the stale worker records no canonical result or projections
- **AND** stops processing with a typed lease-lost outcome

### Requirement: Worker deletes SQS message on successful processing
System SHALL report per-record SQS success after work reaches a terminal state and any required canonical notification outbox item is durably committed; notification dispatch is not an execution-message acknowledgement prerequisite.

#### Scenario: Successful processing
- **WHEN** worker commits the canonical result, terminal marker removal, and any required transition outbox item atomically
- **THEN** it returns per-message success so the event source acknowledges the message

#### Scenario: Terminal duplicate processing
- **WHEN** worker receives a duplicate for completed or skipped work
- **THEN** it returns success without another HTTP request

## MODIFIED Requirements

### Requirement: Worker claims work with a recoverable lease
System SHALL use conditional claims, expiring leases, and fencing tokens rather than relying on SQS visibility as the execution claim.

#### Scenario: Pending work is claimed
- **WHEN** worker conditionally claims pending work
- **THEN** work becomes `in_progress` with a fresh fencing token, lease expiry, start time, and incremented attempt count

#### Scenario: Active work is delivered again
- **WHEN** duplicate delivery occurs before the current lease expires
- **THEN** the second worker performs no HTTP request and cannot complete the work

#### Scenario: Lease becomes stale
- **WHEN** `in_progress` work has an expired lease
- **THEN** a later worker may reclaim it with a new fencing token
- **AND** the old token can no longer complete or skip the work

#### Scenario: Recovery searches for expired work
- **WHEN** bounded recovery searches configured due buckets for nonterminal work
- **THEN** it queries directly addressable work-recovery markers rather than scanning tenant work
- **AND** validates canonical state and fencing data before republishing or removing a stale marker

### Requirement: Worker revalidates current execution eligibility
System SHALL strongly re-read monitor and status state after claim and before the HTTP side effect.

#### Scenario: Monitor no longer exists
- **WHEN** claimed work references a deleted monitor
- **THEN** worker conditionally marks work skipped with a typed not-found reason
- **AND** creates no `CheckRun`, status change, incident/activity, or outbox item

#### Scenario: Monitor is disabled or in maintenance
- **WHEN** current monitor is disabled or current monitor state is maintenance
- **THEN** worker conditionally marks work skipped with the corresponding typed reason
- **AND** performs no HTTP request

#### Scenario: Current configuration is invalid or recurring work is superseded
- **WHEN** current execution configuration is invalid, `scheduleDefinitionVersion` changed, or accepted `scheduledFor` is no longer eligible
- **THEN** worker conditionally records a typed terminal skip
- **AND** does not silently execute stale configuration

### Requirement: Worker failures have typed retry policy
System SHALL classify worker outcomes as retryable failure, terminal skip, idempotent duplicate, claim conflict, lease loss, or successful completion using typed errors with safe run context.

#### Scenario: Expected conditional conflict occurs
- **WHEN** DynamoDB rejects a claim or transition condition
- **THEN** worker decodes the structured cancellation reason
- **AND** does not depend on matching AWS error message strings

### Requirement: Worker reports partial SQS batch failures
System SHALL enable SQS partial batch responses and report only malformed or retryable records as failed, while successful, terminal-skip, and idempotent-duplicate records complete independently.

#### Scenario: Batch has success and retryable failure
- **WHEN** one execution record reaches terminal success and another has retryable result persistence failure
- **THEN** the response contains only the retryable record's SQS item identifier
- **AND** the successful record is not redelivered because of the failed record
