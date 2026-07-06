## Context

The dashboard already has fast server-rendered pages, server actions, route-level loading states, toasts, and a polling provider. The remaining UX issue is not raw latency; it is interaction continuity. Several flows respond quickly but feel abrupt because state changes appear after a redirect or refresh, feedback can appear in more than one place, and some interactive surfaces do not preserve predictable focus or tap behavior.

Existing specs are intentionally conservative about routing. `dashboard-router-convention` requires `<Link>` for navigation and restricts client router APIs to polling-driven revalidation. `dashboard-ui-action-state-results` allows converted forms to use typed action state but explicitly preserves navigation-first redirects unless a task changes the UX. This design keeps those contracts intact.

## Goals / Non-Goals

**Goals:**
- Improve the dashboard's perceived smoothness without a broad rewrite of routing or server actions.
- Establish one feedback surface per mutation event so operators do not see duplicate toast and inline messages for the same action.
- Prioritize same-page mutation flows for inline pending/completion feedback because those are the flows where redirects feel most abrupt.
- Keep polling refreshes as background work so they do not compete with operator-initiated interactions.
- Remove invalid nested interactive controls that can make mobile and keyboard interactions feel unpredictable.
- Verify existing loading and focus requirements in the flows touched by this change.

**Non-Goals:**
- Do not replace navigation-first server-action redirects with client-side `router.push()`.
- Do not remove all `redirect()` calls from dashboard actions.
- Do not remove all `revalidatePath()` calls; server-rendered data still needs cache invalidation.
- Do not redesign the visual system, navigation shell, or backend API contracts.
- Do not treat the external `UX.md` review as authoritative implementation guidance.

## Decisions

### Preserve the router convention

Decision: Smoothness work will not introduce new `useRouter`, `router.push`, or `router.replace` usage outside the polling provider.

Rationale: The current specs intentionally keep dashboard navigation declarative and server-action based. Breaking that convention would make this change larger than necessary and would require modifying `dashboard-router-convention`.

Alternative considered: Convert create/update forms to `ActionState` and navigate with `router.push()` after success. This was rejected for this change because it conflicts with current specs and is not required to address the highest-value same-page smoothness issues.

### Improve same-page mutations first

Decision: Same-page mutations, such as monitor enable/disable and incident acknowledge/resolve, are the primary candidates for inline pending and completion feedback. Navigation-first create/delete flows may continue to redirect.

Rationale: Same-page mutations are where a redirect back to the same URL feels like an unnecessary cut. Create and delete flows often intentionally move the operator to a new resource or parent list, and existing specs already require those redirects in some cases.

Alternative considered: Convert all mutation flows uniformly. This was rejected because it increases scope and risks spec conflict without proving that every flow needs the same interaction model.

### Choose one feedback owner per event

Decision: Each mutation event will have a single feedback owner: either inline page feedback, action-state feedback, or toast feedback. The same query parameter or action result must not produce both an inline banner and a toast for the same event.

Rationale: Duplicate feedback makes an app feel bolted together even when each individual mechanism works. A feedback inventory should determine which surface owns each event before removing existing toast behavior.

Alternative considered: Delete the global toast watcher immediately. This was rejected because some routes may currently rely on toast-only feedback.

### Treat polling refresh as non-urgent background work

Decision: Polling-driven `router.refresh()` calls should run inside React transition scheduling and avoid redundant refresh bursts on visibility changes.

Rationale: Polling is background freshness work. It should not interrupt operator actions or produce avoidable state jumps.

Alternative considered: Remove polling refresh entirely. This was rejected because live operational dashboards need passive freshness.

### Fix interaction semantics before animation polish

Decision: Accessibility and semantic interaction fixes take priority over motion tuning, such as skeleton pulse timing.

Rationale: Invalid nested controls and unreliable focus restoration directly affect usability. Animation tuning is lower impact and should only follow after structural issues are resolved.

Alternative considered: Tune animations first for quick visual polish. This was rejected because it would not address the underlying abruptness operators feel during real actions.

## Risks / Trade-offs

- [Risk] Same-page action-state conversions could create a second mutation pattern that feels inconsistent with navigation-first forms. -> Mitigation: document which flows use inline state and keep create/delete navigation semantics unchanged.
- [Risk] Removing duplicate feedback could accidentally remove the only visible feedback for some routes. -> Mitigation: inventory every query-parameter and action-state feedback path before deleting or disabling any feedback surface.
- [Risk] Deferring broad redirect migration may leave some abrupt flows in place. -> Mitigation: prioritize the highest-frequency same-page actions first, then use operator feedback to decide whether a later router-convention proposal is justified.
- [Risk] Polling transition changes may not visibly improve perceived smoothness if flicker comes from another source. -> Mitigation: verify with browser QA before and after the polling change.
- [Risk] Adding loading placeholders without evidence can create more skeleton noise. -> Mitigation: add or adjust loading UI only where current route transitions are verified to be blank, mismatched, or layout unstable.

## Migration Plan

1. Inventory current feedback paths and identify duplicate toast/banner cases.
2. Apply polling refresh scheduling changes and verify no stale-data regression.
3. Fix invalid nested interactive mobile monitor cards.
4. Convert selected same-page mutation flows to inline pending/completion feedback where doing so does not require client router navigation.
5. Verify delete focus restoration and route loading behavior for touched flows.
6. Keep existing server-action redirects for navigation-first flows unless a follow-up change explicitly modifies the router convention.

## Open Questions

- Which feedback surface should be preferred for redirect destination events: inline banner, toast, or a route-specific rule?
- Which same-page mutation should be the reference implementation: monitor enable/disable, incident acknowledge/resolve, or manual run trigger?
- Does browser QA confirm visible polling refresh interruption, or is the smoothness issue mostly mutation-feedback related?
