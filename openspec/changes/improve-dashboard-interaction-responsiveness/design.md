## Context

The dashboard is a Next 15 App Router application built around Server Components, server actions, server-side API calls, route-level loading UI, and a small polling provider. The existing smoothness work established typed `ActionState` for monitor enable/disable, incident transitions, scheduler configuration, channel test sends, and monitor-history pagination. It also scheduled polling refreshes as transition work and prohibited imperative router navigation outside polling.

The remaining inconsistency is architectural rather than visual. Some actions that return to the same rendered route still redirect and communicate through query parameters, including manual runs, maintenance toggles, channel/policy updates, and service archival. Successful actions often revalidate a list, detail, and nested detail path regardless of which surfaces actually derive changed data. Client search calls the API directly, while history already uses a server action, and loading boundaries are primarily route-wide rather than explicitly aligned with independently resolving sections.

This change must retain server-side truth and the opaque session boundary. Browser islands may own ephemeral input, pending, optimistic, or appended-page presentation, but they must not own canonical resource data, receive refresh credentials, or become a parallel client cache.

## Goals / Non-Goals

**Goals:**

- Make every mutation that stays on the current rendered route use a typed, locally rendered `ActionState` instead of a feedback-only redirect.
- Show immediate, action-specific pending state and prevent duplicate logical submissions without freezing unrelated page controls.
- Introduce optimistic presentation only through an explicit allowlist with rollback and server reconciliation.
- Reduce mutation-induced work by invalidating only data dependencies affected by a successful action.
- Preserve the shared shell and resolved sibling content during route streaming, local requests, and mutation reconciliation.
- Keep search, polling, and history pagination as focused client islands over server-owned interfaces.
- Preserve declarative `Link` navigation, soft transitions, appropriate framework prefetch, and the polling-only router exception.
- Add behavior-based accessibility and responsiveness tests in addition to narrow source-policy guards.

**Non-Goals:**

- No static export or migration away from the Next.js server runtime.
- No client-owned canonical query cache, global state store, or new data-fetching library.
- No browser access to refresh tokens, private session material, or direct authenticated backend credentials.
- No conversion of create/delete or dedicated edit-route workflows away from server redirects when success changes destination.
- No optimistic treatment of destructive, irreversible, externally side-effecting, or otherwise unsafe operations.
- No broad dashboard visual redesign, API redesign, or polling removal.
- No new imperative `useRouter`, `router.push`, `router.replace`, or ad hoc `router.refresh` usage outside the polling provider.

## Decisions

### Classify actions by destination semantics

Decision: Build and test an action inventory before conversion. The deciding question is whether successful completion leaves the operator on the same rendered route, not whether the action currently happens to call `redirect()`.

The expected categories are:

| Interaction | Result model | Navigation |
| --- | --- | --- |
| Monitor enable/disable, maintenance toggle, manual run | Typed same-page action state | None |
| Incident acknowledge/resolve, scheduler update, channel test | Typed same-page action state | None |
| Channel/policy update rendered on its detail route | Typed same-page action state | None |
| Service archive rendered on service detail | Typed same-page action state | None |
| Create service/monitor/policy/channel | Existing server redirect | New resource or collection destination |
| Service/monitor update submitted from a dedicated edit route | Existing server redirect | Detail destination |
| Service/monitor/policy/channel delete | Existing confirmation and server redirect | Parent/list destination |

The inventory is authoritative if current route placement differs from this expected matrix. Legacy redirect action variants that have no remaining navigation-first caller should be removed rather than retained as parallel behavior.

Rationale: Same-route redirects add a route cut and query-feedback coordination without changing destination. Creates, deletes, and dedicated edit forms benefit from explicit post-success navigation and already fit declarative server-action semantics.

Alternative considered: Convert all actions to state and navigate client-side after success. Rejected because it violates the router convention and needlessly moves orchestration into browser code.

### Use one reusable typed action-state contract, with local form ownership

Decision: Keep the existing serializable `ActionState<T>` error contract and server action `runServerAction` boundary. Same-page forms will consume the framework pending signal through `useActionState` or `useFormStatus`, render one local result surface, and disable every submit path for the same logical operation while pending.

