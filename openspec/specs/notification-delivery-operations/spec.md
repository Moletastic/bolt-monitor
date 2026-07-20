## ADDED Requirements

### Requirement: API exposes incident-scoped delivery outcomes
The system SHALL provide an incident-scoped API operation that lists notification deliveries for an existing tenant-owned incident in stable chronological order. Each result SHALL include delivery identity, transition identity, policy step, channel ID and type, exactly one state from `pending`, `in_flight`, `retryable_failed`, `ambiguous`, `delivered`, or `terminal_failed`, attempt count, timestamps, normalized outcome classification, and sanitized provider metadata, and SHALL use the standard response envelope. Recovery suppression SHALL be returned separately as escalation/replay eligibility, not as a delivery state.

#### Scenario: Operator lists incident deliveries
- **WHEN** an operator requests deliveries for an existing incident
- **THEN** the API returns only deliveries belonging to that incident and tenant
- **AND** orders them by creation time with delivery identity as a stable tie-breaker

#### Scenario: Incident has no deliveries
- **WHEN** an existing incident has no delivery records
- **THEN** the API returns a successful response with an empty deliveries collection

#### Scenario: Incident does not exist
- **WHEN** an operator requests deliveries for an unknown incident
- **THEN** the API returns the typed incident-not-found error

#### Scenario: Delivery metadata is returned
- **WHEN** a delivery contains provider outcome metadata
- **THEN** the API returns only allowlisted sanitized fields
- **AND** does not return credentials, raw channel config, authorization headers, raw provider bodies, or secret-bearing URLs

### Requirement: Operators can replay eligible failed deliveries safely
The system SHALL provide an incident-scoped replay operation for a specific delivery in `terminal_failed` state. The operation SHALL require an `Idempotency-Key`, preserve the original delivery identity, increment replay audit metadata once, reset it to `pending` through a conditional transaction, and create one canonical `delivery_replay` dispatch record consumed by the same DynamoDB Stream-to-notification-SQS dispatcher. `delivered`, `pending`, `in_flight`, `retryable_failed`, `ambiguous`, unknown, cross-tenant, and recovery-ineligible deliveries SHALL NOT be replayed.

#### Scenario: Operator replays terminal failed delivery
- **WHEN** an operator requests replay of a tenant-owned delivery in `terminal_failed` state with an unused `Idempotency-Key` and the incident remains eligible for that delivery
- **THEN** the system conditionally returns the same delivery identity to `pending`
- **AND** creates retry work for the canonical Stream dispatcher
- **AND** records the replay action without resending any delivered sibling channel

#### Scenario: Replay enqueue fails
- **WHEN** the replay transaction commits but Stream dispatch or queue enqueue fails
- **THEN** durable pending dispatch work remains available for retry
- **AND** the API returns a typed failure rather than reporting successful enqueue

#### Scenario: Operator requests replay of delivered delivery
- **WHEN** an operator requests replay of a `delivered` delivery
- **THEN** the API rejects the request with a typed conflict error
- **AND** no provider request is made

#### Scenario: Incident recovered before replay
- **WHEN** an operator requests replay for escalation delivery after the incident resolved or escalation was suppressed
- **THEN** the API rejects the replay as ineligible
- **AND** no notification work is enqueued

### Requirement: Delivery replay requests are idempotent for bounded retention
Replay idempotency SHALL be scoped to tenant, incident, delivery, operation, and `Idempotency-Key`. The system SHALL persist a canonical request fingerprint and replay result identity for a named bounded retention duration longer than the maximum dispatch and retry window. Repeating the same key and request during retention SHALL return the original result without creating another replay. Reusing the key with a different request fingerprint SHALL return a typed conflict.

#### Scenario: Client retries the same replay request
- **WHEN** the same `Idempotency-Key`, path identity, and request payload are submitted during retention
- **THEN** the API returns the original replay result
- **AND** replay count, state reset, audit record, and dispatch record occur only once

