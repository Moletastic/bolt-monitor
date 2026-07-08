## Why

Deep dashboard pages currently rely on sidebar state and page headings only, so operators lose parent context on create, detail, and nested monitor pages. Breadcrumbs give fast orientation and safe parent navigation without adding new backend behavior.

## What Changes

- Add breadcrumb navigation to dashboard pages that sit below a module landing page.
- Omit breadcrumbs on dashboard root and module landing pages where breadcrumb text would duplicate the page title or sidebar.
- Use human-readable resource names for dynamic detail crumbs when data is already loaded by the page.
- Use stable fallback labels when names are unavailable, such as `Service`, `Monitor`, `Incident`, `Route`, or `Channel`.
- Make parent crumbs links and current page crumb non-link text.
- Preserve App Router conventions by using `<Link>` for breadcrumb navigation.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `dashboard-web-app`: Add dashboard breadcrumb rules, rendering, accessibility, and route-specific labels.

## Impact

- Affected shell/component surface: `apps/dashboard/components/app-shell.tsx` or a new breadcrumb component used near `AppShell` content.
- Affected dashboard pages: create/detail/nested pages for services, monitors, notification routes, channels, incidents, settings subpages, audit trail, and probe locations.
- No backend API changes are expected.
