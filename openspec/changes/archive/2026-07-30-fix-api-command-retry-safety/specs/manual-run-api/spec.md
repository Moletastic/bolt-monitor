## MODIFIED Requirements

### Requirement: Manual request and result retries are idempotent
System SHALL canonicalize the service-scoped command into a deterministic request fingerprint and retain a bounded idempotency record so repeated use of the same key and request converges on one manual `runId`, terminal work record, canonical `CheckRun`, and canonical public response.

#### Scenario: Same request is replayed
- **WHEN** the same scoped `Idempotency-Key` and request fingerprint are received within `MANUAL_IDEMPOTENCY_RETENTION`
- **THEN** system returns the same sanitized in-progress or completed public result
- **AND** does not execute another completed run or create recurring projection effects

#### Scenario: Same key is reused for a different request
- **WHEN** the same scoped `Idempotency-Key` is received with a different canonical request fingerprint
- **THEN** system returns `IDEMPOTENCY_CONFLICT` in the standard error envelope
- **AND** does not mutate or execute either request

#### Scenario: Idempotency retention is configured
- **WHEN** a manual idempotency record is created
- **THEN** it stores fingerprint, `runId`, sanitized canonical result or pending state, and TTL only
- **AND** `MANUAL_IDEMPOTENCY_RETENTION` is bounded, documented, and tested
