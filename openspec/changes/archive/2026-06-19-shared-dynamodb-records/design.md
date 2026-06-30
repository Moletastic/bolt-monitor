## Context

`monitor-api` and `check-runtime` both define DynamoDB record structs for persisting domain objects. Many are identical; some are similar but with field ordering differences or extra fields in one version. Both files also use inline `strings.ToUpper/ToLower/TrimSpace` instead of the existing `normalizeField` and `normalizeToken` helpers in `shared/dynamodbschema/schema.go`.

## Goals / Non-Goals

**Goals:**
- Move6 identical record types to shared package
- Unify 3 similar record types (monitor-api version wins)
- Co-locate helper constructors with their types
- Replace inline normalization with calls to existing schema helpers

**Non-Goals:**
- No interfaces — repository interfaces remain per-service (different operational needs)
- No new value object types (typed TenantID etc.)
- No behavioral changes
- No changes to DynamoDB schema itself

## Decision 1: File Layout

**Choice:** One file per domain entity group in `shared/dynamodbrecord/`

```
shared/dynamodbrecord/
  audit.go # auditChangeRecord, auditEventRecord
  execution.go    # executionWorkItemRecord
  incident.go    # incidentRecord, incidentItemRecord, incidentActivityRecord
  monitor.go     # monitorItemRecord
  service.go     # serviceItemRecord, serviceStatusRecord
  scheduler.go   # schedulerConfigItemRecord
```

**Rationale:** Groups related records, avoids 10+ tiny files. Natural grouping matches the domain.

**Alternatives considered:**
- One file per type: Too many files
- All in one file: Too large, loses grouping signal

## Decision 2: Canonical Version for Similar Types

**Choice:** monitor-api version is canonical for all three similar types

| Type | Resolution |
|------|------------|
| `monitorItemRecord` | monitor-api field order wins (TenantID before ServiceID) |
| `serviceItemRecord` | monitor-api full version wins (check-runtime gains unused fields) |
| `schedulerConfigItemRecord` | monitor-api full version wins |

**Rationale:** monitor-api has the richer model. check-runtime only needs the subset it uses today; extra fields are harmless.

## Decision 3: Helper Constructor Location

**Choice:** Each `newXxxItemRecord()` function lives in the same file as its record type

**Rationale:** Keeps type and construction logic together. Aligns with existing pattern in `dynamodbschema/schema.go`.

## Decision 4: Normalization Calls

**Choice:** Replace inline normalization in repository constructors with calls to `dynamodbschema.NormalizeField()` and `dynamodbschema.NormalizeToken()`

```go
// Before
TenantID: strings.ToUpper(strings.TrimSpace(service.TenantID)),

// After
TenantID: dynamodbschema.NormalizeField(service.TenantID),
```

**Rationale:** Existing helpers already exist and are tested. Centralizes normalization logic. `NormalizeField` = lowercase, `NormalizeToken` = uppercase.

## Decision 5: Record Type Method Naming

**Choice:** Keep existing method names (`toMonitor()`, `toService()`, `toIncident()`, `toWork()`) as-is

**Rationale:** These are already exported and used internally. No reason to rename.

## Identical Types to Move

| Type | Fields | Helper Functions |
|------|--------|-----------------|
| `auditChangeRecord` | PK, SK, EntityType, AuditID, FieldPath, OldValue, NewValue | `newAuditChangeRecord()` |
| `executionWorkItemRecord` | PK, SK, EntityType, TenantID, ServiceID, MonitorID, RunID, ProbeLocationID, Trigger, AcceptedAt, Status, StartedAt, CompletedAt, LastError | `newExecutionWorkItemRecords()`, `executionWorkItemRecordFromWork()` |
| `incidentRecord` | IncidentID, ServiceID, MonitorID, TenantID, Summary, Status, OpenedAt, AcknowledgedAt, ResolvedAt, UpdatedAt, Origin | None |
| `incidentItemRecord` | PK, SK, EntityType, TenantID, ServiceID, MonitorID, IncidentID, Summary, Status, OpenedAt, AcknowledgedAt, ResolvedAt, UpdatedAt, Origin, GSI1PK, GSI1SK | `newIncidentMonitorItemRecord()`, `newIncidentRefItemRecord()`, `newIncidentMetaItemRecord()` |
| `serviceStatusRecord` | PK, SK, EntityType, TenantID, ServiceID, LifecycleState, RollupStatus, MonitorCount, EnabledMonitorCount, UpdatedAt, GSI2PK, GSI2SK | `newServiceStatusItemRecord()` |
| `incidentActivityRecord` | PK, SK, EntityType, TenantID, IncidentID, ActivityID, Action, Timestamp | `newIncidentActivityRecord()` |
| `auditEventRecord` | PK, SK, EntityType, TenantID, ServiceID, MonitorID, AuditID, Action, ResourceID, Timestamp, Origin | `newAuditEventRecord()` |

## Similar Types to Unify

| Type | monitor-api | check-runtime | Resolution |
|------|------------|---------------|------------|
| `monitorItemRecord` | 11 fields, TenantID first | 11 fields, different order | monitor-api order wins |
| `serviceItemRecord` | 8 fields (Name, Description, LifecycleState, TechnologyKey, CreatedAt, UpdatedAt + base) | 3 fields (base only) | monitor-api full version wins |
| `schedulerConfigItemRecord` | 7 fields (PK, SK, EntityType, TenantID, RecurringEnabled, StopControlMode, UpdatedAt) | 2 fields (RecurringEnabled, StopControlMode) | monitor-api full version wins |

## Migration Plan

1. Create `shared/dynamodbrecord/` directory
2. Create `shared/dynamodbrecord/audit.go` with `auditChangeRecord`, `auditEventRecord` and helpers
3. Create `shared/dynamodbrecord/execution.go` with `executionWorkItemRecord` and helpers
4. Create `shared/dynamodbrecord/incident.go` with `incidentRecord`, `incidentItemRecord`, `incidentActivityRecord` and helpers
5. Create `shared/dynamodbrecord/monitor.go` with `monitorItemRecord` and helpers
6. Create `shared/dynamodbrecord/service.go` with `serviceItemRecord`, `serviceStatusRecord` and helpers
7. Create `shared/dynamodbrecord/scheduler.go` with `schedulerConfigItemRecord` and helpers
8. Update `services/monitor-api/repository.go` to import from `shared/dynamodbrecord/`
9. Update `services/check-runtime/repository.go` to import from `shared/dynamodbrecord/`
10. Replace inline normalization with `dynamodbschema.NormalizeField()` / `NormalizeToken()` calls
11. Delete duplicate type definitions from both repository files
12. Run `go build` and `go test` to verify

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Two-phase cutover (define in shared, then delete from repos) | Do it in single commit: add shared + update imports + delete originals |
| Any method renaming would break callers | Keep existing method names; no refactoring of method signatures |
| check-runtime might need fields it doesn't currently use | Full monitor-api version covers all use cases |
