## MODIFIED Requirements

### Requirement: API exposes incident-scoped delivery outcomes
The system SHALL provide an incident-scoped API operation that lists notification deliveries for an existing tenant-owned incident in stable chronological order through bounded cursor pages. Each result SHALL include delivery identity, transition identity, policy step, channel ID and type, exactly one state from `pending`, `in_flight`, `retryable_failed`, `ambiguous`, `delivered`, or `terminal_failed`, attempt count, timestamps, normalized outcome classification, and sanitized provider metadata, and SHALL use the standard response envelope. Recovery suppression SHALL be returned separately as escalation/replay eligibility, not as a delivery state.

#### Scenario: Operator lists first delivery page
- **WHEN** an operator requests deliveries for an existing incident without a cursor
- **THEN** the API returns at most the validated limit in creation-time and delivery-ID order
- **AND** includes an opaque continuation cursor only when more matching deliveries exist

#### Scenario: Operator follows delivery cursor
- **WHEN** an operator supplies a continuation cursor for the same incident delivery collection
- **THEN** the API returns the next records without duplicating prior records

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
