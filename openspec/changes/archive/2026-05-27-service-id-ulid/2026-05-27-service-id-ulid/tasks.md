## 1. Domain Model Changes

- [x] 1.1 Remove `ServiceID` field from `CreateServiceRequest` in `shared/monitorconfig/model.go`
- [x] 1.2 Remove `MonitorID` field from `CreateMonitorRequest` in `shared/monitorconfig/model.go`
- [x] 1.3 Update `Service.Validate()` to remove serviceId slug validation
- [x] 1.4 Update `Monitor.Validate()` to remove monitorId slug validation

## 2. ID Generation Functions

- [x] 2.1 Add `newServiceID(now time.Time) string` function returning `SVC_` prefix + ULID
- [x] 2.2 Add `generateMonitorID(monitorType, targetURL, name string) string` function that derives slug from type+URL or name fallback
- [x] 2.3 Ensure existing ULID functions in `ids.go` remain unchanged

## 3. Handler Layer Changes

- [x] 3.1 Update `createService` handler to generate serviceId server-side, not read from request
- [x] 3.2 Update `createMonitor` handler to generate monitorId server-side, not read from request
- [x] 3.3 Update `toServiceResponse` to include generated serviceId in response
- [x] 3.4 Update `toMonitorResponse` to include generated monitorId in response
- [x] 3.5 Ensure `Location` header in 201 responses contains generated IDs

## 4. Testing Changes

- [x] 4.1 Update `main_test.go` to remove `serviceId` from create service request fixtures
- [x] 4.2 Update `main_test.go` to remove `monitorId` from create monitor request fixtures
- [x] 4.3 Update tests that assert on generated IDs (now returned in response, not provided in request)
- [x] 4.4 Verify all Go tests pass with `cd services/monitor-api && go test ./...`
