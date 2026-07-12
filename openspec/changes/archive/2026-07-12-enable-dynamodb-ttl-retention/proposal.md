## Why

High-volume DynamoDB records currently accumulate without enforced retention, increasing storage cost and query overhead over time. `CheckRun` already carries a 30-day TTL attribute, but the application table does not enable DynamoDB TTL, and `ExecutionWork` has no cleanup model.

## What Changes

- Enable DynamoDB Time to Live on the primary SST DynamoDB table using the `TTL` attribute.
- Enforce existing 30-day `CheckRun` retention through DynamoDB TTL deletion.
- Add TTL metadata to persisted `ExecutionWork` items so completed, skipped, and stale work records age out automatically.
- Keep archival to S3 out of scope; this change deletes expired operational records instead of preserving cold copies.
- Keep audit, incident, configuration, status, and derived search records out of scope.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `dynamodb-single-table-storage`: require the application table to enable DynamoDB TTL on the shared `TTL` attribute for expiring eligible item families.
- `check-result-status-model`: make raw `CheckRun` retention operationally enforced through table TTL, not merely documented in item shape.
- `check-execution-pipeline`: define retention expectations for persisted execution work records.

## Impact

- `infra/stacks/bootstrap.ts`: SST Dynamo table gains `ttl: 'TTL'`.
- `shared/dynamodbrecord/execution.go` and related models: `ExecutionWork` records gain numeric Unix epoch TTL metadata.
- `services/check-runtime`: execution work persistence writes TTL values without changing queue behavior.
- Tests: Go model/repository tests and infra type check should verify TTL wiring.
- Operations: expired items are deleted asynchronously by DynamoDB; items without `TTL` remain unaffected.
