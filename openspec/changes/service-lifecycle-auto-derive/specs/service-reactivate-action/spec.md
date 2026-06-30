## ADDED Requirements

### Requirement: System allows reactivating archived services via explicit action
System SHALL provide an explicit endpoint to transition an archived service back to an active or draft state.

#### Scenario: Client reactivates archived service with enabled monitors
- **WHEN** client sends `POST /api/v1/services/{serviceId}/reactivate` for an archived service
- **AND** archived service has `enabledCount` greater than zero
- **THEN** system SHALL transition service lifecycle to `active`

#### Scenario: Client reactivates archived service with no enabled monitors
- **WHEN** client sends `POST /api/v1/services/{serviceId}/reactivate` for an archived service
- **AND** archived service has `enabledCount` of zero
- **THEN** system SHALL transition service lifecycle to `draft`

#### Scenario: Client reactivates non-archived service
- **WHEN** client sends `POST /api/v1/services/{serviceId}/reactivate` for a service that is not archived
- **THEN** system SHALL return 409 conflict with error message indicating service is not archived

#### Scenario: Client reactivates missing service
- **WHEN** client sends `POST /api/v1/services/{serviceId}/reactivate` for non-existent service
- **THEN** system SHALL return 404 not found

### Requirement: Reactivation preserves monitor configuration
System SHALL preserve monitor configurations when service is reactivated.

#### Scenario: Reactivated service keeps monitor configurations
- **WHEN** archived service is reactivated
- **THEN** all monitors under the service SHALL retain their configuration
- **AND** no monitors SHALL be automatically enabled or disabled
