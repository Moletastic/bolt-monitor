## ADDED Requirements

### Requirement: System exposes incident activity/timeline through HTTP API

System SHALL allow clients to read activity/timeline events for an individual incident through HTTP API. This surfaces the full state-transition history of an incident  not just its current status.

#### Scenario: Operator requests incident activity timeline
- **WHEN** operator calls `GET /api/v1/incidents/{incidentId}/activities` for an existing incident
- **THEN** system returns all activity records associated with that incident, sorted by timestamp ascending

#### Scenario: Incident has no recorded activity
- **WHEN** operator requests activity for an existing incident with no state transitions recorded beyond creation
- **THEN** system returns an empty activities collection with successful response

#### Scenario: Requested incident does not exist
- **WHEN** operator calls `GET /api/v1/incidents/{incidentId}/activities` for a non-existent incident
- **THEN** system returns a 404 response with an error message

### Requirement: Incident activity responses expose state-transition metadata

System SHALL return activity records containing the action/event type and timestamp for each state transition.

#### Scenario: Activity records are returned for an incident
- **WHEN** system returns activity for an incident
- **THEN** each activity record includes the action/event type and the timestamp at which the transition occurred

#### Scenario: Activity records reflect chronological order
- **WHEN** system returns activity for an incident
- **THEN** records are sorted by timestamp in ascending order (oldest first), enabling a timeline view
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
