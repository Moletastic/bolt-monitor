## ADDED Requirements

### Requirement: OpenAPI documents delivery pagination
OpenAPI SHALL document optional opaque `cursor`, bounded `limit`, and cursor pagination response metadata for incident delivery listing.

#### Scenario: Developer reviews delivery list operation
- **WHEN** a developer reads the incident delivery list operation
- **THEN** the contract explains how to request the first and following pages without calculating a total
