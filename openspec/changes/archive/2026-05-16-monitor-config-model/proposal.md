## Why

Project can answer `/api/health` but cannot describe what should actually be monitored. Defining monitor configuration model now gives future CRUD APIs, schedulers, probers, DynamoDB items, and dashboards one stable contract to build around.

## What Changes

- Define first-class monitor configuration capability for healthcheck resources.
- Specify required and optional fields for HTTP monitor definitions.
- Define monitor lifecycle state and validation expectations.
- Establish ownership and persistence expectations for monitor records in DynamoDB-backed system.

## Capabilities

### New Capabilities
- `monitor-configuration`: Canonical configuration model for monitor definitions, including identity, target, check behavior, cadence, and enablement state.

### Modified Capabilities

## Impact

- Affects future DynamoDB single-table design and item shapes.
- Affects future monitor CRUD API design under `/api/v1/...`.
- Affects scheduler, prober execution, alerting, and audit systems that will consume monitor configuration.
