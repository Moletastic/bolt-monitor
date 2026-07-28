## Why

Ephemeral expiry currently prevents the supported removal command from loading its target. Expired stacks can therefore remain deployed and billable, contradicting verified cleanup and FinOps posture.

## What Changes

- Permit an expired but otherwise valid ephemeral target only for exact-stage removal and residual verification.
- Continue rejecting expired ephemeral targets for status, development, deployment, invitation, and key rotation.
- Add regression tests for expired cleanup and expired mutation refusal.
- FinOps: reuse existing SST removal and residual inventory; add no janitor, schedule, Lambda, queue, or scan.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `stage-resource-lifecycle`: Expired ephemeral targets must remain removable through the verified cleanup path.

## Impact

- `infra/deployment-target.ts`
- `infra/scripts/ops.mjs`
- Target validation and lifecycle tests
- Lifecycle operations documentation
