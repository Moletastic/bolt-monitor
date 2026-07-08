## 1. Shared Breadcrumb Components

- [x] 1.1 Add a reusable breadcrumb item type with `label` and optional `href`.
- [x] 1.2 Add breadcrumb rendering to `AppShell` or a shared dashboard breadcrumb component.
- [x] 1.3 Ensure parent crumbs render as links and current crumb renders with current-page semantics.
- [x] 1.4 Add responsive truncation/wrapping so long labels do not break narrow layouts.

## 2. Route Breadcrumbs

- [x] 2.1 Add breadcrumbs for service create, service detail, monitor create, and monitor detail pages.
- [x] 2.2 Add breadcrumbs for notification route create and detail pages.
- [x] 2.3 Add breadcrumbs for channel create and detail pages.
- [x] 2.4 Add breadcrumbs for incident detail, scheduler settings, probe locations, and audit trail pages.
- [x] 2.5 Omit breadcrumbs on root and module landing pages.

## 3. Dynamic Labels And Cleanup

- [x] 3.1 Use loaded service, monitor, policy, channel, and incident data for dynamic breadcrumb labels.
- [x] 3.2 Add stable fallback labels for unavailable dynamic resource names.
- [x] 3.3 Remove redundant ad hoc “back to list” links where breadcrumbs now provide the same parent navigation.

## 4. Verification

- [x] 4.1 Run `make lint-dashboard`.
- [x] 4.2 Run `make check-dashboard`.
- [x] 4.3 Run `make test-dashboard` if breadcrumb tests are added or changed.