#### Scenario: Client changes payload under the same key
- **WHEN** an `Idempotency-Key` is reused with a different canonical request fingerprint during retention
- **THEN** the API returns a typed idempotency conflict
- **AND** no delivery or dispatch state changes

#### Scenario: Concurrent requests use the same key
- **WHEN** equivalent replay requests race with the same `Idempotency-Key`
- **THEN** the conditional transaction creates one replay
- **AND** both requests converge on the same replay result

#### Scenario: Idempotency record expires
- **WHEN** the bounded retention period has elapsed
- **THEN** the record may expire
- **AND** a later request is evaluated as new and succeeds only if the delivery is again replay-eligible

### Requirement: Dashboard shows delivery state and replay controls in incident context
The incident detail dashboard SHALL show all six exact per-step and per-channel delivery states in the escalation area, including provider-acceptance wording, attempt count, timestamps, sanitized failure classification, ambiguous-outcome risk, and visible pending, empty, and error states. It SHALL show recovery suppression separately, offer replay only for API-eligible `terminal_failed` deliveries, and use a server action rather than imperative router mutation.

#### Scenario: Multi-channel step partially succeeds
- **WHEN** one channel is `delivered` and another is `terminal_failed`
- **THEN** the dashboard shows each channel outcome independently
- **AND** labels the successful channel as accepted by the provider

#### Scenario: Replay is available
- **WHEN** a `terminal_failed` delivery is eligible for replay
- **THEN** the dashboard presents a replay action with visible pending feedback
- **AND** reports the typed result inline without claiming human receipt

#### Scenario: Delivery is not replayable
- **WHEN** a delivery is `delivered`, nonterminal, or ineligible because escalation is suppressed
- **THEN** the dashboard does not present an enabled replay action for it

### Requirement: Operations runbook covers diagnosis, replay, and legacy cleanup
The repository SHALL contain an operations runbook for notification delivery that documents safe delivery-state inspection, CloudWatch log and metric correlation, source-kind-aware notification DLQ inspection, bounded recent-bucket and point-identity outbox reconciliation, manual repair, idempotent terminal-failure replay, recovery checks, and legacy EventBridge rule and Lambda permission cleanup. Commands SHALL be tenant- and resource-scoped, SHALL prohibit blind cross-source redrive and unbounded scans, and SHALL avoid printing secrets or raw provider payloads.

#### Scenario: Poison message reaches DLQ
- **WHEN** an operator follows the runbook for a notification DLQ message
- **THEN** the procedure identifies the message type and safe identity fields
- **AND** explains source-specific conditions for reconciliation, eligibility recheck, validated SQS redrive, API replay, or quarantine

#### Scenario: Operator investigates ambiguous timeout
- **WHEN** a delivery failed after timeout retries
- **THEN** the runbook explains that provider acceptance may be ambiguous
- **AND** instructs the operator to check safe provider identifiers before replay where available

#### Scenario: Operator cleans legacy schedules
- **WHEN** an operator follows the legacy cleanup section
- **THEN** the procedure inventories resources before deletion
- **AND** limits deletion to the documented legacy naming and permission patterns

### Requirement: Delivery assurance has automated regression coverage
The change SHALL include tests for canonical outbox dependency and sole dispatch ownership, Stream exhaustion and bounded reconciliation, deterministic identities, all six delivery states and claims, retry/ambiguity classification, timeout/lease/visibility/retry ordering, partial multi-channel success, source-correct partial batches and DLQ redrive restrictions, recovery suppression, one-time schedule configuration and cleanup, least-privilege infrastructure, API visibility and idempotent replay retention/conflict, secret redaction, dashboard states, OpenAPI, and Bruno route coverage.

#### Scenario: Delivery assurance test suite runs
- **WHEN** repository Go, dashboard, infrastructure, Bruno, and contract checks run
- **THEN** the suite verifies the delivery, scheduling, API, and operator behavior required by this change
