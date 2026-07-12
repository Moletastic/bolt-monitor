## 1. Infrastructure TTL

- [x] 1.1 Enable DynamoDB Time to Live on the SST `AppTable` using the `TTL` attribute.
- [x] 1.2 Run infra type checking to verify the SST `ttl` configuration is accepted.

## 2. CheckRun Retention Verification

- [x] 2.1 Confirm `CheckRun` records continue to write numeric Unix epoch-second `TTL` values at the 30-day retention window.
- [x] 2.2 Keep existing run list behavior unchanged while relying on DynamoDB TTL for eventual removal.

## 3. ExecutionWork Retention

- [x] 3.1 Add `TTL` metadata to `ExecutionWorkItemRecord` using a documented execution-work retention window.
- [x] 3.2 Set `ExecutionWork` TTL when scheduler/manual-run paths create work records.
- [x] 3.3 Preserve or recompute `ExecutionWork` TTL when worker paths mark work in progress, completed, or skipped.
- [x] 3.4 Add or update Go tests covering `ExecutionWork` TTL serialization and retention calculation.

## 4. Verification

- [x] 4.1 Run `make test-go-all`.
- [x] 4.2 Run `make check-infra`.
- [x] 4.3 Run `openspec status --change enable-dynamodb-ttl-retention` and confirm all tasks are ready for implementation tracking.

## 5. Temporary Backfill Endpoint

- [x] 5.1 Add a temporary admin endpoint that backfills missing `ExecutionWork` TTL values.
- [x] 5.2 Deploy the temporary endpoint to staging.
- [x] 5.3 Invoke the endpoint with `curl`.
- [x] 5.4 Validate affected DynamoDB records have numeric `TTL` values.
- [x] 5.5 Remove the temporary endpoint and route after validation.
- [x] 5.6 Re-run verification after removal.