Where a dialog or menu submits an external form through `requestSubmit()`, pending ownership must cover both the trigger and confirm control rather than only the eventual submit button. Action-specific labels such as `Running...`, `Saving...`, and `Enabling...` replace generic feedback where context is available.

Rationale: The existing contract already serializes typed API failures safely and integrates with server actions. Extending it is smaller and more consistent than introducing a mutation framework.

Alternative considered: Keep redirect query parameters and add loading spinners. Rejected because pending feedback alone does not remove the same-route navigation cut or duplicated feedback ownership.

### Allow optimistic presentation only for monitor enable/disable

Decision: The initial optimistic allowlist contains monitor enable/disable presentation only. Its local prior value is small and explicit, the operation has a direct inverse, and existing surfaces already expose the toggle state. The action still executes on the server, duplicate toggles are blocked, successful state is reconciled through narrow revalidation, and typed failure restores the captured previous value before announcing the error.

Maintenance mode, incident transitions, manual runs, scheduler changes, service archival, notification sends, edits, and deletes remain server-confirmed. The allowlist may only expand through a spec-backed change that identifies inverse behavior and side effects.

Rationale: Optimism is useful only where rollback is trustworthy and a brief speculative state cannot misrepresent an irreversible outcome. A one-operation allowlist proves the pattern without normalizing optimism for operational actions.

Alternative considered: Optimistically update every toggle-like control. Rejected because maintenance and scheduler controls have operational side effects, and incident transitions do not currently expose an undo operation.

### Revalidate from a dependency map after success only

Decision: Inventory each mutation's rendered consumers and encode the smallest required set of `revalidatePath` calls, or use existing cache tags if the codebase has a stable tag boundary by implementation time. Failed actions perform no revalidation. A same-page action may invalidate its current detail surface plus a parent summary only when both render changed data.

Examples:

- A manual run invalidates monitor status/history surfaces and any service summary that visibly derives run status; it does not invalidate the entire services tree by default.
- A channel test send returns action state but invalidates no resource route when the channel itself did not change.
- A channel edit invalidates the channel detail and channel collection; it does not invalidate policies unless their rendered data includes changed channel display fields.
- A delete invalidates the destination collection and any proven parent aggregate, not the now-unreachable deleted detail path merely for completeness.

Rationale: Revalidation is the bridge back to server truth, but broad invalidation turns a small mutation into unnecessary server fetch and render work.

Alternative considered: Keep the current defensive list/detail/nested-detail invalidation everywhere. Rejected because the behavioral specs now require bounded work and the affected consumers can be enumerated.

### Keep islands ephemeral and route all protected data through the server boundary

Decision: Preserve Server Components for route data and use client islands only where browser interaction requires them:

- Search owns input, deferred/debounced query state, result-list keyboard behavior, and stale-request suppression. It invokes a server-owned search interface and never receives session credentials.
- Polling owns interval/visibility scheduling and remains the sole client-router exception. It refreshes server-rendered truth as non-urgent work and coalesces overlapping refresh requests.
- Monitor history owns selected-tab state, pending state, and appended pages for the mounted view. Each page is loaded through the typed server action, cursors remain opaque, retries preserve loaded rows, and duplicate page requests are blocked.

Each island receives only the minimum serializable initial data. It does not mirror whole route resources or become a canonical cache. Search should use deferred input in addition to bounded request scheduling where that improves typing responsiveness; stale responses must not replace newer results.

Rationale: These interactions need local coordination, but their data access does not justify moving the route or session model to the client.

Alternative considered: Add a client query library for caching, retries, and request deduplication. Rejected because the current three islands are bounded and existing React/Next primitives cover the required behavior.

### Preserve soft navigation and default prefetch deliberately

Decision: Internal navigation remains `next/link`. Default prefetch stays enabled for likely destinations. Explicit `prefetch={false}` is permitted only where a measured route count, payload cost, or volatile destination demonstrates harm and the reason is captured by a focused test or nearby rationale. Forms with destination changes continue to redirect on the server.

Rationale: Link already provides soft navigation, accessibility semantics, and framework-managed prefetch without introducing imperative router state.

