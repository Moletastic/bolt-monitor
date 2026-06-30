## Context

The dashboard has reached the point where most operator flows work but the surface area has accumulated inconsistencies that block it from feeling like one product and from being safely extended. The audit identified one critical config bug (excluded from this change) and a cluster of related polish items that all share a single theme: the dashboard currently relies on client-side constants, swallowed errors, and `window.confirm` to paper over gaps that should be expressed in the spec. Several pieces of internal scaffolding (the "Bootstrap assumptions" sidebar panel, the audit-trail placeholder) leak into the operator UI. One route ships in light-mode tokens against a forced-dark `<html>`, and the probe-location story is hard-coded in three layers with no shared contract.

This change is dashboard-only. The probe-location read API (`probe-location-read-api`) and the notification-route channel-reference API (`notification-route-channel-reference`) already provide everything the dashboard needs. No backend or shared Go modules change.

## Goals / Non-Goals

**Goals**

- One honest probe-location contract: the form, the action default, and the help copy all derive from the enabled subset of the canonical probe-location catalog.
- One source of truth for destructive confirmations: a Radix AlertDialog-based dialog that fits the dark operator aesthetic, is keyboard-accessible, and is testable.
- One graceful-degradation pattern: an `app/error.tsx` plus per-section fallback cards for `await`-of-API calls that currently throw the whole page.
- One visual language across all `(monitoring)` routes: `AppShell` wrapper + the standard status banners (`border-status-up/10 text-status-up` and the `/10`/`/30` opacity tokens) used in `services/[serviceId]/page.tsx`.
- Removal of internal scaffolding from operator surfaces (audit-trail placeholder, bootstrap-assumptions sidebar panel, light-mode policy edit page, sky-100 token in escalation tab).
- Accessibility baseline: per-page `<h1>`, skip-to-main link, `aria-live` empty states, `aria-label` on status dots.

**Non-Goals**

- Fixing the `MONITOR_API_BASE_URL` / `NEXT_PUBLIC_MONITOR_API_BASE_URL` mismatch (excluded per request; tracked separately).
- Adding real audit-trail content (route is removed or replaced with an empty state).
- Replacing the polling provider with `revalidate` (deferred to a future change; current polling stays as-is so this change stays reviewable).
- Multi-region probe-location enablement on the backend (this change makes the dashboard honest about the single-region preview; the API remains a single enabled location until the catalog is expanded).
- Refactoring the channels table into a generic sortable table primitive — the immediate goal is to use the existing `<Table>` shadcn primitive so the styling matches the rest of the app.

## Decisions

### Probe-location presentation is server-derived, not constant-driven

**Why:** The current `DEFAULT_PROBE_LOCATION = 'iad'` constant and the single hard-coded `<option>` mean that when the catalog expands, three places (form, action default, help copy) have to change in lockstep, with no compile-time forcing function. A server-side fetch of `listProbeLocations()` plus a conditional render closes that gap.

**How:** A new server-component helper, `apps/dashboard/lib/probe-locations.ts`, exposes a `getMonitorLocationField(locations)` function returning `{ kind: 'single-fixed', location } | { kind: 'multi', locations }`. The monitor form calls this and renders either a chip (`single-fixed`) or a multi-select (`multi`). The action default reads from the same source. `formatProbeLocations()` in `lib/utils.ts` already accepts arrays; it is reused unchanged.

**Alternatives considered:**

- *Keep the constant, add a runtime assertion.* This pushes the failure to first request, not first deploy. Rejected — the failure mode is silent (form still submits `iad`).
- *Render a disabled multi-select with one option.* Confuses operators and is the same dishonest UX we are replacing. Rejected.

### Escalation-policy create no longer wires service-scoped business hours

**Why:** The audit found that `createEscalationPolicyAction` calls `updateService(created.policyId, …)` with `policyId` where `serviceId` is expected, and the new-policy form does not submit a `serviceId`. The `.catch(() => undefined)` hides the failure, so this code path is dead code today and a future footgun. Policies do not own business hours; services do.

**How:** Remove the `updateService(...)` block in `apps/dashboard/lib/actions.ts:438-440`. Drop `businessHours` from the `EscalationPolicy` create payload type and from the form's hidden `businessHoursPathPayload` / `businessHoursPayload` writes in `components/escalation-policy-form.tsx`. If operators later need to attach a policy to a service's business hours, that is a separate spec change requiring the API to grow a real binding endpoint.

**Alternatives considered:**

