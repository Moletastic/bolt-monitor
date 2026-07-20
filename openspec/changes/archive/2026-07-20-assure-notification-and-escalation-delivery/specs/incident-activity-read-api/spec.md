## ADDED Requirements

### Requirement: Incident transitions provide stable delivery correlation identity
Notification-relevant incident activity records SHALL retain the stable deterministic identity supplied by the retry-safe canonical transition/outbox contract. Its `activityId` SHALL equal canonical `eventId` and SHALL be used as `transitionId` for dispatch, delivery, schedule, and operator correlation. This change SHALL consume that identity without creating another transition or activity record. Existing activity ordering and response fields SHALL remain compatible.

#### Scenario: Incident down transition starts delivery work
- **WHEN** the system records the incident activity that represents the down transition
- **THEN** its `activityId` is reused as the transition identity in associated delivery work
- **AND** the Stream dispatcher consumes the matching retry-safe canonical outbox record

#### Scenario: Operator correlates delivery with activity
- **WHEN** an operator compares an incident activity response with an incident delivery response
- **THEN** the delivery's `transitionId` matches the corresponding activity's `activityId`

#### Scenario: Existing incident activity is listed
- **WHEN** a client reads incident activities after this change
- **THEN** records remain sorted by timestamp ascending
- **AND** retain `activityId`, `incidentId`, `action`, and `timestamp`
