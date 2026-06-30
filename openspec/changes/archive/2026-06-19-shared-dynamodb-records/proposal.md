## Why

Record types are duplicated across `services/monitor-api/repository.go` and `services/check-runtime/repository.go`. Six types are literally identical. Three are similar but differ in field ordering or completeness. Additionally, inline normalization (`strings.ToUpper/ToLower/TrimSpace`) is used directly in record constructors instead of calling the existing helpers in `shared/dynamodbschema/schema.go`.

This duplication causes:
- Maintenance burden: changing a field requires edits in two places
- Inconsistency risk: similar types can drift apart over time
- Cognitive load: understanding the data model requires reading multiple files

## What Changes

- Move6 identical record types to `shared/dynamodbrecord/`
- Unify 3 similar record types (pick monitor-api version as canonical)
- Move helper constructors alongside their types
- Update both repository files to import from shared and use `dynamodbschema.NormalizeField()` / `NormalizeToken()` instead of inline normalization

## Capabilities

### Modified Capabilities

- `dynamodb-record-types`: Consolidated into `shared/dynamodbrecord/` package
- `dynamodb-repositories`: Both `dynamoMonitorRepository` and `dynamoRuntimeRepository` import shared record types

## Impact

- New directory `shared/dynamodbrecord/` with 6 files
- `services/monitor-api/repository.go`: imports shared record types, uses schema helpers
- `services/check-runtime/repository.go`: imports shared record types, uses schema helpers
- No behavioral changes — pure data structure consolidation
