## Why

Manual-run idempotency records calculate a 30-day expiry but omit DynamoDB TTL persistence. Keys never expire, retaining stale data and permanently blocking later valid use.

## What Changes

- Persist configured DynamoDB TTL as absolute Unix epoch seconds from manual idempotency record expiry.
- Add request-shape regression coverage for expiry and retention-window behavior.
- Document DynamoDB TTL's asynchronous cleanup semantics; request logic remains correct if physical deletion lags.
- FinOps: TTL deletes expired records without write capacity. No new table, GSI, Lambda, scan, or periodic cleanup job.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `manual-run-api`: Bounded manual idempotency retention must persist an expiry usable by DynamoDB TTL.

## Impact

- `services/monitor-api/manual_idempotency.go`
- `services/monitor-api/manual_idempotency_repository.go`
- Manual-run repository tests
- Existing AppTable TTL configuration; no new AWS resources or dependencies
