## 1. Domain Model Changes

- [x] 1.1 Remove `LifecycleState` from `CreateServiceRequest` struct in `shared/monitorconfig/model.go`
- [x] 1.2 Remove `LifecycleState` from `updateServiceRequest` struct in `services/monitor-api/types.go`
- [x] 1.3 Keep `LifecycleState` field in `Service` struct (computed field, not settable)
- [x] 1.4 Update `Service.Validate()` to remove lifecycle state validation (no longer client-settable)

## 2. Repository Layer Changes

- [x] 2.1 Modify `SetMonitorEnabled` to recompute and persist derived lifecycle state in same transaction
- [x] 2.2 Add `ArchiveService` method to transition service to `archived` lifecycle state
- [x] 2.3 Add `ReactivateService` method to transition archived service to `draft` or `active` based on enabledCount
- [x] 2.4 Simplify `deriveServiceRollup` to remove lifecycle-based shortcuts (draft/archived no longer short-circuit to special rollup values)
- [x] 2.5 Ensure `GetService` returns computed lifecycle state based on `enabledCount`

## 3. Handler Layer Changes

- [x] 3.1 Remove `lifecycleState` parameter handling from `createService` handler (lifecycle always starts as `draft`)
- [x] 3.2 Remove `lifecycleState` parameter handling from `updateService` handler (lifecycle not mutable via PATCH)
- [x] 3.3 Remove activation guard check (`EnabledCount == 0` guard) from `createService` and `updateService`
- [x] 3.4 Add handler for `POST /api/v1/services/{serviceId}/archive` endpoint
- [x] 3.5 Add handler for `POST /api/v1/services/{serviceId}/reactivate` endpoint
- [x] 3.6 Update route matching in `handleRequest` for new endpoints

## 4. API Response Changes

- [x] 4.1 Ensure `serviceResponse` still includes `lifecycleState` as read-only computed field
- [x] 4.2 Update OpenAPI spec if generated (or mark lifecycleState as read-only in API docs)

## 5. Testing Changes

- [x] 5.1 Update `main_test.go` to remove tests for lifecycle state transitions via PATCH
- [x] 5.2 Add tests for lifecycle auto-derive on monitor enable/disable
- [x] 5.3 Add tests for archive endpoint (active, draft, already archived, not found)
- [x] 5.4 Add tests for reactivate endpoint (with enabled monitors, without, non-archived, not found)
- [x] 5.5 Update existing service creation tests to expect initial `draft` lifecycle
- [x] 5.6 Verify all Go tests pass with `cd services/monitor-api && go test ./...`

## 6. Integration Verification

- [x] 6.1 Run `cd services/monitor-api && go test ./...` to verify all tests pass
- [x] 6.2 Run `cd infra && npm run check` to verify infra typechecks
- [x] 6.3 Run `cd apps/dashboard && npm run lint` to verify dashboard lints clean
