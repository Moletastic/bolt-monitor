## ADDED Requirements

### Requirement: Selected operator commands are retry-safe
The system SHALL require bounded idempotency for notification test send, incident acknowledgement, and incident resolution commands.

#### Scenario: Same command is retried
- **WHEN** an operator repeats a command with the same scoped `Idempotency-Key` and canonical request
- **THEN** the system returns the original sanitized result
- **AND** performs its durable mutation and external side effect at most once

#### Scenario: Key is reused for another command
- **WHEN** an operator reuses a scoped key with a different canonical request
- **THEN** the system returns `IDEMPOTENCY_CONFLICT` through the standard error envelope
- **AND** performs no new side effect

#### Scenario: Initial command outcome is not durably completed
- **WHEN** a command record remains `pending` because processing stopped after reservation
- **THEN** a same-key retry returns its original sanitized pending result
- **AND** does not resume an external notification or incident mutation

### Requirement: Idempotency records are bounded
The system SHALL store only `tenantId`, operation, resource identity, sanitized command result data, fingerprint, state, optional run identity, and absolute DynamoDB TTL for command idempotency.

#### Scenario: Record expires
- **WHEN** the configured retention has elapsed
- **THEN** the system does not depend on physical TTL deletion timing
- **AND** evaluates a later request as new only when command state permits it
