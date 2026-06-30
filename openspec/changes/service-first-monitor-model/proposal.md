## Why

Dashboard UX already presents `Services` as the operator-facing module, but the backend, shared contracts, and DynamoDB schema still treat `Monitor` as the top-level managed resource. That mismatch now blocks a clean service health model and hides partial outages that require multiple monitors under one logical service.

## What Changes

- Add a top-level `Service` resource with nested `Monitor` child resources.
- Replace flat monitor management APIs with service CRUD plus nested monitor CRUD, lifecycle, and runtime routes.
- Replace global monitor identity with tenant-scoped `serviceId` slugs and service-scoped `monitorId` slugs.
- Add derived service rollup states based on enabled child monitor states.
- Add one optional validated `technologyKey` per service for primary service icon rendering.
- Keep monitor icon presentation frontend-derived from monitor protocol or type rather than persisted monitor metadata.
- Replace monitor-first DynamoDB item families and access patterns with service-first storage layout while preserving monitor-scoped runtime history.
- **BREAKING** Remove reliance on flat top-level monitor routes and monitor-first storage assumptions as canonical product behavior.

## Capabilities

### New Capabilities
- `service-management-api`: top-level service CRUD, lifecycle, summary, and service metadata behavior including optional `technologyKey`.
- `service-status-rollup-model`: derived service health semantics based on lifecycle state and enabled child monitor states.

### Modified Capabilities
- `monitor-configuration`: monitor identity becomes service-scoped child identity with stable slug semantics under a parent service.
- `monitor-crud-api`: monitor creation, reads, updates, and lifecycle actions move under nested service paths.
- `monitor-status-read-api`: monitor status and run reads move under nested service-monitor routes and support service-backed dashboard reads.
- `incident-management-api`: monitor-scoped incident reads move to nested service-monitor paths while preserving incident overview and operator actions.
- `manual-run-api`: manual run commands target nested service-monitor identities instead of flat monitor IDs.
- `audit-event-read-api`: monitor audit history reads target nested service-monitor identities instead of flat monitor IDs.
- `dynamodb-single-table-storage`: core key patterns and item families pivot from monitor-first storage to service-first storage.
- `check-result-status-model`: status and run persistence continue monitor-scoped runtime storage under tenant-aware composite service-monitor identities and feed service rollups.
- `dashboard-web-app`: services module becomes real service-first UX with nested monitor management rather than a monitor overview wearing service labeling.

## Impact

- Affected APIs: `services/monitor-api` service and monitor routes, incident reads, audit reads, manual run commands, dashboard fetch contracts.
- Affected storage: `shared/dynamodbschema`, repository record mappings, status/run persistence keys, service summary item families.
- Affected domain models: `shared/monitorconfig`, `shared/resultstatus`, dashboard types, shared validation contracts.
- Affected UI: `/services` overview, monitor detail navigation, service forms, nested monitor forms, service icon rendering.
- Development data impact: old DynamoDB items may be deleted or reset because repository is still in development phase and production migration is out of scope.
