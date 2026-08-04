## Why

The README command and layout reference has drifted from the repository's validation surface and directory responsibilities. Correcting it gives contributors an accurate entry point without changing runtime behavior.

## What Changes

- Document `make test-infra` as the infrastructure and root-script test command.
- Clarify that `shared/` contains Go domain and backend-platform modules.
- Add `tools/` to the repository layout and accurately describe root `scripts/` as validation automation.

## Capabilities

### New Capabilities

- `repository-documentation`: Contributor-facing repository commands and top-level ownership remain accurate.

### Modified Capabilities

- None.

## Impact

- Affected file: `README.md`.
- No APIs, infrastructure, dependencies, or runtime behavior change.
