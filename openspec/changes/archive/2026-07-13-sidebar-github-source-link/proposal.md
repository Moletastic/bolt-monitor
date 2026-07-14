## Why

Operators and contributors currently have no direct path from the dashboard to the public source repository. Adding a clearly separated repository link to the sidebar improves source discoverability without mixing external resources into product navigation.

## What Changes

- Add an accessible `View source on GitHub` external link to the sidebar utility area.
- Point the link to the public `Moletastic/bolt-monitor` repository.
- Keep the link visually separate from module navigation and exclude it from active-route highlighting.
- Open the repository in a new browser tab with appropriate external-link semantics.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dashboard-sidebar-navigation`: Add an external source-repository affordance in the sidebar utility area while preserving module navigation behavior.

## Impact

- `apps/dashboard/components/app-shell.tsx` sidebar markup and icon imports.
- Dashboard component tests or guard tests covering sidebar link semantics.
- Existing dashboard sidebar navigation specification.
- No backend, API, AWS, environment, or package dependency changes.
