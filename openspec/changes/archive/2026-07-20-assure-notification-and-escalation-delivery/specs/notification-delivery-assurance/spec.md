## ADDED Requirements

### Requirement: Incident transition dispatch is durable
The system SHALL consume the single canonical incident transition event/outbox record created atomically with the recurring result and incident transition by the corrected `make-check-execution-retry-safe` contract. This change SHALL NOT create a second transition event/outbox or use a direct post-commit transition publisher. Its sole transition dispatch path SHALL consume canonical record inserts from DynamoDB Stream, enqueue the stable identity to the existing notification queue, and conditionally acknowledge dispatch on that same canonical record.

#### Scenario: Notification queue enqueue fails
- **WHEN** a canonical transition record is committed but its Stream dispatcher cannot confirm notification queue acceptance
- **THEN** the canonical record remains dispatch-pending
- **AND** the failure is returned or recorded for retry rather than discarded
- **AND** a later dispatcher attempt sends the same transition identity without recreating the incident transition

#### Scenario: Dispatch retry follows an ambiguous enqueue result
- **WHEN** a dispatcher retries a transition because queue acceptance was not confirmed
- **THEN** the message contains the same transition identity as the prior attempt
- **AND** downstream delivery processing remains idempotent if the queue accepted both attempts

#### Scenario: Queue accepts transition work
- **WHEN** the notification queue accepts the transition message
- **THEN** the system conditionally records the canonical item as dispatch-acknowledged without deleting the incident transition history

#### Scenario: Competing publisher is considered
- **WHEN** transition dispatch is implemented
- **THEN** only the DynamoDB Stream dispatcher sends canonical transition records to the notification queue
- **AND** execution-result handling does not directly publish or create another transition dispatch record

### Requirement: Stream exhaustion remains durably and boundedly recoverable
The system SHALL retain every unacknowledged canonical dispatch record in a sparse tenant/time-bucketed pending access path after DynamoDB Stream retries exhaust. Automatic reconciliation SHALL read a configured maximum number of recent buckets and pages, and manual repair SHALL address one canonical `eventId`. Neither path SHALL perform an unbounded table scan or silently mark an unconfirmed dispatch successful.

#### Scenario: Stream retries exhaust
- **WHEN** DynamoDB Stream exhausts its bounded retries for a canonical record
- **THEN** the source-tagged Stream failure reaches the notification DLQ
- **AND** the canonical record remains pending and addressable by bucket and `eventId`

#### Scenario: Reconciler repairs pending dispatch
- **WHEN** bounded reconciliation finds a pending canonical identity
- **THEN** it dispatches the same immutable payload and conditionally acknowledges the same record
- **AND** dispatcher/reconciler races cannot create a second logical transition

#### Scenario: Operator repairs one transition
- **WHEN** an operator supplies a pending canonical `eventId`
- **THEN** the repair path uses a point lookup and the same conditional dispatch protocol
- **AND** does not scan unrelated records

### Requirement: Delivery identity is durable per transition, policy step, and channel
The system SHALL derive one deterministic delivery identity from the tenant-scoped incident transition identity, selected policy step, and channel ID. It SHALL persist that delivery in the existing DynamoDB table before calling a provider and SHALL enforce uniqueness so duplicate queue or schedule invocations do not create independent deliveries.

#### Scenario: Duplicate transition message is processed
- **WHEN** the notification queue delivers the same incident transition more than once
- **THEN** the system resolves the same delivery identity for each channel in the policy step
- **AND** it does not create duplicate delivery records

#### Scenario: Same channel appears in different steps
- **WHEN** one channel is selected by two different policy steps for the same incident transition
- **THEN** each step has a distinct delivery identity

#### Scenario: Tenant scope is enforced
- **WHEN** a delivery record is read or updated
- **THEN** its tenant and incident transition ownership are validated
- **AND** an identity from another tenant cannot access or mutate the delivery

### Requirement: Delivery lifecycle and provider metadata are persisted safely
Each delivery SHALL have exactly one of `pending`, `in_flight`, `retryable_failed`, `ambiguous`, `delivered`, or `terminal_failed` as its current state. `pending` means durable unclaimed work; `in_flight` means one fenced attempt lease is active; `retryable_failed` means a confirmed transient failure awaits bounded retry; `ambiguous` means the request may have reached the provider but acceptance is unknown; `delivered` means terminal provider acceptance; and `terminal_failed` means terminal rejection/configuration failure or exhausted retry budget. Recovery suppression SHALL be escalation eligibility/state and SHALL NOT be a delivery state. The system SHALL record attempt count and timestamps, a sanitized outcome classification, and only allowlisted provider metadata such as HTTP status class, provider request identifier, or retry-after time. It MUST NOT persist credentials, authorization headers, channel configuration secrets, raw request payloads, raw provider response bodies, or secret-bearing URLs.

#### Scenario: Delivery is created before provider I/O
- **WHEN** a policy step resolves a channel that has not been attempted
- **THEN** the system persists the delivery as `pending` before provider I/O

