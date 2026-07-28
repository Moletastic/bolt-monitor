## Why

Deploy postflight accepts absent or unscoped SST outputs and can skip API health verification. Operators can receive a successful deploy result without proving target-specific application outputs or public health.

## What Changes

- Require target-scoped deploy outputs and required public/resource identifiers after deploy.
- Fail postflight when outputs are absent, stage-mismatched, malformed, or missing required fields.
- Require public health verification for every deployed target.
- FinOps: avoids failed rollout churn and accidental validation of another stage. No new health checker, metric, resource, or recurring request stream.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `stage-resource-lifecycle`: Ordinary deploy completion must verify target-specific outputs and public health.

## Impact

- `infra/scripts/ops.mjs`
- SST output parsing and operations tests
- Deployment lifecycle documentation
