## Why

Monitoring product should run checks from system-managed probe locations, not arbitrary user-defined region strings. Adding a probe-location catalog now keeps scheduling, metrics, routing, and per-tenant monitor configuration consistent before CRUD APIs and execution logic ship.

## What Changes

- Define a first-class probe-location catalog capability managed by the system.
- Define how probe locations are identified, enabled, and exposed for monitor selection.
- Update monitor configuration requirements so monitors select probe locations from the system catalog instead of free-form region values.
- Establish tenant-safe validation expectations for probe-location selection.

## Capabilities

### New Capabilities
- `probe-location-catalog`: Canonical system-defined catalog of valid probe locations that monitors can target.

### Modified Capabilities
- `monitor-configuration`: Replace free-form execution region semantics with system-defined probe-location selection and validation.

## Impact

- Affects shared monitor model naming and validation rules.
- Affects future monitor CRUD APIs and scheduler routing.
- Affects DynamoDB item shapes and tenant-scoped query flows for monitor configuration.
