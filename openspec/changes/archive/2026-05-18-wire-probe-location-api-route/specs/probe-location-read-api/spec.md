## ADDED Requirements

### Requirement: Bootstrap API exposes probe-location collection route
System SHALL expose the probe-location collection route through the bootstrap API infrastructure.

#### Scenario: Client requests probe-location collection through bootstrap API
- **WHEN** client requests `GET /api/v1/probe-locations` from the SST-managed API surface
- **THEN** infrastructure routes the request to the monitor API handler
- **AND** the handler can return the selectable probe-location collection response defined by the probe-location read capability
