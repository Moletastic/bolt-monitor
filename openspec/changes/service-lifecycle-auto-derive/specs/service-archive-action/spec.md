## ADDED Requirements

### Requirement: System allows archiving services via explicit action
System SHALL provide an explicit endpoint to transition a service to archived lifecycle state.

#### Scenario: Client archives active service
- **WHEN** client sends `POST /api/v1/services/{serviceId}/archive` for an active service
- **THEN** system SHALL transition service lifecycle to `archived`
- **AND** service retains all monitor configurations and enabled/disabled states

#### Scenario: Client archives draft service
- **WHEN** client sends `POST /api/v1/services/{serviceId}/archive` for a draft service
- **THEN** system SHALL transition service lifecycle to `archived`

#### Scenario: Client archives already archived service
- **WHEN** client sends `POST /api/v1/services/{serviceId}/archive` for an already archived service
- **THEN** system SHALL return 200 OK with current archived service (idempotent)

#### Scenario: Client archives missing service
- **WHEN** client sends `POST /api/v1/services/{serviceId}/archive` for non-existent service
- **THEN** system SHALL return 404 not found

### Requirement: Archived service retains monitor configuration
System SHALL preserve monitor configurations when service is archived.

#### Scenario: Archived service keeps monitor configurations
- **WHEN** service is archived
- **THEN** all monitors under the service SHALL retain their configuration (enabled/disabled state, probe locations, interval)
- **AND** no monitors SHALL be automatically disabled

### Requirement: Archived service can be queried normally
System SHALL allow normal read operations on archived services.

#### Scenario: Client fetches archived service
- **WHEN** client fetches archived service via `GET /api/v1/services/{serviceId}`
- **THEN** system SHALL return service with lifecycleState `archived`
- **AND** monitors SHALL be listed with their current configurations
