## Why

Operators need a fast way to jump to services, monitors, notification routes, and notification channels without navigating module-by-module. A global top-bar search can reduce operational friction, but it needs a backend search endpoint and DynamoDB access pattern that avoids expensive scans as the tenant dataset grows.

## What Changes

- Add a global dashboard top-bar search input with search icon, debounced requests, loading/empty/error feedback, and clickable result links.
- Add a new backend endpoint for tenant-scoped global resource search across services, monitors, escalation policies, and notification channels.
- Return typed search results with enough resource identity and display data for the dashboard to render icons, text, and destination links.
- Add a sparse DynamoDB search-index item pattern that supports prefix queries by normalized searchable terms without scanning resource collections.
- Maintain search index items when services, monitors, escalation policies, and notification channels are created, updated, or deleted.
- Define searchable fields for each resource beyond opaque IDs/ULIDs.

## Capabilities

### New Capabilities

- `global-resource-search-api`: Backend API for searching services, monitors, escalation policies, and notification channels.

### Modified Capabilities

- `dashboard-web-app`: Add global top-bar search UI, debounce behavior, result rendering, and navigation links.
- `dynamodb-single-table-storage`: Add tenant-scoped sparse search-index records and access pattern for low-I/O global resource search.

## Impact

- Affected API route: new `GET /api/v1/search` route in `services/monitor-api` and `infra/stacks/bootstrap.ts`.
- Affected repository/storage: write and delete search-index records for services, monitors, escalation policies, and notification channels.
- Affected shared schema/records: new search-index record type and key helpers.
- Affected dashboard: `AppShell` top bar and API client/types for search.
- No external search service is introduced.
- Existing service, monitor, escalation policy, and channel APIs remain compatible.