- *Fix the call to pass the form's `serviceId`.* The form has no `serviceId` field today; adding one is a scope expansion. Rejected — out of scope.
- *Keep the call, remove the `.catch`, surface the 404.* Still leaves a dead/wrong code path. Rejected.

### Destructive confirmations use a Radix AlertDialog

**Why:** `window.confirm` is browser-styled, blocking, not focus-trapped, and skipped by some screen readers. The audit confirms this is the only confirmation for permanent deletion of services, monitors, channels, and policies — the most consequential action in the app.

**How:** Add `@radix-ui/react-alert-dialog` (it follows the same patterns already used by `react-checkbox` and `react-toast` in the dashboard). Build a small `<ConfirmDialog>` component in `components/ui/confirm-dialog.tsx` that takes `trigger`, `title`, `description`, `confirmLabel`, `cancelLabel`, and an `onConfirm`. `DeleteResourceForm` is refactored to render its destructive submit inside the dialog's `ConfirmDialog`. The existing `confirmMessage` strings become the `description`. Focus moves to the cancel button by default (destructive actions should require an extra deliberate click). After deletion, focus moves to the next list item or the parent "Create" CTA, not `<body>`.

**Alternatives considered:**

- *Use the existing `<Toast>` primitive to ask for confirmation.* Toasts auto-dismiss; wrong primitive for irreversible actions. Rejected.
- *Build a non-Radix dialog from scratch.* Rejected — Radix is already a dependency, and a11y for modals is hard to get right.

### Top-level error boundary + per-section fallback

**Why:** When `MONITOR_API_BASE_URL` is missing or the API is down, pages like `/admin/scheduler`, `/locations`, and the audit-trail placeholder currently throw the entire route. The user sees Next's default error boundary, which leaks implementation details and breaks the shared shell.

**How:** Add `apps/dashboard/app/error.tsx` that renders inside `AppShell` and shows an "unavailable" card with the error message and a `reset()` retry link. For pages that already work but have unguarded awaits (`/admin/scheduler`, `/locations`, `/incidents/[id]`), wrap the `await apiRequest(...)` in try/catch and render an `<UnavailableCard>` component. Extract `<UnavailableCard>` to `components/unavailable-card.tsx` so the pattern is reusable. The home page's existing per-section fallback (`app/page.tsx`) becomes the reference implementation.

**Alternatives considered:**

- *One global error boundary only.* Doesn't help pages where most data is fine but one section failed. Rejected — we already have the per-section pattern in the home page; reuse it.
- *Throw a typed `UnavailableError` and catch in a wrapper.* Same outcome, more code. Rejected.

### Visual consistency: AppShell + dark tokens across all policy surfaces

**Why:** `app/(monitoring)/policies/[policyId]/page.tsx` is the only `(monitoring)` route that ships without `AppShell` and uses raw Tailwind light tokens (`rose-50`, `sky-700`) that don't exist in the dark HSL palette. `escalation-state-tab.tsx:97` has a stray `bg-sky-100 text-sky-800`. These read as broken pages.

**How:** Wrap the policy edit page in `<AppShell currentPath="/policies">` and replace the inline status banners with the standard `border-status-up/10 text-status-up` / `border-status-down/10 text-status-down` banners used by `services/[serviceId]/page.tsx:152-160`. Replace `bg-sky-100 text-sky-800` with `bg-primary/10 text-primary` (the existing attention-queue tone). This is a mechanical reskin and does not change behavior.

### Chrome cleanup: remove internal scaffolding from operator surfaces

**Why:** The "Bootstrap assumptions" sidebar panel exposes spec language like "Single tenant context" and "Probe location picker uses current built-in catalog assumption" to operators. The audit-trail route is reachable but its content is developer notes. Three product names in one screen ("Monitoring command center", "Operator Console", "Dashboard") make the product feel undecided.

**How:**

- Remove the "Bootstrap assumptions" panel from `app-shell.tsx`. The information it carries belongs in the developer-only `/config` page or in a help modal, not in the persistent sidebar.
- Replace `app/audit-trail/page.tsx` with a real empty state pointing operators to `/incidents` (where audit information is surfaced today) and remove the developer-note paragraph.
- Pick one product name. Recommend: keep "Operator Console" in the sidebar header (already there) and update `app/page.tsx:411` from "Monitoring command center" to "Operator overview" to match.

**Alternatives considered:**

