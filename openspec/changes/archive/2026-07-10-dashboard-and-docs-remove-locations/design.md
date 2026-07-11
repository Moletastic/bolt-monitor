## Context

The current dashboard intentionally makes one enabled location visible:

```text
New/Edit Monitor
  └─ Probe location chip: IAD / US East
       └─ hidden input submitted as probeLocation
            └─ server action sends probeLocations: [value]

Locations page
  └─ lists enabled probe locations

Monitor views
  └─ display configured probe locations or lastProbeLocationId
```

That was honest while the backend had a location catalog. Once the backend no longer exposes locations, the UI should become simpler rather than replacing `IAD` with another placeholder.

## Goals / Non-Goals

**Goals:**

- Remove all dashboard operator-visible location concepts.
- Remove catalog fetches from monitor forms and settings pages.
- Remove OpenAPI location schemas and examples.
- Keep monitor create/update workflows otherwise unchanged.

**Non-Goals:**

- Backend contract changes; handled by `backend-single-execution-environment`.
- Adding environment, worker, or deployment-zone status in the UI.
- Keeping a disabled/empty Locations page.

## Decisions

### Decision: Delete the Locations page instead of showing a static environment page

**Choice:** Remove the route and navigation surface for probe locations.

**Rationale:** A static replacement page would still imply operators need to reason about execution geography. If a future runtime-environment diagnostics page is needed, it should be proposed separately with a different purpose.

### Decision: Monitor forms submit no execution-location data

**Choice:** Create/update actions build monitor payloads from monitor configuration fields only.

**Rationale:** This aligns the UI with the simplified API and removes a failure mode where catalog fetches block monitor creation.

## Risks / Trade-offs

- **[Risk] Removing the Locations nav item changes smoke-test routes.** Mitigation: update dashboard tests and docs at the same time.
- **[Risk] Dashboard may compile against stale generated/manual API types.** Mitigation: remove types and references in one sweep, then run dashboard typecheck.
- **[Trade-off] Operators lose visibility into execution geography.** Accepted because geography is no longer a product control or supported diagnostic dimension.
