## Why

The Node orchestrator can preflight `TARGET_FILE`, while SST configuration ignores it and resolves a repository target from `TARGET`. One command can validate one target but evaluate or mutate another.

## What Changes

- Establish one canonical target-path resolution contract shared by the orchestrator and SST config.
- Reject conflicting target name/path inputs before any AWS mutation.
- Add tests proving preflight and SST evaluate the exact same target identity.
- FinOps: prevent accidental duplicate/wrong-stage deployments; no new AWS resources or recurring cost.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `stage-resource-lifecycle`: Every credentialed infrastructure workflow must resolve one identical target file.

## Impact

- `infra/scripts/ops.mjs`
- `infra/sst.config.ts`
- Target-resolution tests and operator documentation
