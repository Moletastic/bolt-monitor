## Why

The dashboard declares `devicons-react`, but its service icons now use the repository-owned `TechIcon` SVGs. Removing the unused package reduces the dependency graph without changing dashboard behavior.

## What Changes

- Remove `devicons-react` from the dashboard manifest and pnpm lockfile.
- Verify the dashboard source, tests, and build do not require it.

## Capabilities

### New Capabilities
- `dependency-hygiene`: Evidence-based removal of an unused dashboard dependency.

## Impact

- Affects `apps/dashboard/package.json` and `apps/dashboard/pnpm-lock.yaml` only.
- Does not change Go modules, used dependencies, install-script policy, application behavior, or API contracts.
