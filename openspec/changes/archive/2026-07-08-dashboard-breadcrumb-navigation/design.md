## Context

The dashboard uses `AppShell` for shared navigation. Sidebar nav identifies the active module, but detail and create pages often sit several levels below the module root. Some pages include ad hoc “back” links, while others rely on page copy only.

Breadcrumbs should add hierarchy without duplicating every page heading or replacing primary navigation. They should use the same App Router convention already documented for dashboard navigation: links for navigation, no imperative router calls.

## Goals / Non-Goals

**Goals:**

- Give operators parent context on detail, create, and nested pages.
- Provide consistent parent navigation for services, monitors, notification routes, channels, incidents, and settings subpages.
- Use resource names in breadcrumbs when pages already load those resources.
- Keep module landing pages clean by omitting redundant breadcrumbs.
- Preserve accessibility with semantic breadcrumb navigation.

**Non-Goals:**

- Add breadcrumbs to the root dashboard page.
- Add backend endpoints just for breadcrumb labels.
- Replace sidebar navigation.
- Add imperative route handling.
- Breadcrumb every transient tab or in-page state.

## Decisions

### Appearance Rules

Breadcrumbs should appear when the current page has meaningful parent context beyond the active sidebar module.

Show breadcrumbs on:

```txt
/services/new
/services/[serviceId]
/services/[serviceId]/monitors/new
/services/[serviceId]/monitors/[monitorId]
/policies/new
/policies/[policyId]
/integrations/channels/new
/integrations/channels/[channelId]
/incidents/[id]
/admin/scheduler
/locations
/audit-trail
```

Omit breadcrumbs on:

```txt
/
/services
/policies
/integrations/channels
/incidents
/config
```

Reason: module landing pages already act as top-level destinations and usually have their own page header. Breadcrumbs there would repeat sidebar labels.

### Route Labels

Use short, operator-facing labels:

| Route pattern | Breadcrumb |
|---|---|
| `/services/new` | `Services / Create service` |
| `/services/[serviceId]` | `Services / {service.name}` |
| `/services/[serviceId]/monitors/new` | `Services / {service.name} / Create monitor` |
| `/services/[serviceId]/monitors/[monitorId]` | `Services / {service.name} / {monitor.name}` |
| `/policies/new` | `Notification routes / Create route` |
| `/policies/[policyId]` | `Notification routes / {policy.name}` |
| `/integrations/channels/new` | `Channels / Create channel` |
| `/integrations/channels/[channelId]` | `Channels / {channel.name}` |
| `/incidents/[id]` | `Incidents / {incident.summary or incidentId}` |
| `/admin/scheduler` | `Settings / Scheduler` |
| `/locations` | `Settings / Probe locations` |
| `/audit-trail` | `Incidents / Audit trail` |

Dynamic crumbs should use loaded page data. If the page cannot load the name but still renders an unavailable state, use fallback labels like `Service`, `Monitor`, `Notification route`, `Channel`, or `Incident`.

### Rendering Model

Either pass `breadcrumbs` into `AppShell` or render a shared `<Breadcrumbs>` component just inside each page. Preferred path: add optional `breadcrumbs` prop to `AppShell` so placement stays consistent across pages.

```ts
type BreadcrumbItem = {
  label: string
  href?: string
}
```

All non-current crumbs have `href`; current crumb has no `href` and uses `aria-current="page"`.

### Placement

Breadcrumbs belong above page-specific content and below any future top bar/search area. On mobile, breadcrumbs can wrap or truncate middle/current labels, but parent links must remain tappable.

## Risks / Trade-offs

- **Duplicate page headings** -> Omit breadcrumbs on module landing pages and keep crumb labels compact.
- **Dynamic label loading complexity** -> Use data already fetched by pages; no extra API just for breadcrumbs.
- **Long names overflow mobile** -> Truncate long crumb labels with accessible full text where practical.
- **Ad hoc page back links conflict** -> Breadcrumbs can replace redundant “back to list” links when implementation touches those pages.
