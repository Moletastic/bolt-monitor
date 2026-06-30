## 1. OpenSpec Finalization

- [x] 1.1 Review generated proposal, design, and capability deltas for service-first scope alignment
- [x] 1.2 Resolve remaining design questions about service deletion and activation behavior before implementation begins

## 2. Shared Domain Models

- [x] 2.1 Replace top-level monitor identity model with top-level service model and nested monitor model in shared contracts
- [x] 2.2 Add immutable slug validation for `serviceId` and `monitorId`
- [x] 2.3 Add optional validated `technologyKey` to service contracts
- [x] 2.4 Update monitor configuration contracts to require `serviceId` ownership and preserve monitor-type-specific settings

## 3. Storage Schema And Repository Mapping

- [x] 3.1 Replace DynamoDB schema helpers with tenant-aware service and monitor composite key helpers
- [x] 3.2 Add item mappings for service summary, service metadata, service status, nested monitor summary, monitor metadata, monitor status, runs, and incidents
- [x] 3.3 Update repository read and write paths to use service-first item families and nested monitor identities
- [x] 3.4 Remove obsolete monitor-first key helpers and record-mapping assumptions

## 4. API Surface Refactor

- [x] 4.1 Add service create, list, get, and update handlers and routes
- [x] 4.2 Replace flat monitor CRUD routes with nested service-monitor CRUD routes
- [x] 4.3 Replace flat monitor enable and disable routes with nested service-monitor action routes
- [x] 4.4 Update manual-run, incident-read, and audit-read routes to use nested service-monitor paths
- [x] 4.5 Enforce service lifecycle transition rules, nested monitor uniqueness, and `technologyKey` validation in handlers

## 5. Runtime Status And Rollups

- [x] 5.1 Update execution result persistence to use tenant-aware service-monitor identities for raw runs and latest monitor status
- [x] 5.2 Derive and persist service monitor summaries from latest nested monitor status
- [x] 5.3 Derive and persist service rollup status from enabled child monitor states
- [x] 5.4 Prevent active services from entering invalid zero-monitor operating state through delete and lifecycle checks

## 6. Dashboard Refactor

- [x] 6.1 Replace monitor-backed `/services` overview with real service summary views
- [x] 6.2 Add service create and edit flows that support `technologyKey` selection
- [x] 6.3 Update monitor create, detail, and management flows to use nested service-monitor APIs
- [x] 6.4 Render one service icon from `technologyKey` and derive monitor icons from monitor protocol or type in frontend

## 7. Development Reset And Verification

- [x] 7.1 Reset or remove obsolete development DynamoDB items created by the old monitor-first model
- [x] 7.2 Update repository, API, runtime, and dashboard tests for service-first identities and nested routes
- [x] 7.3 Verify service CRUD, nested monitor CRUD, rollup derivation, manual runs, incidents, audit reads, and dashboard service views end to end
