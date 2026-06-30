## Why

Currently `serviceId` is a client-provided slug, which creates coupling between client and server and limits ID generation strategies. Server-generated IDs (ULIDs) provide better isolation, collision resistance, and monotonic ordering properties that slugs cannot offer. Similarly, `monitorId` is client-provided slug — it should be server-generated from protocol+URL or name for consistency.

## What Changes

1. **ServiceId: server-generated ULID** — remove `serviceId` from `CreateServiceRequest`, generate server-side using existing ULID infrastructure
2. **MonitorId: server-generated slug** — remove `monitorId` from `CreateMonitorRequest`, generate server-side from `type+target URL` or fallback to `name`
3. **Return generated IDs in response** — clients receive the server-generated ID in the 201 response `Location` header and response body
4. **Remove ID validation** — no longer need slug validation for IDs that are server-generated

## Capabilities

### New Capabilities

- **service-id-auto-generation**: Service ID is server-generated ULID with `SVC_` prefix. Client no longer provides `serviceId` on create. Response includes generated `serviceId`.
- **monitor-id-auto-generation**: Monitor ID is server-generated slug derived from monitor type + target URL or name fallback. Client no longer provides `monitorId` on create. Response includes generated `monitorId`.

### Modified Capabilities

- **service-management-api**: Remove `serviceId` field from create request. Service ID returned in response body and Location header.
- **monitor-crud-api**: Remove `monitorId` field from create request. Monitor ID returned in response body and Location header.

## Impact

- **API**: Breaking change — `serviceId` and `monitorId` no longer accepted on create
- **Client**: Must handle server-assigned IDs returned in responses
- **Dashboard**: Must be updated to work with generated IDs instead of client-generated slugs
