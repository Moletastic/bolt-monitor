## Why

Dashboard surfaces feel too soft because cards, panels, buttons, and controls use relatively rounded corners across dense operational UI. Reducing general border radius will make the console feel sharper, more technical, and easier to scan.

## What Changes

- Reduce dashboard border-radius tokens globally.
- Apply smaller radii to cards, panels, inputs, selects, buttons, menus, dialogs, and inline feedback banners.
- Preserve intentional pill/circle shapes for status chips, traffic lights, avatars/icons, progress bars, and floating action buttons.
- Avoid one-off radius overrides unless a component needs a specific semantic shape.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `dashboard-web-app`: Tune dashboard radius scale and component corner treatment.

## Impact

- Affected dashboard styling: `apps/dashboard/tailwind.config.ts`, shared UI components, and page-level `rounded-*` usage.
- No backend, API, or data behavior changes.
