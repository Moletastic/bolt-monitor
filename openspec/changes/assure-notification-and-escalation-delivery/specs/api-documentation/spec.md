## ADDED Requirements

### Requirement: API contract documents incident notification delivery operations
The source-controlled OpenAPI document SHALL describe the incident-scoped notification delivery list and replay operations, their standard response envelopes, the exact `pending`, `in_flight`, `retryable_failed`, `ambiguous`, `delivered`, and `terminal_failed` delivery enum, separate recovery-suppression eligibility, sanitized metadata shape, replay eligibility errors, required `Idempotency-Key`, bounded idempotency retention, mismatch conflict, and provider-acceptance semantics.

#### Scenario: Developer reviews delivery list contract
- **WHEN** a developer reads the OpenAPI path for incident notification deliveries
- **THEN** the contract documents delivery identity, transition identity, policy step, channel reference and type, state, attempts, timestamps, safe outcome fields, and empty and not-found responses

#### Scenario: Developer reviews replay contract
- **WHEN** a developer reads the OpenAPI path for delivery replay
- **THEN** the contract documents required `Idempotency-Key`, one-result same-key/same-request behavior, payload-mismatch conflict, retention semantics, and typed errors for unknown, cross-tenant, non-`terminal_failed`, or recovery-ineligible deliveries

#### Scenario: Developer reviews lifecycle contract
- **WHEN** a developer reads the delivery schema
- **THEN** it defines all six delivery states and their meanings
- **AND** does not include recovery suppression in the delivery-state enum

#### Scenario: API examples describe delivery success
- **WHEN** the OpenAPI examples show a `delivered` outcome
- **THEN** they describe provider acceptance without claiming human receipt
- **AND** contain no credentials, raw provider bodies, authorization headers, or secret-bearing URLs
