## ADDED Requirements

### Requirement: System exposes service-scoped incident read API

The monitor API SHALL expose a bounded service-scoped incident read endpoint for incidents related to monitors under a service.

#### Scenario: Operator requests service incidents
- **WHEN** the dashboard calls `GET /api/v1/services/{serviceId}/incidents?limit=<n>` for an existing service
- **THEN** system returns a success response envelope containing incidents for monitors under that service
- **AND** results are sorted newest first
- **AND** results are limited by the accepted limit

#### Scenario: Service incident request uses default limit
- **WHEN** the dashboard calls `GET /api/v1/services/{serviceId}/incidents` without a limit
- **THEN** system applies a bounded default limit suitable for a recent alerts panel

#### Scenario: Service does not exist
- **WHEN** the dashboard calls the service incidents endpoint for a missing service
- **THEN** system returns the existing service-not-found error response

#### Scenario: Service has no incidents
- **WHEN** no incidents exist for monitors under the service
- **THEN** system returns a success response envelope with an empty incident list

#### Scenario: Service incident query is bounded
- **WHEN** system reads service-scoped incidents
- **THEN** system avoids listing unrelated incidents across all services
- **AND** system avoids issuing an unbounded request per monitor
