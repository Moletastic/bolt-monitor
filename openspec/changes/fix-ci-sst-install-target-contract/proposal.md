## Why

Infrastructure CI runs `sst install` with only `TARGET=example`. SST now requires the canonical target path and a stage matching that target, so platform type generation fails before release gates run.

## What Changes

- Pass the committed example target path and its declared stage to the CI SST install command.
- Keep platform type generation credential-free and deterministic.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `ci-tooling-bootstrap`: CI platform type generation uses the same target-path and stage contract as SST configuration.

## Impact

- `.github/workflows/ci.yml`
- CI bootstrap specification and verification