#### Scenario: Provider call starts
- **WHEN** the worker is about to call the channel provider
- **THEN** the system atomically moves an eligible delivery to `in_flight` with a fencing token and lease
- **AND** increments its attempt count and records the attempt timestamp

#### Scenario: Provider accepts the notification
- **WHEN** the provider returns its documented acceptance response
- **THEN** the system marks the delivery `delivered`
- **AND** records only sanitized allowlisted provider metadata

#### Scenario: Unsafe provider response is received
- **WHEN** a provider failure response contains a body, headers, URL, or values that may include secrets or personal data
- **THEN** the persisted delivery and operator-facing response omit those raw values
- **AND** retain only the normalized failure classification and safe metadata

#### Scenario: Transient failure is confirmed
- **WHEN** a provider returns a confirmed retryable response before the retry budget is exhausted
- **THEN** the system moves the delivery to `retryable_failed` with bounded retry eligibility

#### Scenario: Acceptance cannot be determined
- **WHEN** a timeout, post-send transport loss, or abandoned in-flight lease cannot establish provider acceptance
- **THEN** the system records or recovers the delivery as `ambiguous`
- **AND** exposes duplicate-side-effect risk separately from a confirmed retryable failure

#### Scenario: Recovery suppresses escalation
- **WHEN** the incident resolves before an unfinished delivery is attempted
- **THEN** escalation eligibility becomes suppressed
- **AND** the delivery state enum is not extended with a suppressed value

### Requirement: Automatic retry follows normalized failure classification
The system SHALL classify network timeouts, transport failures, HTTP `429`, and HTTP `5xx` as retryable. It SHALL classify invalid or missing channel configuration, unsupported channel type, and terminal provider HTTP `4xx` other than `429` as non-retryable. Confirmed transient failures SHALL become `retryable_failed`, uncertain acceptance SHALL become `ambiguous`, and both SHALL use the existing notification queue retry and redrive behavior while budget remains. Non-retryable failures SHALL become `terminal_failed` and be acknowledged without consuming further automatic retries.

#### Scenario: Provider times out
- **WHEN** a delivery attempt times out before acceptance is confirmed
- **THEN** the delivery becomes `ambiguous` with a retryable timeout classification while budget remains
- **AND** the queue message is not acknowledged as successfully processed

#### Scenario: Provider throttles delivery
- **WHEN** a provider returns HTTP `429`
- **THEN** the system classifies the attempt as `retryable_failed`
- **AND** records a sanitized retry-after value when the provider supplies a valid one

#### Scenario: Provider is unavailable
- **WHEN** a provider returns HTTP `5xx`
- **THEN** the system classifies the attempt as `retryable_failed`

#### Scenario: Channel configuration is terminally invalid
- **WHEN** a channel is missing, unsupported, or cannot pass sender configuration validation
- **THEN** the system marks its delivery `terminal_failed` with a terminal configuration classification
- **AND** it does not retry that delivery automatically

#### Scenario: Provider rejects a valid request with terminal client status
- **WHEN** a provider returns HTTP `4xx` other than `429`
- **THEN** the system marks the delivery `terminal_failed` with a terminal provider-rejection classification
- **AND** it does not retry that delivery automatically

#### Scenario: Retry budget is exhausted
- **WHEN** a retryable message reaches the notification queue maximum receive count without successful processing
- **THEN** the system records known unfinished deliveries as `terminal_failed` with a retry-exhausted classification while preserving the last outcome class
- **AND** returns failure so SQS moves the message to the existing notification DLQ

### Requirement: Retries do not resend delivered channels
Before every provider call, the system SHALL conditionally claim `pending`, retry-eligible `retryable_failed`, retry-eligible `ambiguous`, or lease-expired `in_flight` delivery identity. A delivery already marked `delivered` SHALL be treated as complete and SHALL NOT be sent again. Provider-supported idempotency or deduplication keys SHALL use the stable delivery identity where the provider contract permits.

#### Scenario: One channel succeeds and another channel times out
- **WHEN** a multi-channel step delivers channel A and receives a retryable failure for channel B
- **THEN** channel A remains `delivered`
- **AND** retry processing skips channel A and attempts only channel B

#### Scenario: Concurrent workers receive duplicate work
- **WHEN** two workers concurrently process the same delivery identity
- **THEN** a conditional write allows at most one worker to claim the next provider attempt
- **AND** the other worker observes the current durable state without sending

#### Scenario: Provider supports an idempotency key
- **WHEN** a sender's provider contract supports request idempotency or event deduplication
- **THEN** the sender uses the stable delivery identity as that key

### Requirement: Poison messages are retained in the notification DLQ
The escalation runtime SHALL fail malformed, unsupported, or repeatedly unprocessable notification queue records rather than logging and acknowledging them. It SHALL return source-correct partial batch identifiers: DynamoDB sequence numbers for Stream records and SQS message IDs for queue records. Partial batch behavior SHALL preserve failed records and allow successfully processed records to complete independently.

