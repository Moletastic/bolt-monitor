## Why

A focused UI/UX audit of `apps/dashboard` (see `apps/dashboard/`) surfaced defects that block the dashboard from feeling like a coherent operator console and from being safely extended as multi-region and richer policies land. Specifically:

- The single-region probe location constraint is presented as a normal multi-option picker in three different layers, so neither operators nor future implementers have an honest contract for what probe-location selection means today or what it must change into tomorrow.
- The escalation policy create action wires `businessHours` into `updateService(created.policyId, …)`, silently dropping service-scoped data because `policyId` is passed where `serviceId` is expected — and the failure is swallowed by `.catch(() => undefined)`.
- Destructive deletes (services, monitors, channels, policies) rely on `window.confirm`, breaking the visual language and accessibility story on the most consequential confirmation in the app.
- Several top-level pages (`/admin/scheduler`, `/locations`, the audit-trail placeholder) have no fallback when their backing API is unavailable, so a single 5xx takes down the entire shell instead of degrading gracefully.
- The dashboard chrome leaks internal scaffolding ("Bootstrap assumptions" panel), repeats three different product names ("Dashboard", "Operator Console", "Monitoring command center"), and ships one route in light-mode tokens against a forced-dark `<html>`.
- Several secondary polish items — empty-state announcements, table consistency, copy clarity, focus management after destructive actions — round out the change.

These are spec-driven work items per `AGENTS.md`; this change introduces the spec that captures them so they can be implemented incrementally without ad-hoc decisions.

## What Changes

- Replace the probe-location `<select>` in the monitor form with an honest presentation that depends on the runtime enabled-location count: a non-interactive chip with helper copy when only one location is enabled, and a real multi-select when the catalog has more than one enabled entry. Default selections must come from server data, not a hard-coded constant.
- Stop accepting `businessHours` from the dashboard escalation-policy create form and stop calling `updateService` from `createEscalationPolicyAction`. Escalation policy creation does not own service-scoped business hours; either remove the wiring or make it correct and observable.
- Replace `window.confirm` for destructive deletes with an in-app confirmation dialog that follows the dashboard design system, is keyboard accessible, and is testable.
- Add a top-level `app/error.tsx` and per-section fallback patterns so unhandled API failures degrade to an "unavailable" card inside the shared shell instead of an unstyled error boundary.
- Wrap top-level awaits in `app/admin/scheduler/page.tsx` and `app/locations/page.tsx` in try/catch with the same unavailable-card pattern used by the home page.
- Remove the audit-trail placeholder page (or replace it with a real empty state) and remove the "Bootstrap assumptions" panel from the sidebar chrome.
- Reskin `app/(monitoring)/policies/[policyId]/page.tsx` to use `AppShell` and the dark token set; replace the raw `bg-sky-100` token in `escalation-state-tab.tsx`.
- Unify table styling: replace the raw `<table>` in `integrations/channels/page.tsx` with the `<Table>` primitive used elsewhere.
- Polish copy: convert `down`/`warn`/`info` attention-queue chips to human labels, refine home empty-state copy, remove duplicate title/description in the DOWN service toast.
- Accessibility: add per-page `<h1>` and a skip-to-main-content link, give `EmptyState` an `aria-live` region, and add `aria-label` to status dots.
- Post-destructive-action focus must move to a sensible next target (next item, parent list, or "Create" CTA), not reset to `<body>`.

Out of scope (handled elsewhere or deferred):
- The `MONITOR_API_BASE_URL` / `NEXT_PUBLIC_MONITOR_API_BASE_URL` mismatch noted in the audit — excluded from this change per request.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dashboard-web-app`: Extend the dashboard web-app spec with requirements for honest probe-location presentation, destructive-action confirmation dialogs, graceful API unavailability fallbacks (top-level error boundary and per-section cards), removal of internal-scaffolding chrome, dark-token consistency across all policy surfaces, and accessibility/per-page-h1/skip-link refinements.
- `escalation-policy-crud`: Clarify that escalation policies do not own service-scoped `businessHours`; the dashboard escalation-policy create payload MUST NOT carry service-scoped business hours, and the dashboard server action MUST NOT call service update APIs from escalation-policy creation.
- `probe-location-catalog`: Clarify that when the dashboard renders a monitor location picker it MUST derive availability from the enabled subset of the canonical catalog, MUST NOT hard-code locations in dashboard actions, and MUST surface a single-enabled-location state honestly to operators until the catalog contains more than one enabled entry.

## Impact

- Dashboard code only: `apps/dashboard/app/`, `apps/dashboard/components/`, `apps/dashboard/lib/`, plus the addition of `apps/dashboard/app/error.tsx`.
- One new client dependency: `@radix-ui/react-alert-dialog` for the destructive confirmation dialog.
- No backend / Go / infra / shared-spec changes. The existing probe-location read API (`probe-location-read-api`) and notification-route channel-reference API already provide what the dashboard needs.
- Verification commands (all from repo root): `make lint-dashboard`, `make check-dashboard`, `make format-dashboard`, `make build-dashboard`.