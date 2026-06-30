## 1. Technology icon sizing

- [x] 1.1 Update `apps/dashboard/components/service-icon.tsx` to pass an explicit numeric `size` prop matching each frame (`sm=36`, `md=44`, `lg=56`).
- [x] 1.2 Verify the icon visually fills its frame on the home service health matrix, services list card grid, and service detail summary card.

## 2. Mono font for timestamps

- [x] 2.1 Sweep `apps/dashboard/app/page.tsx`, `apps/dashboard/app/(monitoring)/services/page.tsx`, `apps/dashboard/app/(monitoring)/services/[serviceId]/page.tsx`, `apps/dashboard/app/(monitoring)/services/[serviceId]/monitors/[monitorId]/page.tsx`, and `apps/dashboard/app/config/page.tsx` for any `formatDateTime(...)` not wrapped in `font-mono`.
- [x] 2.2 Add `font-mono` (with appropriate text size) to every non-table timestamp string identified in the sweep.
- [ ] 2.3 Run `make check-dashboard` and `make lint-dashboard` to confirm no type or lint regressions.

## 3. Channel type iconography

- [x] 3.1 Create `apps/dashboard/components/channel-type-icon.tsx` exporting `<ChannelTypeIcon type={...} />` with the lucide glyphs described in the design (`Send`, `Mail`, `MessageSquare`, `Webhook`, `Siren`).
- [x] 3.2 Update `apps/dashboard/app/(monitoring)/integrations/channels/page.tsx` to render `<ChannelTypeIcon type={channel.type} />` next to the type label in each row.
- [x] 3.3 Verify the icon renders alongside the readable type label (text not replaced).

## 4. Top bar removal

- [x] 4.1 Remove the sticky `<header>` block from `apps/dashboard/components/app-shell.tsx`.
- [x] 4.2 Update the main content wrapper `min-h-[calc(100vh-73px)]` to `min-h-screen` (or equivalent) in the same file.
- [x] 4.3 Verify each route still exposes a primary `<h1>` and the existing module-level "Create service" CTA remains reachable.

## 5. Service form lifecycle field removal

- [x] 5.1 Remove the read-only Lifecycle label block (and its helper text) from `apps/dashboard/components/service-form.tsx`.
- [x] 5.2 Verify lifecycle state remains visible on the service detail summary card, services list card, and home service health matrix.

## 6. Verification

- [x] 6.1 Run `make check-dashboard` and `make lint-dashboard`.
- [x] 6.2 Run `make build-dashboard`.
