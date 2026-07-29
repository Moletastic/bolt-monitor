## ADDED Requirements

### Requirement: Manual idempotency expiry is persisted as DynamoDB TTL
Each manual-run idempotency record SHALL persist an absolute Unix-epoch expiry in the configured DynamoDB TTL attribute, derived from `MANUAL_IDEMPOTENCY_RETENTION`. The system SHALL not rely on physical TTL deletion occurring at the exact expiry time.

#### Scenario: Manual idempotency record is created
- **WHEN** a new manual-run idempotency key is reserved
- **THEN** its persisted TTL is the record expiry as Unix epoch seconds
- **AND** it bounds storage retention without a cleanup worker or scan

#### Scenario: TTL cleanup lags
- **WHEN** DynamoDB has not yet physically removed an expired idempotency record
- **THEN** the system does not claim that TTL deletion is immediate
- **AND** expiry behavior remains safe for the configured replay contract
