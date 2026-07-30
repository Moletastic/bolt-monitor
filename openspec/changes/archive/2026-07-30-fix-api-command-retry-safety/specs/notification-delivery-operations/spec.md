## MODIFIED Requirements

### Requirement: Delivery replay requests are idempotent for bounded retention
Replay idempotency SHALL be scoped to tenant, incident, delivery, operation, and `Idempotency-Key`, read case-insensitively from the HTTP request. The system SHALL persist a canonical request fingerprint and replay result identity for a named bounded retention duration longer than the maximum dispatch and retry window. Repeating the same key and request during retention SHALL return the original result without creating another replay. Reusing the key with a different request fingerprint SHALL return a typed conflict.

#### Scenario: Client retries the same replay request
- **WHEN** the same `Idempotency-Key`, regardless of HTTP header casing, path identity, and request payload are submitted during retention
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
