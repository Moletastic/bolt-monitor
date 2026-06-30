## Why

The operator dashboard accumulates visual rough edges as new modules come online. Technology icons render smaller than their frames, dates mix serif-adjacent typography across surfaces, the notification-channel type is text-only, and the service form carries a read-only lifecycle field that duplicates information already shown elsewhere. A focused visual polish pass removes the rough edges without introducing new product behavior.

## What Changes

- Enlarge technology icons inside `ServiceIcon` so they consistently fill their frame across `sm`, `md`, and `lg` sizes, including on the services list card grid, the home service health matrix, and the service detail summary.
- Apply the mono font token to every rendered timestamp on operator surfaces (service cards, summary cards, status banners, settings cards, and the remaining non-table date strings) so timestamps read as data rather than prose.
- Replace the text-only channel type cell on the channels list with a `ChannelTypeIcon` rendered next to the type label, using a `lucide-react` glyph per channel type (telegram, email, sms, webhook, pagerduty). The icon is informational, not interactive.
- Remove the sticky top bar from the shared `AppShell`. The sidebar already carries primary navigation and the existing "Create service" CTA appears in module landing pages, so the top bar adds chrome without adding navigation. The "Operator Workspace" copy in the top bar is internal scaffolding and disappears with the bar.
- Remove the read-only Lifecycle field from the service create/edit form. Lifecycle state remains visible on the service detail summary and the home service health matrix; the form only collects inputs the operator can change.

## Capabilities

### New Capabilities
- `dashboard-channel-type-icon`: visual identity for notification channel types on dashboard surfaces.

### Modified Capabilities
- `dashboard-web-app`: tighten the technology icon footprint requirement, extend the human-readable and design-language requirements to cover mono timestamps, channel-type iconography, top-bar removal, and removal of the lifecycle form field.

## Impact

- `apps/dashboard/components/service-icon.tsx`: pass explicit size to `devicons-react` icons.
- `apps/dashboard/components/app-shell.tsx`: drop the sticky top header.
- `apps/dashboard/components/service-form.tsx`: remove the Lifecycle label block.
- `apps/dashboard/app/(monitoring)/integrations/channels/page.tsx`: render `ChannelTypeIcon` in the type cell.
- `apps/dashboard/app/page.tsx`, `apps/dashboard/app/(monitoring)/services/page.tsx`, `apps/dashboard/app/(monitoring)/services/[serviceId]/page.tsx`, `apps/dashboard/app/(monitoring)/services/[serviceId]/monitors/[monitorId]/page.tsx`, `apps/dashboard/app/config/page.tsx`, and other surfaces that render dates: ensure mono font on `formatDateTime` output where missing.
- `apps/dashboard/components/` (new): `channel-type-icon.tsx` mapping channel type to lucide glyph.
