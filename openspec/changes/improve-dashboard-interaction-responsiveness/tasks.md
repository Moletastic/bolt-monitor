## 1. Characterize Interaction Boundaries

- [ ] 1.1 Inventory every dashboard server-action caller with its current route, successful destination, feedback owner, and whether the operation is reversible.
- [ ] 1.2 Record the exact server-rendered list, detail, aggregate, and history consumers affected by each mutation as the revalidation dependency map.
- [ ] 1.3 Add characterization coverage for current same-page forms, navigation-first redirects, polling refresh bounds, search request ordering, and history row retention before refactoring.
- [ ] 1.4 Replace broad source-string-only interaction assertions with behavior tests where a component or action boundary can be exercised, retaining source guards only for cross-codebase policy rules.

## 2. Action-State Interaction Primitive

- [ ] 2.1 Extend the shared same-page action form pattern to expose action-specific pending labels and disable every submit path for the same logical mutation while leaving unrelated controls enabled.
- [ ] 2.2 Standardize local success as a polite status announcement and local failure or rollback as an alert using typed action error messages.
- [ ] 2.3 Preserve initiating-control focus after non-navigation success and failure, and move focus to the first actionable validation target when structured field details identify one.
- [ ] 2.4 Add component tests for immediate pending feedback, rapid duplicate activation, pending reset, typed success/error rendering, and focus continuity.

## 3. Same-Page Mutation Conversion

- [ ] 3.1 Convert monitor maintenance toggles and manual-run triggers from same-route redirects to typed `ActionState` without query-string feedback.
- [ ] 3.2 Verify monitor enable/disable callers share one typed state action and remove any unused redirect-only toggle variant.
- [ ] 3.3 Verify incident acknowledge/resolve and scheduler configuration use the standardized pending, announcement, duplicate-prevention, and focus behavior.
- [ ] 3.4 Convert notification-channel and escalation-policy updates rendered on their current detail routes to typed `ActionState` while preserving submitted form values on failure.
- [ ] 3.5 Convert service archival on the current service detail route to typed `ActionState` while preserving its confirmation and server-confirmed completion.
- [ ] 3.6 Preserve server redirects for create flows, deletes, and service/monitor edits submitted from dedicated edit routes, with destination focus and exactly one result feedback surface.
- [ ] 3.7 Remove obsolete same-route query feedback keys, ownership entries, and redirect helpers after all callers use action state.
- [ ] 3.8 Add success, typed failure, duplicate-submission, announcement, and no-same-route-redirect coverage for every converted domain group.

## 4. Safe Optimistic Monitor State

- [ ] 4.1 Add optimistic presentation for monitor enable/disable only, capturing the prior enabled value before submitting the server action.
- [ ] 4.2 Reconcile successful optimistic state with server-rendered monitor data without remounting unrelated page content or interactive islands.
- [ ] 4.3 Roll back the optimistic enabled value before announcing a typed server failure and re-enable the control for a valid retry.
- [ ] 4.4 Add behavior tests for immediate optimistic presentation, blocked repeated toggles, successful reconciliation, failure rollback, and polling interaction during an in-flight toggle.
- [ ] 4.5 Add a guard that keeps destructive, irreversible, and side-effect-heavy operations outside the optimistic allowlist.

## 5. Narrow Server Revalidation

- [ ] 5.1 Apply the reviewed dependency map to each converted same-page action and remove unrelated `revalidatePath` calls.
- [ ] 5.2 Narrow navigation-first create, update, archive, and delete invalidation to proven destination and aggregate consumers without invalidating unreachable detail paths by default.
- [ ] 5.3 Ensure revalidation occurs only after successful mutations and that actions with no changed resource data, such as a channel test send, perform no route invalidation.
- [ ] 5.4 Add action-level tests asserting the exact successful revalidation set and zero revalidation for typed failures across each mutation category.

## 6. Focused Client Islands

- [ ] 6.1 Route global search requests through a typed server-owned interface so the client island receives results and safe errors without direct access to session credentials or authenticated backend details.
- [ ] 6.2 Keep global-search input responsive with deferred/debounced query work, suppress stale responses, bound requests per settled query, and preserve keyboard listbox behavior and `Link` result navigation.
- [ ] 6.3 Coalesce polling refresh work so interval and visibility events cannot overlap or issue redundant refreshes, retaining `startTransition` and the polling provider as the sole router exception.
- [ ] 6.4 Refine monitor-history pagination to block duplicate tab/page requests, preserve loaded rows during append and retry, announce pending/errors, and treat cursors as opaque server-action inputs.
- [ ] 6.5 Add client-island tests for stale search suppression, request bounds, polling overlap prevention, history append order, retained rows on failure, and retry behavior.

## 7. Navigation And Loading Continuity

- [ ] 7.1 Audit internal dashboard navigation and replace any non-Link internal navigation with `next/link` while preserving server-action redirects for destination-changing forms.
- [ ] 7.2 Keep default Link prefetch on likely destinations and document plus test any explicit prefetch opt-out supported by measured route cost or volatility.
- [ ] 7.3 Audit changed routes so `AppShell`, navigation, breadcrumbs, and resolved siblings remain outside local loading fallbacks.
- [ ] 7.4 Add or adjust Suspense boundaries only for independently resolving server sections, using shape-matched fallbacks and stable keys that do not remount search/history/action islands.
- [ ] 7.5 Add behavior and guard coverage for soft internal navigation, polling-only router usage, shared-shell persistence, resolved-sibling persistence, and reduced-motion-compatible fallbacks.

## 8. Verification

- [ ] 8.1 Run `make format-dashboard` and confirm formatting changes remain within the implementation scope.
- [ ] 8.2 Run `make lint-dashboard`.
- [ ] 8.3 Run `make check-dashboard`.
- [ ] 8.4 Run `make test-dashboard` and confirm deterministic request/revalidation bounds pass without wall-clock timing thresholds.
- [ ] 8.5 Run `make build-dashboard` to verify Server Component, server action, Suspense, and client boundary compatibility in production mode.
- [ ] 8.6 Browser-test converted mutations, search, polling, history pagination, Link navigation, and loading continuity at desktop and mobile widths.
- [ ] 8.7 Keyboard and screen-reader test pending, success, failure, rollback, validation focus, and post-redirect destination focus for changed flows.
- [ ] 8.8 Confirm no static migration, client canonical cache, direct browser refresh-token access, new data library, broad visual redesign, or imperative router call outside polling was introduced.
