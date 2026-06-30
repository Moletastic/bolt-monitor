## 1. Create shared record package

- [x] 1.1 Create `shared/dynamodbrecord/` directory
- [x] 1.2 Add `shared/dynamodbrecord/go.mod` with module name `bolt-monitor/shared/dynamodbrecord`

## 2. Move audit record types

- [x] 2.1 Create `shared/dynamodbrecord/audit.go`
- [x] 2.2 Define `auditChangeRecord` struct with DynamoDB tags
- [x] 2.3 Add `newAuditChangeRecord()` constructor
- [x] 2.4 Define `auditEventRecord` struct with DynamoDB tags
- [x] 2.5 Add `newAuditEventRecord()` constructor

## 3. Move execution record types

- [x] 3.1 Create `shared/dynamodbrecord/execution.go`
- [x] 3.2 Define `executionWorkItemRecord` struct with DynamoDB tags
- [x] 3.3 Add `newExecutionWorkItemRecords()` constructor
- [x] 3.4 Add `executionWorkItemRecordFromWork()` constructor

## 4. Move incident record types

- [x] 4.1 Create `shared/dynamodbrecord/incident.go`
- [x] 4.2 Define `incidentRecord` struct (plain, no DynamoDB tags)
- [x] 4.3 Define `incidentItemRecord` struct with DynamoDB tags
- [x] 4.4 Add `newIncidentMonitorItemRecord()`, `newIncidentRefItemRecord()`, `newIncidentMetaItemRecord()` constructors
- [x] 4.5 Define `incidentActivityRecord` struct with DynamoDB tags
- [x] 4.6 Add `newIncidentActivityRecord()` constructor

## 5. Move monitor record types

- [x] 5.1 Create `shared/dynamodbrecord/monitor.go`
- [x] 5.2 Define `monitorItemRecord` struct with DynamoDB tags (monitor-api field order)
- [x] 5.3 Add `newMonitorItemRecord()` constructor
- [x] 5.4 Add `newServiceMonitorRefItemRecord()` constructor
- [x] 5.5 Add `toMonitor()` method

## 6. Move service record types

- [x] 6.1 Create `shared/dynamodbrecord/service.go`
- [x] 6.2 Define `serviceItemRecord` struct with DynamoDB tags (full monitor-api version)
- [x] 6.3 Add `newServiceItemRecord()`, `newServiceRefItemRecord()` constructors
- [x] 6.4 Add `toService()` method
- [x] 6.5 Define `serviceStatusRecord` struct with DynamoDB tags
- [x] 6.6 Add `newServiceStatusItemRecord()` constructor

## 7. Move scheduler record types

- [x] 7.1 Create `shared/dynamodbrecord/scheduler.go`
- [x] 7.2 Define `schedulerConfigItemRecord` struct with DynamoDB tags (full monitor-api version)
- [x] 7.3 Add `newSchedulerConfigItemRecord()` constructor
- [x] 7.4 Add `toSchedulerConfig()` method

## 8. Update monitor-api repository

- [x] 8.1 Add import for `bolt-monitor/shared/dynamodbrecord`
- [x] 8.2 Add import for `bolt-monitor/shared/dynamodbschema`
- [x] 8.3 Replace inline record type definitions with imports from `dynamodbrecord`
- [x] 8.4 Replace inline normalization with `dynamodbschema.NormalizeField()` / `NormalizeToken()`
- [x] 8.5 Delete duplicate type definitions from `repository.go`
- [x] 8.6 Verify `go build ./services/monitor-api/...` passes

## 9. Update check-runtime repository

- [x] 9.1 Add import for `bolt-monitor/shared/dynamodbrecord`
- [x] 9.2 Add import for `bolt-monitor/shared/dynamodbschema`
- [x] 9.3 Replace inline record type definitions with imports from `dynamodbrecord`
- [x] 9.4 Replace inline normalization with `dynamodbschema.NormalizeField()` / `NormalizeToken()`
- [x] 9.5 Delete duplicate type definitions from `repository.go`
- [x] 9.6 Verify `go build ./services/check-runtime/...` passes

## 10. Verify

- [x] 10.1 Run `go build ./services/... ./shared/...`
- [x] 10.2 Run `go test ./services/... ./shared/...`
- [x] 10.3 Run `make lint-go`
