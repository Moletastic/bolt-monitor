## MODIFIED Requirements

### Requirement: System documents monitor API without probe-location contracts
System SHALL document the monitor API according to the single-execution-environment product contract.

#### Scenario: API documentation shows monitor payloads
- **WHEN** OpenAPI examples or schemas describe monitor create, update, read, status, runs, or manual-run responses
- **THEN** they do not include `probeLocations`, `probeLocationId`, `lastProbeLocationId`, or hard-coded location examples such as `iad`

#### Scenario: API documentation lists monitor API paths
- **WHEN** OpenAPI paths are rendered
- **THEN** probe-location catalog endpoints are not documented as supported product APIs
