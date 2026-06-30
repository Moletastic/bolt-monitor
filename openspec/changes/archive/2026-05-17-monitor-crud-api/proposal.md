## Why

Project has monitor model, probe-location catalog, and DynamoDB single-table contract, but no product API for users to create or manage monitors. Basic CRUD endpoints are next step because they turn the shared contracts into usable product surface and unblock scheduler, alerting, and dashboard work.

## What Changes

- Add basic monitor CRUD HTTP endpoints under `/api/v1/monitors`.
- Validate monitor payloads against shared monitor and probe-location contracts.
- Persist monitor data using single-table DynamoDB item families.
- Write audit records for configuration changes.
- Return monitor resources in stable API response shape.

## Capabilities

### New Capabilities
- `monitor-crud-api`: Basic HTTP API for creating, reading, listing, updating, enabling, and disabling monitors.

### Modified Capabilities

## Impact

- Affects API Gateway/Lambda routing and Go backend handlers.
- Affects DynamoDB repository code for monitor and audit item writes.
- Reuses `shared/monitorconfig`, `shared/probelocationcatalog`, and `shared/dynamodbschema` as source contracts.
