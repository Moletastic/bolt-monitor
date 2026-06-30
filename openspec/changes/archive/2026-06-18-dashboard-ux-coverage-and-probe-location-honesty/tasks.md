## 1. Probe-location honesty

- [x] 1.1 Add `apps/dashboard/lib/probe-locations.ts` exposing `getMonitorLocationField(locations)` returning `{ kind: 'single-fixed', location }` or `{ kind: 'multi', locations }` based on the enabled subset.
- [x] 1.2 Replace the single `<option>IAD · US East</option>` in `components/monitor-form.tsx` with conditional rendering driven by `getMonitorLocationField`: chip + helper copy for `single-fixed`, multi-select for `multi`.
- [x] 1.3 Remove the hard-coded `DEFAULT_PROBE_LOCATION` constant in `lib/actions.ts` and derive the submitted probe location from server-side catalog data in the create and update monitor actions.
- [x] 1.4 Update the helper text on `app/services/[serviceId]/monitors/new/page.tsx` to reflect the server-derived location selection rather than referring to a built-in catalog assumption.
- [x] 1.5 Verify the `formatProbeLocations()` helper in `lib/utils.ts` already handles arrays of locations; adjust if not.

## 2. Escalation policy action correctness

- [x] 2.1 Remove the `updateService(created.policyId, …)` block in `lib/actions.ts` `createEscalationPolicyAction` and stop accepting `businessHours` from the form payload.
- [x] 2.2 Drop `businessHours` writes from the hidden form fields in `components/escalation-policy-form.tsx` (`businessHoursPathPayload`, `businessHoursPayload`, `offHoursPathPayload`).
- [x] 2.3 Add a dev-only `console.warn` when the submitted form payload still contains `businessHours`, so any regression is visible during development.
- [x] 2.4 Verify the channel `<Select>` in `escalation-policy-form.tsx` shows an empty `Pick a channel` option as its first entry so client-side validation matches the server rule.

## 3. Destructive-action confirmation dialog

- [x] 3.1 Add `@radix-ui/react-alert-dialog` to `apps/dashboard/package.json` and run `npm install`.
- [x] 3.2 Create `components/ui/confirm-dialog.tsx` exposing `<ConfirmDialog>` with `trigger`, `title`, `description`, `confirmLabel`, `cancelLabel`, and `onConfirm`. Cancel control receives focus on open.
- [x] 3.3 Refactor `components/delete-resource-form.tsx` to render its destructive submit inside `<ConfirmDialog>` using the existing `confirmMessage` strings as `description`.
- [x] 3.4 Move keyboard focus to the next list item, parent list, or Create CTA after a successful destructive delete — not back to `<body>`.

## 4. Graceful API unavailability

- [x] 4.1 Add `apps/dashboard/app/error.tsx` rendering an unavailable card inside `AppShell` with a `reset()` retry control and a collapsible details block containing the underlying error message.
- [x] 4.2 Extract `components/unavailable-card.tsx` from the existing per-section fallback pattern used by `app/page.tsx`.
- [x] 4.3 Wrap the unguarded `await getSchedulerConfig()` in `app/admin/scheduler/page.tsx` in try/catch and render `<UnavailableCard>` on failure.
- [x] 4.4 Wrap the unguarded `await listProbeLocations()` in `app/locations/page.tsx` in try/catch and render `<UnavailableCard>` on failure.
- [x] 4.5 Apply the per-section fallback pattern to `app/(monitoring)/services/[serviceId]/monitors/[monitorId]/page.tsx` so that a single 5xx in one of the six parallel fetches does not blank the entire page.

## 5. Visual consistency for policy surfaces

- [x] 5.1 Wrap `app/(monitoring)/policies/[policyId]/page.tsx` in `<AppShell currentPath="/policies">` and replace the inline light-mode status banners with the standard `border-status-up/10 text-status-up` and `border-status-down/10 text-status-down` banners used by `services/[serviceId]/page.tsx`.
- [x] 5.2 Replace `bg-sky-100 text-sky-800` in `app/(monitoring)/incidents/[id]/escalation-state-tab.tsx:97` with `bg-primary/10 text-primary`.

## 6. Chrome cleanup

- [x] 6.1 Remove the "Bootstrap assumptions" panel from `components/app-shell.tsx` and move any retained content into `AGENTS.md` gotchas.
- [x] 6.2 Replace `app/audit-trail/page.tsx` with a real empty state that points operators to `/incidents` for current audit information, or remove the route and update sidebar navigation accordingly.
- [x] 6.3 Unify product naming on the home page: change the section header at `app/page.tsx:411` to use "Operator overview" so the home page heading matches the sidebar brand.

## 7. Accessibility baseline

- [x] 7.1 Add a per-page `<h1>` inside `<main>` on each dashboard page; demote the sidebar "Operator Console" header to a brand `<span>` (or keep as `<h1>` only on the home page).
- [x] 7.2 Add a skip-to-main-content link as the first focusable element of `app/layout.tsx`.
- [x] 7.3 Give `EmptyState` an `aria-live="polite"` region and ensure each instance explains the next operator action.
- [x] 7.4 Add `aria-label` to `StatusChip`'s status dot.
- [x] 7.5 Migrate the existing `<Link>`-based `Tabs` primitive in `components/ui/tabs.tsx` to Radix Tabs with proper `role="tablist"` semantics.

## 8. Table styling unification

- [x] 8.1 Replace the raw `<table>` in `app/(monitoring)/integrations/channels/page.tsx:37-65` with `<Table>` / `<TableHeader>` / `<TableBody>` / `<TableRow>` / `<TableHead>` / `<TableCell>` from `components/ui/table.tsx`, matching the incidents table pattern.

## 9. Copy and microcopy

- [x] 9.1 Convert `down` / `warn` / `info` attention-queue chip labels in `app/page.tsx:213-215` to "Action needed" / "At risk" / "Heads-up".
- [x] 9.2 Change the all-clear empty state in `app/page.tsx:197-198` to "All clear across loaded modules.".
- [x] 9.3 Remove the redundant `description: service.name` in the DOWN toast at `components/service-list-status-toast.tsx:51-55`.

## 10. Verification

- [x] 10.1 Run `make lint-dashboard`.
- [x] 10.2 Run `make check-dashboard`.
- [x] 10.3 Run `make format-dashboard` and confirm the format check passes.
- [x] 10.4 Run `make build-dashboard` and confirm the production build succeeds.
- [ ] 10.5 Smoke test `/services`, `/services/<id>/monitors/<id>`, `/policies/<id>`, `/integrations/channels`, `/incidents`, `/locations`, `/admin/scheduler`, and the home page on staging; confirm the AppShell wraps every page, the probe-location chip renders, the destructive dialog opens on delete, and the unavailable card renders when each API is offline.