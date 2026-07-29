## ADDED Requirements

### Requirement: Canonical transition queue payload uses event identity
The Stream dispatcher SHALL enqueue the outbox `eventId` as `transitionId`. It SHALL retain a dispatch-pending canonical outbox record until the queue consumer completes transition handling, after which the consumer SHALL conditionally acknowledge that same `eventId`.

#### Scenario: Stream dispatch reaches queue consumer
- **WHEN** a pending canonical transition outbox record is inserted
- **THEN** its SQS envelope carries the outbox event identity as `transitionId`
- **AND** the consumer loads and acknowledges that same pending outbox record only after handling succeeds

#### Scenario: Queue handling fails
- **WHEN** transition handling fails after Stream delivery reaches SQS
- **THEN** the outbox record remains pending
- **AND** SQS retries the canonical transition without a new outbox record
