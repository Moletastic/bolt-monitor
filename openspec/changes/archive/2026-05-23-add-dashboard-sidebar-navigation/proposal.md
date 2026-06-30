## Why

 Dashboard app already uses `AppShell`, but root route still doubles as monitor overview and navigation still reflects bootstrap-era monitor-first links. Operators need a stable sidebar and a clearer module hierarchy where `Dashboard` is the default console landing area and monitor operations live under `Services`.

## What Changes

- Add a dedicated dashboard sidebar navigation model with top-level modules: `Dashboard`, `Services`, `Integrations`, `Audit Trail`, and `Config`.
- Update dashboard shell so every dashboard page renders the new module-oriented sidebar with correct active state behavior.
- Move current monitor overview content from `apps/dashboard/app/page.tsx` into the `Services` module.
- Make root dashboard page a lightweight `Dashboard` landing surface with WIP/empty-state messaging for now.
- Add module landing pages for routes that do not yet have feature content so navigation remains functional instead of pointing at missing pages.

## Capabilities

### New Capabilities
- `dashboard-sidebar-navigation`: Module-oriented sidebar navigation and landing pages for dashboard application.

### Modified Capabilities
- `dashboard-web-app`: Dashboard navigation and page framing requirements expand to include module-oriented sidebar structure across dashboard surfaces.

## Impact

- Affects `apps/dashboard` layout, shell, route structure, root landing content, and current monitor overview placement.
- No backend API changes required.
- Requires new dashboard routes for module landing pages and updated navigation behavior for existing monitor pages.
