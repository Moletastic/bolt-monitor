## ADDED Requirements

### Requirement: Test sends are idempotent
The notification channel test-send command SHALL require `Idempotency-Key` and retain one bounded sanitized result per channel and canonical request.

#### Scenario: Test send is retried
- **WHEN** an operator repeats a test-send request with the same key and request
- **THEN** the system returns the original result
- **AND** sends at most one provider notification

#### Scenario: Provider outcome is uncertain
- **WHEN** provider delivery returns an ambiguous error after command reservation
- **THEN** the system records a sanitized pending result
- **AND** a retry with the same key does not send another provider notification
