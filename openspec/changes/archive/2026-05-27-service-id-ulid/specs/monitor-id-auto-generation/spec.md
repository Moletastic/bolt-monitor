## ADDED Requirements

### Requirement: Monitor ID is generated server-side from type and target
System SHALL generate monitor IDs server-side as slugs derived from monitor type and target URL.

#### Scenario: Monitor ID derived from HTTP target URL
- **WHEN** monitor of type `http` is created with target `https://api.example.com/health`
- **THEN** system generates a monitor ID slug derived from type + URL
- **AND** ID format is deterministic for same type+URL combination

#### Scenario: Monitor ID fallback to name-derived slug
- **WHEN** monitor is created but target URL cannot be parsed
- **THEN** system generates monitor ID slug derived from monitor name
- **AND** slug is URL-safe and human-readable

#### Scenario: Monitor ID returned in response
- **WHEN** monitor is created successfully
- **THEN** response body includes the generated `monitorId`
- **AND** `Location` header contains URL with generated monitor ID

#### Scenario: Monitor ID includes collision prevention
- **WHEN** monitor ID derived from type+URL already exists under same service
- **THEN** system generates unique ID with additional suffix to avoid collision

### Requirement: Client does not provide monitor ID
System SHALL NOT accept client-provided monitor ID on create.

#### Scenario: Create request without monitorId
- **WHEN** client submits create request without `monitorId` field
- **THEN** system generates monitor ID server-side

#### Scenario: Create request with monitorId field
- **WHEN** client submits create request including `monitorId` field
- **THEN** system ignores the provided value
- **AND** generates server-side ID instead
