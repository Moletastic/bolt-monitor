## Why

Project now has monitor configuration contract, but no persistence contract for how monitor data and future operational records should live in DynamoDB. Defining single-table design now prevents CRUD APIs, scheduler output, incidents, and audit features from inventing incompatible storage patterns.

## What Changes

- Define canonical single-table DynamoDB design for core monitoring entities.
- Specify partition/sort key strategy for tenant, monitor, status, run, incident, and audit items.
- Identify initial GSIs and retention expectations for high-volume records.
- Establish access-pattern-first storage contract for future API and execution work.

## Capabilities

### New Capabilities
- `dynamodb-single-table-storage`: Canonical single-table DynamoDB storage design for monitor, status, run, incident, and audit entities.

### Modified Capabilities

## Impact

- Affects future monitor CRUD handlers and DynamoDB write/read code.
- Affects scheduler, prober, incident, alerting, and audit storage patterns.
- Constrains future model-to-record mapping in shared contracts such as `shared/monitorconfig`.
