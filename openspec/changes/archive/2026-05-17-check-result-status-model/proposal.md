## Why

Execution output is only useful if persisted in a shape that supports latest status, run history, and future incident logic. Result and status model should land before alerting or dashboard work so those features build on stable execution data.

## What Changes

- Define canonical result model for executed healthchecks.
- Define current status model derived from latest execution outcomes.
- Define persistence semantics for raw check runs and status snapshots in single-table DynamoDB.
- Define retention expectations for raw runs versus current status state.

## Capabilities

### New Capabilities
- `check-result-status-model`: Canonical storage and shared model for raw check results and latest monitor status.

### Modified Capabilities
- `dynamodb-single-table-storage`: Extend item-family usage with concrete result/status item semantics.

## Impact

- Affects execution pipeline output contract and DynamoDB writes.
- Affects future read APIs, incidents, and dashboards.
- Constrains what status fields are available before alerting logic exists.