- *Move "Bootstrap assumptions" into `/config`.* Possible, but `/config` is being repurposed as the settings overview in the prior change (`dashboard-operator-ux-polish`). Keeping it out of the operator chrome entirely is simpler. Rejected.

### Accessibility baseline

**Why:** The audit found one `<h1>` (in the sidebar) that becomes the page H1 for every page, no skip-to-main link, `EmptyState` not announced, status dots relying on color + a single shape, and tabs implemented as `<a>` rather than `role="tab"`.

**How:**

- Add per-page `<h1>` inside each page's `<main>`, and demote the sidebar "Operator Console" to a `<div>` brand mark (or keep it as `<h1>` on the home page only).
- Add a skip-to-main-content link as the first focusable element of `app/layout.tsx`.
- Give `<EmptyState>` an `aria-live="polite"` region.
- Add `aria-label` to `StatusChip`'s dot (the text label is already uppercase and screen-reader-friendly, so this is belt-and-braces).
- Replace the `Tabs` primitive's `<Link>`-based implementation with Radix Tabs. Radix is already in the dependency tree.

**Alternatives considered:**

- *Only fix the missing skip link.* This is the WCAG 2.4.1 minimum; we want to fix the rest in one pass while we are touching these files. Rejected — partial fix.

### Table styling unification

**Why:** The integrations/channels table uses a raw `<table>` whose row dividers, padding, and typography don't match the rest of the app. Other tables (`incidents`, `runs`) use `<Table>` from `components/ui/table.tsx`. The channels table is the only outlier.

**How:** Replace the raw `<table>` in `app/(monitoring)/integrations/channels/page.tsx:37-65` with `<Table>`/`<TableHeader>`/`<TableBody>`/`<TableRow>`/`<TableHead>`/`<TableCell>`. Mobile card fallback is out of scope for this change.

## Risks / Trade-offs

- **[Radix AlertDialog dependency]** → Mitigation: `@radix-ui/react-alert-dialog` is small (~10 KB gz) and aligns with the existing Radix usage pattern. Pin the version in `package.json` alongside the other `@radix-ui/*` packages.
- **[Removing the `updateService` block may surprise any caller that relied on the dead code path]** → Mitigation: the audit confirms the call is dead (the form does not submit `serviceId`). Add a `console.warn` in dev if `businessHours` is present in the form payload, so any future regression is visible immediately.
- **[Honest probe-location rendering may look like a regression to operators used to the multi-option `<select>`]** → Mitigation: the helper copy explicitly says "single-region preview" and links to `/locations`, so the change is intentional and discoverable.
- **[Adding `app/error.tsx` may mask API bugs by showing a generic unavailable message]** → Mitigation: the message includes the underlying error string in a collapsible details block so support can still see it.
- **[Tabs migration to Radix]** → Mitigation: Radix Tabs require `<Tabs.List>` / `<Tabs.Trigger>` / `<Tabs.Content>` and the dashboard has only a handful of tab sites (monitor detail, incident detail). Migration is mechanical.
- **[Removing the "Bootstrap assumptions" panel may break a checklist of internal assumptions the team uses]** → Mitigation: move the contents into the repo's `AGENTS.md` gotchas section (which already lists them) or into a developer-only section of `/config`.

## Migration Plan

This is a dashboard-only change with no API or infra impact. Deploy order:

1. Merge the change into `production`.
2. Local validation runs `make lint-dashboard`, `make check-dashboard`, `make format-dashboard`, and `make build-dashboard`.
3. `make deploy-infra` re-deploys the dashboard stage.
4. Smoke test `/services`, `/services/<id>/monitors/<id>`, `/policies/<id>`, `/integrations/channels`, `/incidents`, `/locations`, `/admin/scheduler`, and the home page on staging.

Rollback: revert the merge commit. No data migration is involved.

## Open Questions

- Should the dashboard still expose a `currentPath` prop to `AppShell`, or should `AppShell` derive it from `usePathname()`? Both are valid; deriving it removes a drift risk but adds a client boundary to the shell. Recommendation: keep the prop for now, revisit when (if) the shell becomes more interactive.
- Should `service-list-status-toast.tsx` de-duplicate toasts (currently each service triggers its own toast on every poll until dismissed)? Out of scope here; mention as a follow-up.
- Should the per-page `<h1>` strategy make the sidebar header a `<div>` (so the home page is the only one with an `<h1>`), or should every page add an `<h1>` and the sidebar become a brand `<span>`? Recommend the latter — per-page `<h1>` matches user mental model and screen-reader behavior.
