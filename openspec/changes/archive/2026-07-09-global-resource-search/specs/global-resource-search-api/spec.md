## ADDED Requirements

### Requirement: System exposes global resource search API

The monitor API SHALL expose a tenant-scoped global search endpoint for services, monitors, escalation policies, and notification channels.

#### Scenario: Operator searches supported resources
- **WHEN** the dashboard calls `GET /api/v1/search?q=<query>` with a valid query
- **THEN** system returns a success response envelope containing search results for matching services, monitors, escalation policies, and notification channels
- **AND** each result includes a resource `type`, stable `id`, display `label`, display `description`, navigation `href`, and icon discriminator
- **AND** monitor results include the parent `serviceId`

#### Scenario: Query is too short
- **WHEN** the dashboard calls `GET /api/v1/search` with a normalized query shorter than the configured minimum length
- **THEN** system returns a success response envelope with an empty result list
- **AND** system does not scan service, monitor, policy, or channel collections

#### Scenario: Search is limited
- **WHEN** the dashboard calls `GET /api/v1/search?q=<query>&limit=<n>`
- **THEN** system returns no more than the accepted result limit
- **AND** system enforces a maximum limit to protect read capacity

#### Scenario: Search type filter is provided
- **WHEN** the dashboard calls `GET /api/v1/search?q=<query>&types=service,policy,channel`
- **THEN** system only returns results for the requested supported resource types
- **AND** system ignores or rejects unsupported resource types with a validation error

#### Scenario: Search result has safe display text
- **WHEN** system returns monitor, escalation policy, or notification channel results
- **THEN** returned display text excludes monitor headers, expected body text, notification channel config JSON, inline channel config, and secret-bearing target fragments

### Requirement: System maintains global search index entries

The monitor API SHALL keep search index entries in sync with service, monitor, escalation policy, and notification channel lifecycle operations.

#### Scenario: Searchable resource is created
- **WHEN** a service, monitor, escalation policy, or notification channel is created
- **THEN** system writes search index entries for the resource using normalized searchable fields

#### Scenario: Searchable resource is updated
- **WHEN** a service, monitor, escalation policy, or notification channel changes searchable fields
- **THEN** system replaces stale search index entries with entries reflecting the updated resource

#### Scenario: Searchable resource is deleted
- **WHEN** a service, monitor, escalation policy, or notification channel is deleted
- **THEN** system removes search index entries for the deleted resource

#### Scenario: Search index is backfilled
- **WHEN** existing resources predate search index maintenance
- **THEN** system provides a bounded backfill path to create search index entries for those resources
