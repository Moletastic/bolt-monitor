## ADDED Requirements

### Requirement: System exposes selectable probe locations through HTTP API
System SHALL expose probe locations that are valid for monitor selection through HTTP API.

#### Scenario: Client requests probe-location catalog
- **WHEN** client requests probe-location collection
- **THEN** system returns probe locations available for operator selection in the current environment

#### Scenario: Catalog contains disabled location
- **WHEN** probe location is not enabled for operator selection
- **THEN** system excludes that location from the public probe-location collection response

### Requirement: System returns selection-friendly probe-location metadata
System SHALL return probe-location metadata suitable for frontend selection controls.

#### Scenario: Dashboard builds monitor form picker
- **WHEN** dashboard reads probe-location collection
- **THEN** system returns stable identifier and human-readable display metadata for each selectable probe location

#### Scenario: Client reads probe-location collection ordering
- **WHEN** system returns selectable probe locations
- **THEN** the collection is ordered by display name for stable operator-facing presentation

### Requirement: System exposes probe-location collection at stable route
System SHALL expose probe-location discovery through a stable top-level collection route.

#### Scenario: Client requests probe-location collection route
- **WHEN** client requests selectable probe locations
- **THEN** system serves the collection from `/api/v1/probe-locations`

### Requirement: Bootstrap API exposes probe-location collection route
System SHALL expose the probe-location collection route through the bootstrap API infrastructure.

#### Scenario: Client requests probe-location collection through bootstrap API
- **WHEN** client requests `GET /api/v1/probe-locations` from the SST-managed API surface
- **THEN** infrastructure routes the request to the monitor API handler
- **AND** the handler can return the selectable probe-location collection response defined by the probe-location read capability

### Requirement: System keeps probe-location discovery separate from monitor mutation
System SHALL expose probe-location discovery through a dedicated read surface rather than requiring clients to infer valid options from monitor mutation failures.

#### Scenario: Client prepares create-monitor form
- **WHEN** client needs valid probe-location options before submitting monitor data
- **THEN** system provides those options through probe-location read API without requiring a failed monitor create or update attempt