Alternative considered: Add click handlers and `router.push` to coordinate loading. Rejected because route loading boundaries provide this feedback declaratively and the approach violates the existing convention.

### Place loading boundaries around independent work, not the shell

Decision: Keep `AppShell`, navigation, breadcrumbs, and stable route context outside local Suspense fallbacks. Audit changed routes for sections that fetch independently and add Suspense only where the section can genuinely stream independently. Search and history islands show local pending UI while preserving results/rows. Narrow mutation reconciliation must not key or remount the entire page or island.

Rationale: A route-wide fallback can make fast local work look like a full reload. Conversely, excessive boundaries add fallback flashes and complexity, so boundaries require an independent fetch or verified continuity benefit.

Alternative considered: Wrap every card in Suspense. Rejected because boundaries without independent work do not improve responsiveness and can increase visual churn.

### Test user-observable responsiveness and retain policy guards

Decision: Add component/behavior tests for pending state timing, duplicate blocking, success/error announcements, optimistic rollback, preserved history rows, stale search result suppression, bounded polling, focus continuity, and Link navigation. Add focused action tests or dependency-map tests that assert exact successful revalidation sets and no failure revalidation. Retain source guards for the global prohibition on imperative router calls, but do not use source-string assertions as the sole proof of interaction behavior.

Performance coverage will use deterministic bounds such as one request per debounced search value, one history request per activation, no overlapping polling refresh, no unrelated revalidation call, and shell persistence. It will not add timing thresholds that are unstable in CI.

Rationale: Request/revalidation counts and mounted-content continuity are reliable regressions signals. Wall-clock microbenchmarks in unit tests are noisy and do not prove the intended architecture.

Alternative considered: Add only end-to-end latency thresholds. Rejected because infrastructure variance would make the tests flaky and obscure which work expanded.

## Risks / Trade-offs

- [Risk] Converting update forms on detail routes may require preserving draft values after typed validation failure. -> Mitigation: keep form state local, return structured error details, and test that submitted values and focus survive failure.
- [Risk] Optimistic monitor state can briefly disagree with a concurrent server or polling update. -> Mitigation: serialize duplicate local mutations, roll back typed failures, and let the next successful server reconciliation replace optimistic presentation.
- [Risk] Narrow revalidation can omit a real aggregate consumer and leave stale server-rendered data until polling. -> Mitigation: create the dependency inventory first, cover exact consumers in tests, and use polling only as a freshness backstop rather than correctness logic.
- [Risk] Suspense boundary changes can remount client islands and lose ephemeral search/history state. -> Mitigation: keep island identity and stable keys outside the smallest revalidating boundary and test row/input retention.
- [Risk] A server-owned search interface can add an extra server hop compared with a direct browser API call. -> Mitigation: keep payloads bounded, debounce/defer requests, and prefer session opacity and consistent typed errors over exposing future authentication details.
- [Risk] Disabling all equivalent submit paths requires coordination across menus and confirmation dialogs. -> Mitigation: assign one pending owner per logical action and test rapid repeated activation through every visible trigger.

## Migration Plan

1. Add characterization tests and an action/destination/revalidation inventory before changing behavior.
2. Extend shared action-state form primitives with duplicate-submission, announcement, and focus behavior.
3. Convert same-page actions in small domain groups, remove unused redirect variants, and update feedback ownership as each flow moves.
4. Add optimistic monitor enable/disable with explicit prior-state rollback; leave all other operations server-confirmed.
5. Replace broad mutation invalidation with the reviewed dependency map and exact-call tests.
6. Harden search, polling, and history islands behind server-owned interfaces with stale/duplicate request protection.
7. Adjust shell and granular loading/Suspense boundaries only where route audits demonstrate independent work.
8. Run dashboard lint, type checks, unit/guard tests, production build, and desktop/mobile browser verification with keyboard and reduced-motion checks.

Rollback is incremental: each domain conversion can return to its previous server redirect while retaining the action-state primitives. Optimistic monitor presentation can be removed without changing the server action contract. No persisted data or backend migration is involved.

## Open Questions

None. The action inventory may refine individual callers, but the destination-based classification and initial optimistic allowlist are fixed by this design.