#### Scenario: Notification message is malformed
- **WHEN** a queue record cannot be parsed or lacks required identity fields
- **THEN** the escalation runtime reports that record as failed
- **AND** repeated failures cause the existing notification queue redrive policy to move it to the notification DLQ

#### Scenario: Batch contains successful and poison records
- **WHEN** one record in a queue batch succeeds and another is poison
- **THEN** the runtime reports only the poison record as failed
- **AND** SQS does not redeliver the successful record because of the poison record

#### Scenario: Stream batch contains successful and failed records
- **WHEN** one canonical Stream record is dispatched and another cannot be dispatched
- **THEN** the runtime reports only the failed record's sequence number
- **AND** acknowledgement updates do not trigger a dispatch loop

### Requirement: Notification DLQ redrive is source-kind aware
Every notification-DLQ envelope SHALL identify `dynamodb_stream`, `notification_sqs`, or `scheduler_target` source kind and safe canonical identity when parseable. Stream envelopes SHALL be repaired through canonical outbox reconciliation and SHALL NOT be sent directly to the notification queue. Scheduler envelopes SHALL be revalidated against current scheduled-step and incident eligibility before enqueue. Only validated canonical SQS envelopes SHALL be eligible for queue redrive; malformed or unknown envelopes SHALL be quarantined.

#### Scenario: Operator encounters Stream failure envelope
- **WHEN** a `dynamodb_stream` envelope is inspected in the notification DLQ
- **THEN** the procedure reconciles its canonical `eventId`
- **AND** never redrives the Stream event shape to SQS

#### Scenario: Operator encounters Scheduler failure envelope
- **WHEN** a `scheduler_target` envelope is considered for redrive
- **THEN** the scheduled-step identity and current incident/escalation eligibility are revalidated first

#### Scenario: Envelope kind is unknown
- **WHEN** a DLQ envelope has malformed or unsupported source metadata
- **THEN** it is quarantined rather than blindly redriven

### Requirement: Delivery timing and retry configuration is ordered
Named configuration SHALL satisfy `ProviderRequestTimeout + ProviderCompletionBuffer < NotificationLambdaTimeout`, `NotificationLambdaTimeout + LambdaTerminationBuffer < DeliveryAttemptLease`, and `ClaimStartBudget + DeliveryAttemptLease + RedeliveryBuffer <= NotificationQueueVisibilityTimeout`. `ClaimStartBudget` SHALL bound work before the attempt claim. `DeliveryAutomaticAttemptLimit` SHALL NOT exceed `NotificationQueueMaxReceiveCount`, bounded retry-after/backoff SHALL be shorter than the attempt lease, and the final allowed receive SHALL terminalize known unfinished deliveries before DLQ redrive. EventBridge Scheduler target retry attempts and age SHALL be finite, SHALL cover its configured maximum target backoff, and SHALL route exhaustion to the source-tagged notification DLQ without consuming a provider attempt. Infrastructure tests SHALL assert these relations.

#### Scenario: Provider request reaches its timeout
- **WHEN** the provider timeout elapses
- **THEN** Lambda retains enough time to persist `ambiguous` before its own timeout

#### Scenario: Lambda terminates during an attempt
- **WHEN** the notification Lambda ends without persisting an outcome
- **THEN** the attempt lease remains active through termination
- **AND** expires before the SQS message is visible for another provider attempt

#### Scenario: Scheduler cannot enqueue a delayed step
- **WHEN** Scheduler exhausts its finite target retry age or attempts
- **THEN** the source-tagged target envelope reaches the notification DLQ
- **AND** no delivery attempt count is consumed

### Requirement: Incident recovery suppresses future escalation delivery
Before creating or claiming any delivery for a scheduled step, the system SHALL verify both the current incident state and escalation state. A resolved incident or suppressed escalation SHALL prevent all not-yet-delivered future steps from calling providers, even when the recovery transition message is delayed or a schedule invocation is duplicated.

#### Scenario: Incident resolves before scheduled step
- **WHEN** a scheduled step reaches the queue after the incident is resolved
- **THEN** the system does not call any provider for that step
- **AND** marks the escalation state suppressed if it is not already terminal

#### Scenario: Recovery event is delayed behind scheduled work
- **WHEN** the incident record is resolved but its recovery notification event has not yet been processed
- **THEN** the scheduled worker suppresses the step based on the incident record

#### Scenario: Delivery completed before recovery
- **WHEN** a channel delivery reached `delivered` before the incident recovered
- **THEN** the system preserves that historical outcome
- **AND** suppresses only future escalation delivery

### Requirement: Delivery means provider acceptance, not human receipt
The system SHALL define `delivered` as acceptance by the configured provider according to that provider's API response. It SHALL NOT represent or promise that a human received, read, or acted on the notification.

#### Scenario: Provider accepts but recipient does not read
- **WHEN** the provider accepts a notification and no human-read signal exists
- **THEN** the delivery remains `delivered`
- **AND** operator-facing text describes provider acceptance rather than human receipt
