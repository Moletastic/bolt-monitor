## Why

Dashboard v1 can create and edit monitors, but it has no API surface for discovering valid probe locations at runtime. The current backend embeds a default catalog in server code, which is enough for bootstrap tests but not a durable frontend contract. A probe-location read API should land as a focused follow-on change so the dashboard can render a real location picker without hardcoding option lists.

## What Changes

- Add an HTTP read endpoint that exposes the system-managed probe-location catalog for frontend consumption.
- Return only probe locations that are valid for operator selection in the current environment.
- Define a stable response shape with location identifier, display name, enabled state, and any minimal selection metadata needed by dashboard forms.
- Keep this change read-only and independent from future worker-routing, tenant override, or billing behavior.

## Capabilities

### New Capabilities
- `probe-location-read-api`: HTTP API for reading the selectable probe-location catalog.

### Modified Capabilities
- `probe-location-catalog`: Expose the canonical catalog through a product API read surface.
- `dashboard-web-app`: Replace hardcoded probe-location assumptions with runtime-discovered options for create and edit flows.

## Non-goals

- Do not add custom user-defined probe locations.
- Do not add tenant-specific location entitlement or RBAC rules in this change.
- Do not change monitor CRUD semantics beyond enabling the frontend to discover valid options.
- Do not implement scheduler routing or worker topology concerns beyond the metadata already implied by the catalog.

## Impact

- Affects API routing and backend response modeling.
- Reduces coupling between dashboard forms and backend hardcoded defaults.
- Creates a cleaner contract for future probe-location growth beyond the single built-in `iad` location.
- Keeps `dashboard-v1` focused by moving probe-location discovery into an explicit API change.
