## ADDED Requirements

### Requirement: System exposes service-level audit history through HTTP API

System SHALL allow clients to read audit history for an individual service through HTTP API. This complements monitor-level audit events for per-incident audit views that need the full picture of what happened to a service and its child monitors.

#### Scenario: Operator requests service audit history
- **WHEN** operator calls `GET /api/v1/services/{id}/audit` for an existing service
- **THEN** system returns audit events associated with that service, including service lifecycle events (archive, reactivate) and any monitor-level events that are relevant to the service scope

#### Scenario: Service has no recorded audit events
- **WHEN** operator requests audit history for an existing service with no recorded mutations
- **THEN** system returns a successful empty audit collection response

#### Scenario: Requested service does not exist
- **WHEN** operator calls `GET /api/v1/services/{id}/audit` for a non-existent service
- **THEN** system returns a 404 response with an error message

### Requirement: Service audit read responses expose the same metadata as monitor audit

System SHALL return service audit event metadata consistent with the monitor audit event shape — stable identity, event type, event timestamp, and origin metadata when available.

#### Scenario: Service audit response follows monitor audit shape
- **WHEN** system returns audit history for a service
- **THEN** each audit event includes stable identity, event type, event timestamp, and origin metadata consistent with the monitor audit event response shape
